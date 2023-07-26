package invokeusecase_test

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/Tap-Team/timerapi/internal/database/postgres/timerstorage"
	"github.com/Tap-Team/timerapi/internal/database/redis/subscriberstorage"
	"github.com/Tap-Team/timerapi/internal/domain/usecase/timerusecase"
	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/internal/model/timermodel/timerfields"
	"github.com/Tap-Team/timerapi/internal/timerservice"
	"github.com/Tap-Team/timerapi/internal/utilityusecases/invokeusecase"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/Tap-Team/timerapi/pkg/postgres"
	"github.com/Tap-Team/timerapi/pkg/rediscontainer"
	"github.com/Tap-Team/timerapi/proto/timerservicepb"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var tickerServiceUrl = "localhost:50001"

type TimerStorage interface {
	invokeusecase.TimerStorage
	timerusecase.TimerStorage
}

var (
	timerStorage      TimerStorage
	subscriberStorage timerusecase.SubscriberCacheStorage
	timerService      timerservice.TimerServiceClient
)

var (
	usecase *invokeusecase.UseCase
)

func TestMain(m *testing.M) {
	os.Setenv("TZ", "UTC")
	ctx := context.Background()
	p, termp, err := postgres.NewContainer(ctx, postgres.DEFAULT_MIGRATION_PATH)
	if err != nil {
		log.Fatalf("create postgres container failed, %s", err)
	}
	defer termp(ctx)
	r, termr, err := rediscontainer.New(ctx)
	if err != nil {
		log.Fatalf("create redis container failed, %s", err)
	}
	defer termr(ctx)
	conn, err := grpc.DialContext(ctx, tickerServiceUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial context failed, %s", err)
	}
	defer conn.Close()

	timerService = timerservice.GrpcClient(timerservicepb.NewTimerServiceClient(conn))
	timerStorage = timerstorage.New(p)
	subscriberStorage = subscriberstorage.New(r)

	usecase = invokeusecase.New(timerService, subscriberStorage, timerStorage)

	m.Run()
}

type randomTimerOption func(t *timermodel.Timer)

func randomTimer(opts ...randomTimerOption) *timermodel.Timer {
	duration := rand.Int31()
	timer := timermodel.NewTimer(
		uuid.New(),
		240,
		rand.Int63(),
		amidtime.DateTime(time.Now().Add(time.Second*time.Duration(duration))),
		amidtime.DateTime{},
		timerfields.DATE,
		"",
		"",
		timerfields.BLUE,
		rand.Intn(3)%2 == 0,
		int64(duration),
		false,
	)
	for _, opt := range opts {
		opt(timer)
	}
	return timer
}

func randomTimerList(size int, opts ...randomTimerOption) []*timermodel.Timer {
	tl := make([]*timermodel.Timer, 0, size)
	for i := 0; i < size; i++ {
		tl = append(tl, randomTimer(opts...))
	}
	return tl
}

func compareIds(ids []uuid.UUID, timers []*timermodel.Timer) (string, bool) {
	if len(ids) != len(timers) {
		return fmt.Sprintf("wrong len, len(ids) = %d, len(timers) = %d", len(ids), len(timers)), false
	}
	for i := range timers {
		if timers[i].ID != ids[i] {
			return fmt.Sprintf("uuid not equal, index %d, timers uuid %s, ids uuid %s", i, timers[i].ID, ids[i]), false
		}
	}
	return "", true
}

type TimerWithSubscribers struct {
	Timer       *timermodel.Timer
	Subscribers []int64
}

func TestInvoke(t *testing.T) {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tamount := 10
	subamount := 10
	creator := rand.Int63()

	duration := 5
	creatorEndTimeOpt := func(t *timermodel.Timer) {
		t.Creator = creator
		t.EndTime = amidtime.DateTime(time.Now().Add(time.Second * time.Duration(duration)))
		t.Duration = int64(duration)
	}

	timers := randomTimerList(tamount, creatorEndTimeOpt)
	timerSubscribers := make([]*TimerWithSubscribers, 0, tamount)
	// generate test data
	for _, timer := range timers {
		timer := timer
		if rand.Int63()%2 == 0 {
			err := timerStorage.InsertDateTimer(ctx, creator, timer.CreateTimer())
			require.NoError(t, err, "failed insert date timer")
		} else {
			err := timerStorage.InsertCountdownTimer(ctx, creator, timer.CreateTimer())
			require.NoError(t, err, "failed insert date timer")
		}
		subs := make([]int64, 0, subamount)
		subs = append(subs, creator)
		for i := 0; i < subamount-1; i++ {
			userId := rand.Int63()
			subs = append(subs, userId)

			// subscribe in postgres storage
			err := timerStorage.Subscribe(ctx, timer.ID, userId)
			require.NoError(t, err, "failed to subscribe user on timer")
		}
		timerSubscribers = append(timerSubscribers, &TimerWithSubscribers{
			Timer:       timer,
			Subscribers: subs,
		})

	}

	err = usecase.Invoke(ctx)
	require.NoError(t, err, "failed to invoke usecase")

	for _, timerSubs := range timerSubscribers {
		subscribers, err := subscriberStorage.TimerSubscribers(ctx, timerSubs.Timer.ID)
		require.NoError(t, err, "failed to get subscribers from storage")
		require.Equal(t, len(timerSubs.Subscribers), len(subscribers), "storage subscribers wrong length")
		for _, sub := range timerSubs.Subscribers {
			_, ok := subscribers[sub]
			require.True(t, ok, "subscriber not found, %s", sub)
		}
	}

	expCh, err := timerService.TimerTick(ctx)
	require.NoError(t, err, "failed to receive timers from ticker service")

	var ids []uuid.UUID
Loop:
	for {
		select {
		case <-ctx.Done():
			break Loop
		case <-time.After(time.Second * 2 * time.Duration(duration)):
			break Loop
		case timerIds, ok := <-expCh:
			if !ok {
				break
			}
			ids = append(ids, timerIds...)
		}
	}

	sort.Slice(ids, func(i, j int) bool {
		return ids[i].String() > ids[j].String()
	})
	sort.Slice(timers, func(i, j int) bool {
		return timers[i].ID.String() > timers[j].ID.String()
	})
	message, ok := compareIds(ids, timers)
	require.True(t, ok, message)

	for _, timer := range timers {
		timerStorage.DeleteTimer(ctx, timer.ID)
		subscriberStorage.DeleteTimer(ctx, timer.ID)
	}
}
