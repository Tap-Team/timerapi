package timerhandler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/pkg/exception"
	"github.com/Tap-Team/timerapi/pkg/vk"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

/*
	group.GET("/user", handler.TimersByUser(ctx))
	group.GET("/:id/subscribers", handler.TimerSubscribers(ctx))

	group.POST("/create", handler.CreateTimer(ctx))
	group.DELETE("/:id", handler.DeleteTimer(ctx))
	group.PUT("/:id", handler.UpdateTimer(ctx))
*/

// // TimersByUser godoc
// //
// //	@Summary		TimersByUser
// //	@Description	get all user timers with offset and limit, timers include created by user and user subscriptions
// //	@Tags			timers
// //	@Param			debug		query	string	false	"you can add secret key to query for debug requests"
// //	@Param			vk_user_id	query	int64	true	"user id"
// //	@Param			offset		query	int64	true	"offset"
// //	@Param			limit		query	int64	true	"limit"
// //	@Produce		json
// //	@Success		200	{array}		timermodel.Timer
// //	@Failure		400	{object}	echoconfig.ErrorResponse
// //	@Failure		404	{object}	echoconfig.ErrorResponse
// //	@Failure		500	{object}	echoconfig.ErrorResponse
// //	@Router			/timers/user [get]
// func (h *Handler) TimersByUser(ctx context.Context) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		// parse offset and limit query
// 		offset, limit, err := offsetLimit(c)
// 		if err != nil {
// 			return exception.Wrap(err, exception.NewCause("parse offset limit", "TimersByUser", _PROVIDER))
// 		}
// 		// parse vk_user_id
// 		userId, err := strconv.ParseInt(c.QueryParam(vk.USER_ID), 10, 64)
// 		if err != nil {
// 			return exception.Wrap(err, exception.NewCause("parse userId", "TimersByUser", _PROVIDER))
// 		}

// 		// get timers from use case
// 		timers, err := h.timerUseCase.UserTimers(ctx, userId, offset, limit)
// 		if err != nil {
// 			return exception.Wrap(err, exception.NewCause("get user timers error", "TimersByUser", _PROVIDER))
// 		}
// 		return c.JSON(http.StatusOK, timers)
// 	}
// }

// UserSubscriptions godoc
//
//	@Summary		UserSubscriptions
//	@Description	get user subscriptions with offset and limit
//	@Tags			timers
//	@Param			debug		query	string	false	"you can add secret key to query for debug requests"
//	@Param			vk_user_id	query	int64	true	"user id"
//	@Param			offset		query	int64	true	"offset"
//	@Param			limit		query	int64	true	"limit"
//	@Produce		json
//	@Success		200	{array}		timermodel.Timer
//	@Failure		400	{object}	echoconfig.ErrorResponse
//	@Failure		404	{object}	echoconfig.ErrorResponse
//	@Failure		500	{object}	echoconfig.ErrorResponse
//	@Router			/timers/user-subscriptions [get]
func (h *Handler) UserSubscriptions(ctx context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		// parse offset and limit query
		offset, limit, err := offsetLimit(c)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("parse offset limit", "UserSubscriptions", _PROVIDER))
		}
		// parse vk_user_id
		userId, err := strconv.ParseInt(c.QueryParam(vk.USER_ID), 10, 64)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("parse userId", "UserSubscriptions", _PROVIDER))
		}

		// get timers from use case
		timers, err := h.timerUseCase.UserSubscriptions(ctx, userId, offset, limit)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("get user timers error", "UserSubscriptions", _PROVIDER))
		}
		return c.JSON(http.StatusOK, timers)
	}
}

// UserCreated godoc
//
//	@Summary		UserCreated
//	@Description	get user created timers with offset and limit
//	@Tags			timers
//	@Param			debug		query	string	false	"you can add secret key to query for debug requests"
//	@Param			vk_user_id	query	int64	true	"user id"
//	@Param			offset		query	int64	true	"offset"
//	@Param			limit		query	int64	true	"limit"
//	@Produce		json
//	@Success		200	{array}		timermodel.Timer
//	@Failure		400	{object}	echoconfig.ErrorResponse
//	@Failure		404	{object}	echoconfig.ErrorResponse
//	@Failure		500	{object}	echoconfig.ErrorResponse
//	@Router			/timers/user-created [get]
func (h *Handler) UserCreated(ctx context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		// parse offset and limit query
		offset, limit, err := offsetLimit(c)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("parse offset limit", "UserCreated", _PROVIDER))
		}
		// parse vk_user_id
		userId, err := strconv.ParseInt(c.QueryParam(vk.USER_ID), 10, 64)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("parse userId", "UserCreated", _PROVIDER))
		}

		// get timers from use case
		timers, err := h.timerUseCase.UserCreatedTimers(ctx, userId, offset, limit)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("get user timers error", "UserCreated", _PROVIDER))
		}
		return c.JSON(http.StatusOK, timers)
	}
}

// TimerSubscribers godoc
//
//	@Summary		TimerSubscribers
//	@Description	return array of id users which subscribe on timer
//	@Tags			timers
//	@Param			debug		query	string	false	"you can add secret key to query for debug requests"
//	@Param			vk_user_id	query	int64	true	"user id"
//	@Param			id			path	string	true	"timer id"
//	@Produce		json
//	@Success		200	{array}		int64
//	@Failure		400	{object}	echoconfig.ErrorResponse
//	@Failure		404	{object}	echoconfig.ErrorResponse
//	@Failure		500	{object}	echoconfig.ErrorResponse
//	@Router			/timers/{id}/subscribers [get]
func (h *Handler) TimerSubscribers(ctx context.Context) echo.HandlerFunc {
	f := func(c echo.Context) error {
		timerId, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return exception.Wrap(err, exception.NewCause("parse timerid", "TimerSubscribers", _PROVIDER))
		}
		subscribers, err := h.timerUseCase.TimerSubscribers(ctx, timerId)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("get timer subscribers", "TimerSubscribers", _PROVIDER))
		}
		return c.JSON(http.StatusOK, subscribers)
	}
	return f
}

// CreateTimer godoc
//
//	@Summary		CreateTimer
//	@Description	create user timer
//	@Tags			timers
//	@Param			debug		query	string					false	"you can add secret key to query for debug requests"
//	@Param			vk_user_id	query	int64					true	"user id"
//	@Param			timer		body	timermodel.CreateTimer	true	"timer"
//	@Produce		json
//	@Accept			json
//	@Success		201
//	@Failure		400	{object}	echoconfig.ErrorResponse
//	@Failure		404	{object}	echoconfig.ErrorResponse
//	@Failure		500	{object}	echoconfig.ErrorResponse
//	@Router			/timers/create [post]
func (h *Handler) CreateTimer(ctx context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		// parse vk_user_id
		userId, err := strconv.ParseInt(c.QueryParam(vk.USER_ID), 10, 64)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("parse userId", "CreateTimer", _PROVIDER))
		}

		// bind body
		timer := new(timermodel.CreateTimer)
		err = c.Bind(timer)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("bind body", "CreateTimer", _PROVIDER))
		}
		// validate body
		err = timer.Validate()
		if err != nil {
			return exception.Wrap(err, exception.NewCause("validate body", "CreateTimer", _PROVIDER))
		}
		err = h.timerUseCase.Create(ctx, userId, timer)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("create timer", "CreateTimer", _PROVIDER))
		}
		return c.NoContent(http.StatusCreated)
	}
}

// DeleteTimer godoc
//
//	@Summary		DeleteTimer
//	@Description	delete user timer
//	@Tags			timers
//	@Param			debug		query	string	false	"you can add secret key to query for debug requests"
//	@Param			vk_user_id	query	int64	true	"user id"
//	@Param			id			path	string	true	"timer id"
//	@Produce		json
//	@Success		204
//	@Failure		400	{object}	echoconfig.ErrorResponse
//	@Failure		404	{object}	echoconfig.ErrorResponse
//	@Failure		500	{object}	echoconfig.ErrorResponse
//	@Router			/timers/{id} [delete]
func (h *Handler) DeleteTimer(ctx context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		userId, timerId, err := userIdTimerId(c)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("parse user,timer id", "DeleteTimer", _PROVIDER))
		}
		// delete timer by id uuid
		err = h.timerUseCase.Delete(ctx, timerId, userId)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("delete timer", "DeleteTimer", _PROVIDER))
		}
		return c.NoContent(http.StatusNoContent)
	}
}

// UpdateTimer godoc
//
//	@Summary		UpdateTimer
//	@Description	update user timer
//	@Tags			timers
//	@Param			debug		query	string						false	"you can add secret key to query for debug requests"
//	@Param			vk_user_id	query	int64						true	"user id"
//	@Param			id			path	string						true	"timer id"
//	@Param			settings	body	timermodel.TimerSettings	true	"timer update settings"
//	@Accept			json
//	@Produce		json
//	@Success		204
//	@Failure		400	{object}	echoconfig.ErrorResponse
//	@Failure		404	{object}	echoconfig.ErrorResponse
//	@Failure		500	{object}	echoconfig.ErrorResponse
//	@Router			/timers/{id} [put]
func (h *Handler) UpdateTimer(ctx context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		// parse user timer id
		userId, timerId, err := userIdTimerId(c)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("parse user,timer id", "UpdateTimer", _PROVIDER))
		}
		// parse body
		settings := new(timermodel.TimerSettings)
		err = c.Bind(settings)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("parse body", "UpdateTimer", _PROVIDER))
		}
		// update timer by id uuid
		err = h.timerUseCase.Update(ctx, timerId, userId, settings)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("update timer", "UpdateTimer", _PROVIDER))
		}
		return c.NoContent(http.StatusNoContent)
	}
}
