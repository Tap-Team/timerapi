package timerhandler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Tap-Team/timerapi/pkg/exception"
	"github.com/labstack/echo/v4"
)

/*
	group.PATCH("/:id/stop", handler.StopTimer(ctx))
	group.PATCH("/:id/start", handler.StartTimer(ctx))
	group.PATCH("/:id/reset", handler.ResetTimer(ctx))
*/

// StopTimer godoc
//
//	@Summary		StopTimer
//	@Description	stop timer by timer id, only owner can stop timer, every subscriber (creator inclusive) will be send stop event
//	@Tags			timers
//	@Param			vk_user_id	query	int64	true	"user id"
//	@Param			debug		query	string	false	"you can add secret key to query for debug requests"
//	@Param			id			path	string	true	"timer id"
//	@Success		204
//	@Failure		400	{object}	echoconfig.ErrorResponse
//	@Failure		404	{object}	echoconfig.ErrorResponse
//	@Failure		500	{object}	echoconfig.ErrorResponse
//	@Router			/timers/:id/stop [patch]
func (h *Handler) StopTimer(ctx context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		userId, timerId, err := userIdTimerId(c)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("parse user,timer id", "StopTimer", _PROVIDER))
		}
		pauseTime, err := strconv.ParseInt(c.QueryParam("pauseTime"), 10, 64)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("parse pause time", "StopTimer", _PROVIDER))
		}
		// delete timer by id uuid
		err = h.countdownTimerUseCase.Stop(ctx, timerId, userId, pauseTime)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("stop timer", "StopTimer", _PROVIDER))
		}
		return c.NoContent(http.StatusNoContent)
	}
}

// StartTimer godoc
//
//	@Summary		StartTimer
//	@Description	start timer by timer id, only owner can start timer, every subscriber (creator inclusive) will be send start event
//	@Tags			timers
//	@Param			vk_user_id	query	int64	true	"user id"
//	@Param			debug		query	string	false	"you can add secret key to query for debug requests"
//	@Param			id			path	string	true	"timer id"
//	@Produce		json
//	@Success		200	{object}	timermodel.Timer
//	@Failure		400	{object}	echoconfig.ErrorResponse
//	@Failure		404	{object}	echoconfig.ErrorResponse
//	@Failure		500	{object}	echoconfig.ErrorResponse
//	@Router			/timers/:id/start [patch]
func (h *Handler) StartTimer(ctx context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		userId, timerId, err := userIdTimerId(c)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("parse user,timer id", "StartTimer", _PROVIDER))
		}
		// start timer
		timer, err := h.countdownTimerUseCase.Start(ctx, timerId, userId)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("start timer", "StartTimer", _PROVIDER))
		}
		return c.JSON(http.StatusOK, timer)
	}
}

// ResetTimer godoc
//
//	@Summary		ResetTimer
//	@Description	reset timer by timer id, only owner can reset timer, every subscriber (creator inclusive) will be send reset event, if timer is started, reset not pause, only update end time
//	@Tags			timers
//	@Param			vk_user_id	query	int64	true	"user id"
//	@Param			debug		query	string	false	"you can add secret key to query for debug requests"
//	@Param			id			path	string	true	"timer id"
//	@Produce		json
//	@Success		200	{object}	timermodel.Timer
//	@Failure		400	{object}	echoconfig.ErrorResponse
//	@Failure		404	{object}	echoconfig.ErrorResponse
//	@Failure		500	{object}	echoconfig.ErrorResponse
//	@Router			/timers/:id/reset [patch]
func (h *Handler) ResetTimer(ctx context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		userId, timerId, err := userIdTimerId(c)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("parse user,timer id", "ResetTimer", _PROVIDER))
		}
		// start timer
		timer, err := h.countdownTimerUseCase.Reset(ctx, timerId, userId)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("reset timer", "ResetTimer", _PROVIDER))
		}
		return c.JSON(http.StatusOK, timer)
	}
}
