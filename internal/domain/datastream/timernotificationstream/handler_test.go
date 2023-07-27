package timernotificationstream_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Tap-Team/timerapi/internal/database/postgres/notificationstorage"
	"github.com/Tap-Team/timerapi/internal/database/postgres/timerstorage"
	"github.com/Tap-Team/timerapi/internal/database/redis/subscriberstorage"
	"github.com/Tap-Team/timerapi/internal/domain/datastream/timernotificationstream"
	"github.com/Tap-Team/timerapi/internal/domain/usecase/timerusecase"
	"github.com/Tap-Team/timerapi/internal/errorutils/timererror"
	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/internal/model/timermodel/timerfields"
	"github.com/Tap-Team/timerapi/internal/testdatamodule"
	"github.com/Tap-Team/timerapi/internal/timerservice"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/Tap-Team/timerapi/pkg/postgres"
	"github.com/Tap-Team/timerapi/pkg/rediscontainer"
	"github.com/Tap-Team/timerapi/proto/timerservicepb"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	path             = postgres.DEFAULT_MIGRATION_PATH
	tickerServiceUrl = "localhost:50001"
)

type TimerStorage interface {
	timernotificationstream.TimerStorage
	timerusecase.TimerStorage
}

type SubscriberCacheStorage interface {
	timernotificationstream.SubscriberCacheStorage
	timerusecase.SubscriberCacheStorage
}

var (
	timerStorage        TimerStorage
	subscriberStorage   timernotificationstream.SubscriberCacheStorage
	notificationStorage timernotificationstream.NotificationStorage

	timerService timerservice.TimerServiceClient
)

func TestMain(m *testing.M) {
	os.Setenv("TZ", "UTC")
	ctx := context.Background()
	p, term, err := postgres.NewContainer(ctx, path)
	if err != nil {
		log.Fatalf("failed to start postgres container, %s", err)
	}
	defer term(ctx)
	fmt.Println(p.Pool.Ping(ctx))
	rc, term, err := rediscontainer.New(ctx)
	if err != nil {
		log.Fatalf("failed to start redis container, %s", err)
	}
	defer term(ctx)
	conn, err := grpc.DialContext(ctx, tickerServiceUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial context failed, %s", err)
	}
	timerStorage = timerstorage.New(p)
	notificationStorage = notificationstorage.New(p)
	subscriberStorage = subscriberstorage.New(rc)
	timerService = timerservice.GrpcClient(timerservicepb.NewTimerServiceClient(conn))

	notificationStream := timernotificationstream.New(timerService, timerStorage, subscriberStorage, notificationStorage)
	go func() { notificationStream.Start(ctx) }()
	m.Run()
}

const listSize = 100

func testData(t *testing.T, ctx context.Context, opts ...testdatamodule.TimerOption) []*timermodel.Timer {
	timers := testdatamodule.RandomTimerList(listSize, opts...)
	for _, timer := range timers {
		switch timer.Type {
		case timerfields.COUNTDOWN:
			err := timerStorage.InsertCountdownTimer(ctx, timer.Creator, timer.CreateTimer())
			require.NoError(t, err, "failed insert countdown timer")
		case timerfields.DATE:
			err := timerStorage.InsertDateTimer(ctx, timer.Creator, timer.CreateTimer())
			require.NoError(t, err, "failed insert date timer")
		}
		err := timerService.Add(ctx, timer.ID, timer.EndTime.Unix())
		require.NoError(t, err, "failed to add timer in timer service")
	}
	return timers
}

func TestExpiredDateTimer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	duration := 3
	endTime := time.Now().Add(time.Second * time.Duration(duration))
	timers := testData(t, ctx, func(t *timermodel.Timer) {
		t.EndTime = amidtime.DateTime(endTime)
		t.Duration = int64(duration)
		t.Type = timerfields.DATE
	})

	time.Sleep(time.Second * time.Duration(duration) * 2)

	for _, timer := range timers {
		_, err := timerStorage.Timer(ctx, timer.ID)
		require.ErrorIs(t, err, timererror.ExceptionTimerNotFound())
	}

	for _, timer := range timers {
		timerStorage.DeleteTimer(ctx, timer.ID)
	}

}

func TestExpiredCountdownTimer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	duration := 3
	endTime := time.Now().Add(time.Second * time.Duration(duration))
	timers := testData(t, ctx, func(t *timermodel.Timer) {
		t.EndTime = amidtime.DateTime(endTime)
		t.Duration = int64(duration)
		t.Type = timerfields.COUNTDOWN
	})

	time.Sleep(time.Second * time.Duration(duration) * 2)

	for _, timer := range timers {
		tm, err := timerStorage.Timer(ctx, timer.ID)
		require.NoError(t, err, "failed to get countdown timer")
		require.Equal(t, tm.EndTime.Unix()-tm.PauseTime.Unix(), tm.Duration, "pause time not updated")
		require.True(t, tm.IsPaused, "timer not paused")
	}

	for _, timer := range timers {
		timerStorage.DeleteTimer(ctx, timer.ID)
	}
}
