package timerhandler

import (
	"context"
	"errors"
	"strconv"

	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/pkg/vk"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const _PROVIDER = "internal/transport/rest/timerhandler"

type TimerUseCase interface {
	Create(ctx context.Context, creator int64, timer *timermodel.CreateTimer) error
	Delete(ctx context.Context, timerId uuid.UUID, userId int64) error
	Update(ctx context.Context, timerId uuid.UUID, userId int64, timer *timermodel.TimerSettings) error
	Subscribe(ctx context.Context, timerId uuid.UUID, userId int64) (*timermodel.Timer, error)
	Unsubscribe(ctx context.Context, timerId uuid.UUID, userId int64) error

	TimerSubscribers(ctx context.Context, timerId uuid.UUID) ([]int64, error)

	UserSubscriptions(ctx context.Context, userId int64, offset, limit int) ([]*timermodel.Timer, error)
	UserCreatedTimers(ctx context.Context, userId int64, offset, limit int) ([]*timermodel.Timer, error)
}

type CountdownTimerUseCase interface {
	Stop(ctx context.Context, timerId uuid.UUID, userId int64, pauseTime int64) error
	Start(ctx context.Context, timerId uuid.UUID, userId int64) (*timermodel.Timer, error)
	Reset(ctx context.Context, timerId uuid.UUID, userId int64) (*timermodel.Timer, error)
}

type Handler struct {
	countdownTimerUseCase CountdownTimerUseCase
	timerUseCase          TimerUseCase
}

func New(countdownTimerUseCase CountdownTimerUseCase, timerUseCase TimerUseCase) *Handler {
	return &Handler{countdownTimerUseCase: countdownTimerUseCase, timerUseCase: timerUseCase}
}

func Init(e *echo.Group, timerUseCase TimerUseCase, countdownTimerUseCase CountdownTimerUseCase) {

	handler := &Handler{timerUseCase: timerUseCase, countdownTimerUseCase: countdownTimerUseCase}
	group := e.Group("/timers")
	ctx := context.Background()

	group.GET("/user-subscriptions", handler.UserSubscriptions(ctx))
	group.GET("/user-created", handler.UserCreated(ctx))
	group.GET("/:id/subscribers", handler.TimerSubscribers(ctx))

	group.POST("/create", handler.CreateTimer(ctx))
	group.DELETE("/:id", handler.DeleteTimer(ctx))
	group.PUT("/:id", handler.UpdateTimer(ctx))

	group.POST("/:id/subscribe", handler.Subscribe(ctx))
	group.POST("/:id/unsubscribe", handler.Unsubscribe(ctx))

	group.PATCH("/:id/stop", handler.StopTimer(ctx))
	group.PATCH("/:id/start", handler.StartTimer(ctx))
	group.PATCH("/:id/reset", handler.ResetTimer(ctx))
}

func offsetLimit(c echo.Context) (offset, limit int, err error) {
	offset, err = strconv.Atoi(c.QueryParam("offset"))
	if err != nil {
		return
	}
	limit, err = strconv.Atoi(c.QueryParam("limit"))
	if err != nil {
		return
	}
	return
}

func userIdTimerId(c echo.Context) (int64, uuid.UUID, error) {
	// parse vk_user_id
	userId, err := strconv.ParseInt(c.QueryParam(vk.USER_ID), 10, 64)
	if err != nil {
		return 0, uuid.Nil, errors.Join(err, errors.New("user id parse error"))
	}
	c.Request().URL.Parse(c.Request().URL.RawPath)
	// parse timer id from :id param
	timerId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return 0, uuid.Nil, errors.Join(err, errors.New("timer id parse error"))
	}
	return userId, timerId, nil
}
