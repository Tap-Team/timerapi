package timerhandler_test

import (
	"context"
	"log"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Tap-Team/timerapi/internal/database/postgres/timerstorage"
	"github.com/Tap-Team/timerapi/internal/database/redis/subscriberstorage"
	"github.com/Tap-Team/timerapi/internal/domain/usecase/countdowntimerusecase"
	"github.com/Tap-Team/timerapi/internal/domain/usecase/timerusecase"
	"github.com/Tap-Team/timerapi/internal/model/notification"
	"github.com/Tap-Team/timerapi/internal/model/timerevent"
	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/internal/model/timermodel/timerfields"
	"github.com/Tap-Team/timerapi/internal/timerservice"
	"github.com/Tap-Team/timerapi/internal/transport/rest/timerhandler"
	"github.com/Tap-Team/timerapi/pkg/amidstr"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/Tap-Team/timerapi/pkg/postgres"
	"github.com/Tap-Team/timerapi/pkg/rediscontainer"
	"github.com/Tap-Team/timerapi/proto/timerservicepb"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const DEBUG_KEY = "d00d93aa-377f-48a8-a82e-1b3772172b13"

func debug(path string) string {
	if strings.Contains(path, "?") {
		return path + "&debug=" + DEBUG_KEY
	}
	return path + "?debug=" + DEBUG_KEY
}

func basePath(path string) string {
	return "/timers" + debug(path)
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

type timerSettingsOption func(t *timermodel.TimerSettings)

func randomTimerSettings(opts ...timerSettingsOption) *timermodel.TimerSettings {
	settings := timermodel.NewTimerSettings(
		timerfields.Name(amidstr.MakeString(timerfields.NameMaxSize)),
		timerfields.Description(amidstr.MakeString(timerfields.DescriptionMaxSize)),
		timerfields.BLUE,
		rand.Intn(3)%2 == 0,
		amidtime.DateTime(time.Now().Add(time.Duration(rand.Int31())*time.Second)),
	)
	for _, opt := range opts {
		opt(settings)
	}
	return settings
}

func compareTimerSettings(timer *timermodel.Timer, settings *timermodel.TimerSettings) bool {
	if timer.Color != settings.Color {
		return false
	}
	if timer.Name != settings.Name {
		return false
	}
	if timer.Description != settings.Description {
		return false
	}
	if timer.WithMusic != settings.WithMusic {
		return false
	}
	if timer.EndTime.Unix() != settings.EndTime.Unix() {
		return false
	}
	return true
}

const tickerServiceUrl = "localhost:50001"

var (
	e                     *echo.Echo = echo.New()
	handler               *timerhandler.Handler
	timerUseCase          timerhandler.TimerUseCase
	countdownTimerUseCase timerhandler.CountdownTimerUseCase
)

var (
	timerStorage      timerusecase.TimerStorage
	subscriberStorage timerusecase.SubscriberCacheStorage
	timerService      timerservice.TimerServiceClient
)

type EventSender interface {
	Send(event timerevent.TimerEvent)
}

type NotificationSender interface {
	Send(notification notification.Notification)
}

type ESender struct{}

func (s *ESender) Send(event timerevent.TimerEvent) {}

type NSender struct{}

func (s *NSender) Send(notification notification.Notification) {}

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
	timerService = timerservice.GrpcClient(timerservicepb.NewTimerServiceClient(conn))
	defer conn.Close()
	ts := timerstorage.New(p)
	timerStorage = ts
	subStorage := subscriberstorage.New(r)
	subscriberStorage = subStorage

	timerUseCase = timerusecase.New(
		timerStorage,
		subStorage,
		timerService,
		&ESender{},
		&NSender{},
	)

	countdownTimerUseCase = countdowntimerusecase.New(
		timerService,
		ts,
		&ESender{},
	)

	handler = timerhandler.New(countdownTimerUseCase, timerUseCase)
	m.Run()
}
