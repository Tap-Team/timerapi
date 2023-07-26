package timerusecase_test

import (
	"context"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/Tap-Team/timerapi/internal/database/postgres/timerstorage"
	"github.com/Tap-Team/timerapi/internal/database/redis/subscriberstorage"
	"github.com/Tap-Team/timerapi/internal/domain/usecase/timerusecase"
	notification "github.com/Tap-Team/timerapi/internal/model/notification"
	timerevent "github.com/Tap-Team/timerapi/internal/model/timerevent"
	timermodel "github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/internal/model/timermodel/timerfields"
	"github.com/Tap-Team/timerapi/internal/timerservice"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/Tap-Team/timerapi/pkg/postgres"
	"github.com/Tap-Team/timerapi/pkg/rediscontainer"
	"github.com/Tap-Team/timerapi/proto/timerservicepb"
	uuid "github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

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

type ESender struct{}

func (s ESender) Send(timerevent.TimerEvent) {}

type NSender struct{}

func (s NSender) Send(notification.Notification) {}

var (
	timerStorage      timerusecase.TimerStorage
	subscriberStorage timerusecase.SubscriberCacheStorage
	esender           timerusecase.EventSender        = ESender{}
	nsender           timerusecase.NotificationSender = NSender{}
	timerService      timerservice.TimerServiceClient
)

const (
	path             = postgres.DEFAULT_MIGRATION_PATH
	tickerServiceUrl = "localhost:50001"
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p, term, err := postgres.NewContainer(ctx, path)
	if err != nil {
		log.Fatalf("failed to start postgres container, %s", err)
	}
	defer term(ctx)
	r, term, err := rediscontainer.New(ctx)
	if err != nil {
		log.Fatalf("failed to start redis container, %s", err)
	}
	timerStorage = timerstorage.New(p)
	subscriberStorage = subscriberstorage.New(r)
	conn, err := grpc.DialContext(ctx, tickerServiceUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial context failed, %s", err)
	}
	timerService = timerservice.GrpcClient(timerservicepb.NewTimerServiceClient(conn))

	m.Run()
}
