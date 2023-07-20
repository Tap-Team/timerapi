package timernotificationstream

import (
	"context"

	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/google/uuid"
)

type TimerStorageStub struct{}

func (s *TimerStorageStub) DeleteTimer(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (s *TimerStorageStub) Timer(ctx context.Context, timerId uuid.UUID) (*timermodel.Timer, error) {
	return new(timermodel.Timer), nil
}
func (s *TimerStorageStub) Unsubscribe(ctx context.Context, timerId uuid.UUID, userId int64) error {
	return nil
}
func (s *TimerStorageStub) UpdatePauseTime(ctx context.Context, timerId uuid.UUID, pauseTime amidtime.DateTime, isPaused bool) error {
	return nil
}

type SubscriberCacheStorageStub struct {
}

func (s *SubscriberCacheStorageStub) DeleteTimer(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (s *SubscriberCacheStorageStub) TimerSubscribers(ctx context.Context, timerId uuid.UUID) (timermodel.Subscribers, error) {
	return make(timermodel.Subscribers), nil
}
