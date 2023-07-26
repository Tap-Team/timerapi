package subscriberstorage

import (
	"context"
	"errors"

	"github.com/Tap-Team/timerapi/internal/errorutils/timererror"
	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/pkg/exception"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const _PROVIDER = "internal/database/redis/subscriberstorage"

type Storage struct {
	rc *redis.Client
}

func New(rc *redis.Client) *Storage {
	return &Storage{rc: rc}
}

func Error(err error, cause exception.Cause) error {
	if errors.Is(err, redis.Nil) {
		return exception.Wrap(timererror.ExceptionTimerSubscribersNotFound(), cause)
	}
	return exception.Wrap(err, cause)
}

func timerPrefix(timerId uuid.UUID) string {
	return "timer_" + timerId.String()
}

// get subscribers by timerId
func (s *Storage) TimerSubscribers(ctx context.Context, timerId uuid.UUID) (timermodel.Subscribers, error) {
	subscribers := make(timermodel.Subscribers)
	err := s.rc.Get(ctx, timerPrefix(timerId)).Scan(&subscribers)
	if err != nil {
		return nil, Error(err, exception.NewCause("get timer subscribers", "TimerSubscribers", _PROVIDER))
	}
	return subscribers, nil
}

func (s *Storage) Subscribe(ctx context.Context, timerId uuid.UUID, userIds ...int64) error {
	subscribers := make(timermodel.Subscribers)
	err := s.rc.Get(ctx, timerPrefix(timerId)).Scan(&subscribers)
	// if error with connection return error
	if err != nil && !errors.Is(err, redis.Nil) {
		return Error(err, exception.NewCause("get subscribers", "Subscribe", _PROVIDER))
	}
	// copy user id to subscribers map
	for _, userId := range userIds {
		if _, ok := subscribers[userId]; ok {
			return Error(timererror.ExceptionUserAlreadySubscriber(), exception.NewCause("add user in subscribers group", "Subscribe", _PROVIDER))
		}
		subscribers[userId] = struct{}{}
	}
	// set subscribers in redis
	err = s.rc.Set(ctx, timerPrefix(timerId), subscribers, 0).Err()
	if err != nil {
		return Error(err, exception.NewCause("set subscribers", "Subscribe", _PROVIDER))
	}
	return nil
}

func (s *Storage) Unsubscribe(ctx context.Context, timerId uuid.UUID, userId int64) error {
	subscribers := make(timermodel.Subscribers)
	err := s.rc.Get(ctx, timerPrefix(timerId)).Scan(&subscribers)
	if err != nil {
		return Error(err, exception.NewCause("get timer subscribers", "Unsubscribe", _PROVIDER))
	}
	delete(subscribers, userId)
	// if delete last subscriber, delete key from redis
	if len(subscribers) == 0 {
		err = s.rc.Del(ctx, timerPrefix(timerId)).Err()
		if err != nil {
			return Error(err, exception.NewCause("delete timer", "Unsubscribe", _PROVIDER))
		}
		return nil
	}
	err = s.rc.Set(ctx, timerPrefix(timerId), subscribers, 0).Err()
	if err != nil {
		return Error(err, exception.NewCause("set subscribers", "Unsubscribe", _PROVIDER))
	}
	return nil
}

func (s *Storage) DeleteTimer(ctx context.Context, timerId uuid.UUID) error {
	err := s.rc.Del(ctx, timerPrefix(timerId)).Err()
	if err != nil {
		return Error(err, exception.NewCause("delete timer", "DeleteTimer", _PROVIDER))
	}
	return nil
}
