package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/Tap-Team/timerapi/internal/config"
	"github.com/Tap-Team/timerapi/internal/database/postgres/notificationstorage"
	"github.com/Tap-Team/timerapi/internal/database/postgres/timerstorage"
	"github.com/Tap-Team/timerapi/internal/database/redis/subscriberstorage"
	"github.com/Tap-Team/timerapi/internal/domain/datastream/timereventstream"
	"github.com/Tap-Team/timerapi/internal/domain/datastream/timernotificationstream"
	"github.com/Tap-Team/timerapi/internal/domain/usecase/countdowntimerusecase"
	"github.com/Tap-Team/timerapi/internal/domain/usecase/notificationusecase"
	"github.com/Tap-Team/timerapi/internal/domain/usecase/timerusecase"
	"github.com/Tap-Team/timerapi/internal/echoconfig"
	"github.com/Tap-Team/timerapi/internal/swagger"
	"github.com/Tap-Team/timerapi/internal/timerservice"
	"github.com/Tap-Team/timerapi/internal/utilityusecases/invokeusecase"
	"github.com/Tap-Team/timerapi/pkg/vk"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/Tap-Team/timerapi/internal/transport/bot"
	"github.com/Tap-Team/timerapi/internal/transport/rest/notificationhandler"
	"github.com/Tap-Team/timerapi/internal/transport/rest/timerhandler"
	"github.com/Tap-Team/timerapi/internal/transport/ws/timersocket"
	"github.com/Tap-Team/timerapi/pkg/postgres"
	"github.com/Tap-Team/timerapi/proto/timerservicepb"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Run() {
	os.Setenv("TZ", "UTC")

	config := config.FromFile("config/config.yaml")
	profilier(config.Profilier.Address())
	ctx := context.Background()

	e := echo.New()
	g := middleWare(e, config)

	p, err := postgres.New(config.Postgres.URL())
	if err != nil {
		log.Fatalf("error while connect to postgres, %s", err)
	}
	opts, err := redis.ParseURL(config.Redis.URL())
	if err != nil {
		fmt.Println(config.Redis.URL())
		log.Fatalf("error parse redis url, %s", err)
	}
	rc := redis.NewClient(opts)
	timerStorage := timerstorage.New(p)
	subscriberStorage := subscriberstorage.New(rc)
	notificationStorage := notificationstorage.New(p)

	timerService := tickerService(config.Ticker)

	notificationStream := timernotificationstream.New(
		timerService,
		timerStorage,
		subscriberStorage,
		notificationStorage,
	)
	go notificationStream.Start(ctx)

	eventSender := timereventstream.New()

	timerUseCase := timerusecase.New(
		timerStorage,
		subscriberStorage,
		timerService,
		eventSender,
		notificationStream,
	)
	countdowntimerUseCase := countdowntimerusecase.New(
		timerService,
		timerStorage,
		eventSender,
	)
	notificationUseCase := notificationusecase.New(
		notificationStorage,
	)

	err = invokeusecase.New(
		timerService,
		subscriberStorage,
		timerStorage,
	).Invoke(ctx)
	if err != nil {
		log.Printf("\nfailed execute invoke use case, %s", err)
	}

	timerhandler.Init(g, timerUseCase, countdowntimerUseCase)
	notificationhandler.Init(g, notificationUseCase)
	timersocket.Init(g, eventSender, notificationStream)

	botmanager := bot.NewManager(api.NewVK(config.VK.BotToken))
	go botmanager.RunMessageHandlers()
	go botmanager.RunNotificationBot(ctx, notificationStream)

	addr := config.Server.Address()

	h2s := &http2.Server{
		IdleTimeout: 10 * time.Second,
	}
	s := http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(e, h2s),
	}
	log.Printf("ECHO APP STARTED ON %s", addr)
	err = s.ListenAndServe()
	log.Fatalf("api server failed start failed, %s", err)
}

func middleWare(e *echo.Echo, config *config.Config) *echo.Group {
	e.HTTPErrorHandler = echoconfig.ErrorHandler
	swagger.New(e, config.Swagger)

	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	loggerMiddleWare(e)

	return e.Group("", vk.VkKeyHandler(config.VK.Key, config.VK.DebugKey))
}

func loggerMiddleWare(e *echo.Echo) {
	skippedUrls := []string{
		"/timers/user-created",
		"/timers/user-subscriptions",
	}
	loggerConfig := middleware.DefaultLoggerConfig
	loggerConfig.Skipper = func(c echo.Context) bool {
		for _, url := range skippedUrls {
			if c.Request().URL.Path == url {
				return true
			}
		}
		return false
	}
	e.Use(middleware.LoggerWithConfig(loggerConfig))
}

func tickerService(config config.TickerConfig) timerservice.TimerServiceClient {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, config.URL(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial context failed, %s", err)
	}
	return timerservice.GrpcClient(timerservicepb.NewTimerServiceClient(conn))
}
