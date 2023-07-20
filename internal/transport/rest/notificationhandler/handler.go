package notificationhandler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Tap-Team/timerapi/internal/model/notification"
	"github.com/Tap-Team/timerapi/pkg/exception"
	"github.com/labstack/echo/v4"
)

const _PROVIDER = "internal/transport/rest/notificationhandler"

type NotificationUseCase interface {
	Delete(ctx context.Context, userId int64) error
	Notifications(ctx context.Context, userId int64) ([]*notification.NotificationDTO, error)
}

type Handler struct {
	useCase NotificationUseCase
}

func New(uc NotificationUseCase) *Handler {
	return &Handler{useCase: uc}
}

func Init(e *echo.Group, ntionUseCase NotificationUseCase) {
	ctx := context.Background()
	handler := &Handler{useCase: ntionUseCase}
	group := e.Group("/notifications")

	group.GET("", handler.NotificationsByUser(ctx))

	group.DELETE("", handler.Delete(ctx))
}

// NotificationsByUser godoc
//
//	@Summary		NotificationsByUser
//	@Description	get user unread notifications, notifications include delete or expire timer
//	@Tags			notifications
//	@Param			vk_user_id	query	int64	true	"user id"
//	@Param			debug		query	string	false	"you can add secret key to query for debug requests"
//	@Produce		json
//	@Success		200	{array}		notification.NotificationDTO
//	@Failure		400	{object}	echoconfig.ErrorResponse
//	@Failure		404	{object}	echoconfig.ErrorResponse
//	@Failure		500	{object}	echoconfig.ErrorResponse
//	@Router			/notifications [get]
func (h *Handler) NotificationsByUser(ctx context.Context) echo.HandlerFunc {
	f := func(c echo.Context) error {
		userId, err := strconv.ParseInt(c.QueryParam("vk_user_id"), 10, 64)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("parse userId param", "NotificationsByUser", _PROVIDER))
		}
		notifications, err := h.useCase.Notifications(ctx, userId)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("get notifications by user", "NotificationsByUser", _PROVIDER))
		}
		return c.JSON(http.StatusOK, notifications)
	}
	return f
}

// Delete godoc
//
//	@Summary		Delete
//	@Description	delete all user notifications
//	@Tags			notifications
//	@Param			vk_user_id	query	int64	true	"user id"
//	@Param			debug		query	string	false	"you can add secret key to query for debug requests"
//	@Produce		json
//	@Success		204
//	@Failure		400	{object}	echoconfig.ErrorResponse
//	@Failure		404	{object}	echoconfig.ErrorResponse
//	@Failure		500	{object}	echoconfig.ErrorResponse
//	@Router			/notifications [delete]
func (h *Handler) Delete(ctx context.Context) echo.HandlerFunc {
	f := func(c echo.Context) error {
		userId, err := strconv.ParseInt(c.QueryParam("vk_user_id"), 10, 64)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("parse userId param", "Delete", _PROVIDER))
		}
		err = h.useCase.Delete(ctx, userId)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("delete notification by userId", "Delete", _PROVIDER))
		}
		return c.NoContent(http.StatusNoContent)
	}
	return f
}
