package timerhandler

import (
	"context"
	"net/http"

	"github.com/Tap-Team/timerapi/pkg/exception"
	"github.com/labstack/echo/v4"
)

// Subscribe godoc
//
//	@Summary		Subscribe
//	@Description	subscribe user on timer by id, user will see timer in subscriptions, get events and notificaitons
//	@Tags			timers
//	@Param			vk_user_id	query	int64	true	"user id"
//	@Param			id			path	string	true	"timer id"
//	@Param			debug		query	string	false	"you can add secret key to query for debug requests"
//	@Produce		json
//	@Success		200	{object}	timermodel.Timer
//	@Failure		400	{object}	echoconfig.ErrorResponse
//	@Failure		404	{object}	echoconfig.ErrorResponse
//	@Failure		500	{object}	echoconfig.ErrorResponse
//	@Router			/timers/:id/subscribe [post]
func (h *Handler) Subscribe(ctx context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		userId, timerId, err := userIdTimerId(c)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("parse user,timer id", "Subscribe", _PROVIDER))
		}
		timer, err := h.timerUseCase.Subscribe(ctx, timerId, userId)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("subscribe error", "Subscribe", _PROVIDER))
		}
		return c.JSON(http.StatusOK, timer)
	}
}

// Unsubscribe godoc
//
//	@Summary		Unsubscribe
//	@Description	unsubscribe user on timer by id, user wont see timer in subscriptions, get events and notificaitons
//	@Tags			timers
//	@Param			vk_user_id	query	int64	true	"user id"
//	@Param			id			path	string	true	"timer id"
//	@Param			debug		query	string	false	"you can add secret key to query for debug requests"
//	@Produce		json
//	@Success		204
//	@Failure		400	{object}	echoconfig.ErrorResponse
//	@Failure		404	{object}	echoconfig.ErrorResponse
//	@Failure		500	{object}	echoconfig.ErrorResponse
//	@Router			/timers/:id/unsubscribe [delete]
func (h *Handler) Unsubscribe(ctx context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		userId, timerId, err := userIdTimerId(c)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("parse user,timer id", "Unsubscribe", _PROVIDER))
		}
		err = h.timerUseCase.Unsubscribe(ctx, timerId, userId)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("unsubscribe error", "Unsubscribe", _PROVIDER))
		}
		return c.NoContent(http.StatusNoContent)
	}
}
