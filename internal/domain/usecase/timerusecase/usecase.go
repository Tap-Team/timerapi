package timerusecase

import (
	"context"
	"time"

	"github.com/Tap-Team/timerapi/internal/errorutils/timererror"
	"github.com/Tap-Team/timerapi/internal/model/notification"
	"github.com/Tap-Team/timerapi/internal/model/timerevent"
	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/internal/model/timermodel/timerfields"
	"github.com/Tap-Team/timerapi/internal/timerservice"
	"github.com/Tap-Team/timerapi/pkg/exception"
	"github.com/Tap-Team/timerapi/pkg/saga"
	"github.com/google/uuid"
)

const _PROVIDER = "internal/domain/timerusecase"

type TimerStorage interface {
	InsertDateTimer(ctx context.Context, creator int64, timer *timermodel.CreateTimer) error
	InsertCountdownTimer(ctx context.Context, creator int64, timer *timermodel.CreateTimer) error
	UpdateTimer(ctx context.Context, timerId uuid.UUID, timerSettings *timermodel.TimerSettings) error
	DeleteTimer(ctx context.Context, id uuid.UUID) error
	Timer(ctx context.Context, timerId uuid.UUID) (*timermodel.Timer, error)

	UserTimers(ctx context.Context, userId int64, limit, offset int) ([]*timermodel.Timer, error)
	UserCreatedTimers(ctx context.Context, userId int64, offset, limit int) ([]*timermodel.Timer, error)
	UserSubscriptions(ctx context.Context, userId int64, offset, limit int) ([]*timermodel.Timer, error)

	Subscribe(ctx context.Context, timerId uuid.UUID, userId int64) error
	Unsubscribe(ctx context.Context, timerId uuid.UUID, userId int64) error
}

type SubscriberCacheStorage interface {
	DeleteTimer(ctx context.Context, id uuid.UUID) error
	TimerSubscribers(ctx context.Context, timerId uuid.UUID) (timermodel.Subscribers, error)
	Subscribe(ctx context.Context, timerId uuid.UUID, userIds ...int64) error
	Unsubscribe(ctx context.Context, timerId uuid.UUID, userId int64) error
}

type EventSender interface {
	Send(event timerevent.TimerEvent)
}

type NotificationSender interface {
	Send(notification notification.Notification)
}

type UseCase struct {
	timerStorage      TimerStorage
	subscriberStorage SubscriberCacheStorage
	timerService      timerservice.TimerServiceClient
	esender           EventSender
	nsender           NotificationSender
}

func New(
	timerStorage TimerStorage,
	timerCache SubscriberCacheStorage,
	timerService timerservice.TimerServiceClient,
	esender EventSender,
	nsender NotificationSender,
) *UseCase {
	return &UseCase{timerStorage: timerStorage, subscriberStorage: timerCache, timerService: timerService, esender: esender, nsender: nsender}
}

func (uc *UseCase) UserSubscriptions(ctx context.Context, userId int64, offset, limit int) ([]*timermodel.Timer, error) {
	timers, err := uc.timerStorage.UserSubscriptions(ctx, userId, offset, limit)
	if err != nil {
		return nil, exception.Wrap(err, exception.NewCause("get timers from storage", "UserSubscriptions", _PROVIDER))
	}
	return timers, nil
}

func (uc *UseCase) UserCreatedTimers(ctx context.Context, userId int64, offset, limit int) ([]*timermodel.Timer, error) {
	timers, err := uc.timerStorage.UserCreatedTimers(ctx, userId, offset, limit)
	if err != nil {
		return nil, exception.Wrap(err, exception.NewCause("get timers from storage", "UserCreatedTimers", _PROVIDER))
	}
	return timers, nil
}

func (uc *UseCase) UserTimers(ctx context.Context, userId int64, offset, limit int) ([]*timermodel.Timer, error) {
	timers, err := uc.timerStorage.UserTimers(ctx, userId, offset, limit)
	if err != nil {
		return nil, exception.Wrap(err, exception.NewCause("get user timers from storage", "UserTimers", _PROVIDER))
	}
	return timers, nil
}

func (uc *UseCase) TimerSubscribers(ctx context.Context, timerId uuid.UUID) ([]int64, error) {
	subscribers, err := uc.subscriberStorage.TimerSubscribers(ctx, timerId)
	if err != nil {
		return nil, exception.Wrap(err, exception.NewCause("get timer subscribers", "TimerSubscribers", _PROVIDER))
	}
	return subscribers.Array(), nil
}

func (uc *UseCase) Timer(ctx context.Context, timerId uuid.UUID) (*timermodel.Timer, error) {
	timer, err := uc.timerStorage.Timer(ctx, timerId)
	if err != nil {
		return nil, exception.Wrap(err, exception.NewCause("get timer from storage", "Timer", _PROVIDER))
	}
	return timer, nil
}

func (uc *UseCase) Create(ctx context.Context, creator int64, timer *timermodel.CreateTimer) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	var err error
	var saga saga.Saga
	defer saga.Rollback()

	// create timer into storage
	switch timer.Type {
	case timerfields.DATE:
		err := uc.timerStorage.InsertDateTimer(ctx, creator, timer)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("create date timer into storage", "Create", _PROVIDER))
		}
	case timerfields.COUNTDOWN:
		err := uc.timerStorage.InsertCountdownTimer(ctx, creator, timer)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("create countdown timer into storage", "Create", _PROVIDER))
		}
	}
	saga.Register(func() {
		uc.timerStorage.DeleteTimer(ctx, timer.ID)
	})

	// subscribe creator to own timer in subscriberStorage
	err = uc.subscriberStorage.Subscribe(ctx, timer.ID, creator)
	if err != nil {
		return exception.Wrap(err, exception.NewCause("subscribe creator to own timer", "Create", _PROVIDER))
	}
	saga.Register(func() {
		uc.subscriberStorage.DeleteTimer(ctx, timer.ID)
	})
	// add timer end time in timer service
	err = uc.timerService.Add(ctx, timer.ID, timer.EndTime.Unix())
	if err != nil {
		return exception.Wrap(err, exception.NewCause("add timer end time to timerService", "Create", _PROVIDER))
	}
	saga.Register(func() {
		uc.timerService.Remove(ctx, timer.ID)
	})

	// if err == nil set saga state is ok
	saga.OK()
	return nil
}

func (uc *UseCase) Delete(ctx context.Context, timerId uuid.UUID, userId int64) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	var err error
	// create new saga
	var saga saga.Saga
	// defer saga was rollback if not all ok
	defer saga.Rollback()
	// check access user to timer
	timer, err := uc.checkAccess(ctx, userId, timerId)
	if err != nil {
		return exception.Wrap(err, exception.NewCause("check access", "Delete", _PROVIDER))
	}

	// delete timer from service if not paused
	if !timer.IsPaused {
		uc.timerService.Remove(ctx, timerId)
	}
	// delete timer from storage
	err = uc.timerStorage.DeleteTimer(ctx, timerId)
	if err != nil {
		return exception.Wrap(err, exception.NewCause("delete timer from storage", "Delete", _PROVIDER))
	}
	// send delete event to event handler
	uc.nsender.Send(notification.NewDelete(*timer))
	// if all ok send saga ok
	saga.OK()
	return nil
}

func (uc *UseCase) Update(ctx context.Context, timerId uuid.UUID, userId int64, settings *timermodel.TimerSettings) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	var saga saga.Saga
	defer saga.Rollback()
	timer, err := uc.checkAccess(ctx, userId, timerId)
	if err != nil {
		return exception.Wrap(err, exception.NewCause("check access", "Update", _PROVIDER))
	}
	err = uc.timerStorage.UpdateTimer(ctx, timerId, settings)
	if err != nil {
		return exception.Wrap(err, exception.NewCause("update timer in storage", "Update", _PROVIDER))
	}
	saga.Register(func() {
		uc.timerStorage.UpdateTimer(ctx, timer.ID, timermodel.NewTimerSettings(timer.Name, timer.Description, timer.Color, timer.WithMusic, timer.EndTime))
	})
	if timer.EndTime != settings.EndTime {
		err = uc.timerService.Update(ctx, timerId, settings.EndTime.Unix())
		if err != nil {
			return exception.Wrap(err, exception.Wrap(err, exception.NewCause("update end time in timerservice", "Update", _PROVIDER)))
		}
	}
	uc.esender.Send(timerevent.NewUpdate(timerId, *settings))
	saga.OK()
	return nil
}

func (uc *UseCase) checkAccess(ctx context.Context, userId int64, timerId uuid.UUID) (*timermodel.Timer, error) {
	timer, err := uc.timerStorage.Timer(ctx, timerId)
	if err != nil {
		return nil, exception.Wrap(err, exception.NewCause("get timer by id", "checkAccess", _PROVIDER))
	}
	if timer.Creator != userId {
		return nil, exception.Wrap(timererror.ExceptionUserForbidden(), exception.NewCause("compare creator and userId", "checkAccess", _PROVIDER))
	}
	return timer, nil
}

func (uc *UseCase) Subscribe(ctx context.Context, timerId uuid.UUID, userId int64) (*timermodel.Timer, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	var err error
	// get timer to check subscriber not owner
	timer, err := uc.timerStorage.Timer(ctx, timerId)
	if err != nil {
		return nil, exception.Wrap(err, exception.NewCause("get timer by id", "Unsubscribe", _PROVIDER))
	}
	if timer.Creator == userId {
		return nil, exception.Wrap(timererror.ExceptionUserAlreadySubscriber(), exception.NewCause("unsubscribe timer", "Unsubscribe", _PROVIDER))
	}
	// create new saga
	var saga saga.Saga
	// defer saga was rollback if not all ok
	defer saga.Rollback()

	// subscribe in subscriber cache storage
	err = uc.subscriberStorage.Subscribe(ctx, timerId, userId)
	if err != nil {
		return nil, exception.Wrap(err, exception.NewCause("subscribe in cache storage", "Subscribe", _PROVIDER))
	}
	// register rollback
	saga.Register(func() { uc.subscriberStorage.Unsubscribe(ctx, timerId, userId) })

	// subscribe in timerStorage
	err = uc.timerStorage.Subscribe(ctx, timerId, userId)
	if err != nil {
		return nil, exception.Wrap(err, exception.NewCause("subscribe in timer storage", "Subscribe", _PROVIDER))
	}
	// register rollback
	saga.Register(func() { uc.subscriberStorage.Unsubscribe(ctx, timerId, userId) })

	saga.OK()
	return timer, nil
}

func (uc *UseCase) Unsubscribe(ctx context.Context, timerId uuid.UUID, userId int64) error {
	var err error
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	timer, err := uc.timerStorage.Timer(ctx, timerId)
	if err != nil {
		return exception.Wrap(err, exception.NewCause("get timer from storage", "Unsubscribe", _PROVIDER))
	}
	if timer.Creator == userId {
		return exception.Wrap(timererror.ExceptionCreatorUnsubscribe(), exception.NewCause("unsubscribe timer", "Unsubscribe", _PROVIDER))
	}
	// create new saga
	var saga saga.Saga
	// defer saga was rollback if not all ok
	defer saga.Rollback()

	// unsubscribe in subscriber cache storage
	err = uc.subscriberStorage.Unsubscribe(ctx, timerId, userId)
	if err != nil {
		return exception.Wrap(err, exception.NewCause("unsubscribe in cache storage", "Unsubscribe", _PROVIDER))
	}
	// register rollback
	saga.Register(func() { uc.subscriberStorage.Subscribe(ctx, timerId, userId) })

	// unsubscribe in timer storage
	err = uc.timerStorage.Unsubscribe(ctx, timerId, userId)
	if err != nil {
		return exception.Wrap(err, exception.NewCause("unsubscribe in timer storage", "Unsubscribe", _PROVIDER))
	}
	// register rollback
	saga.Register(func() { uc.subscriberStorage.Subscribe(ctx, timerId, userId) })

	saga.OK()
	return nil
}
