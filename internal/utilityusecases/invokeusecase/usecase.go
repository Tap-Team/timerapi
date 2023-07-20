package invokeusecase

import (
	"context"

	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/pkg/exception"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

const _PROVIDER = "internal/utilityusecases/invokeusecase"

type TimerStorage interface {
	TimerWithSubscribers(ctx context.Context, offset, limit int) ([]*timermodel.TimerSubscribers, error)
}

type Subscriber interface {
	Subscribe(ctx context.Context, timerId uuid.UUID, userIds ...int64) error
}

type ManyTimersAdder interface {
	AddMany(ctx context.Context, timers map[uuid.UUID]int64) error
}

type UseCase struct {
	timersAdder  ManyTimersAdder
	subscriber   Subscriber
	timerStorage TimerStorage
}

func New(
	timerAdder ManyTimersAdder,
	subscriber Subscriber,
	timerStorage TimerStorage,
) *UseCase {
	return &UseCase{
		timersAdder:  timerAdder,
		subscriber:   subscriber,
		timerStorage: timerStorage,
	}
}

func (uc *UseCase) Invoke(ctx context.Context) error {
	var err error
	offset, limit := 0, 100
	errgr, gctx := errgroup.WithContext(ctx)
	for {
		timers, err := uc.timerStorage.TimerWithSubscribers(ctx, offset, limit)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("get timer with subscribers from storage", "Invoke", _PROVIDER))
		}
		if len(timers) == 0 {
			break
		}
		errgr.Go(func() error {
			err = uc.subscribe(gctx, timers)
			if err != nil {
				return err
			}
			return nil
		})
		offset += limit
	}
	err = errgr.Wait()
	if err != nil {
		return err
	}
	return nil
}

func (uc *UseCase) subscribe(ctx context.Context, timers []*timermodel.TimerSubscribers) error {
	var err error
	errgr, gctx := errgroup.WithContext(ctx)

	timersEndTime := make(map[uuid.UUID]int64, len(timers))
	for _, timer := range timers {
		timersEndTime[timer.ID] = timer.EndTime.Unix()
	}
	// in loop add timer ticker service timer with end time
	errgr.Go(func() error {
		err = uc.timersAdder.AddMany(gctx, timersEndTime)
		if err != nil {
			return exception.Wrap(err, exception.NewCause("add many timers in timersAdder", "subscribe", _PROVIDER))
		}
		return nil
	})

	// in loop subscribe user on timer in cache
	for _, timer := range timers {
		timer := *timer
		errgr.Go(func() error {
			err = uc.subscriber.Subscribe(gctx, timer.ID, timer.Subscribers...)
			if err != nil {
				return exception.Wrap(err, exception.NewCause("subscribe timer in cache storage", "subscribe", _PROVIDER))
			}
			return nil
		})
	}
	err = errgr.Wait()
	if err != nil {
		return err
	}
	return nil
}
