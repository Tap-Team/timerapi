package notificationusecase

import (
	"context"

	"github.com/Tap-Team/timerapi/internal/model/notification"
	"github.com/Tap-Team/timerapi/pkg/exception"
)

const _PROVIDER = "internal/domain/usecase/notificationusecase"

type NotificationStorage interface {
	UserNotifications(ctx context.Context, userId int64) ([]*notification.NotificationDTO, error)
	DeleteUserNotifications(ctx context.Context, userId int64) error
}

type UseCase struct {
	nstorage NotificationStorage
}

func New(nstorage NotificationStorage) *UseCase {
	return &UseCase{nstorage: nstorage}
}

func (uc *UseCase) Delete(ctx context.Context, userId int64) error {
	err := uc.nstorage.DeleteUserNotifications(ctx, userId)
	if err != nil {
		return exception.Wrap(err, exception.NewCause("delete user notifications", "Delete", _PROVIDER))
	}
	return nil
}

func (uc *UseCase) Notifications(ctx context.Context, userId int64) ([]*notification.NotificationDTO, error) {
	notifications, err := uc.nstorage.UserNotifications(ctx, userId)
	if err != nil {
		return nil, exception.Wrap(err, exception.NewCause("get user notifications", "Notifications", _PROVIDER))
	}
	return notifications, nil
}
