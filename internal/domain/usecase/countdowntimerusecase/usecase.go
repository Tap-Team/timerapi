package countdowntimerusecase

import (
	"context"
	"time"

	"github.com/Tap-Team/timerapi/internal/errorutils/timererror"
	"github.com/Tap-Team/timerapi/internal/model/timerevent"
	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/internal/timerservice"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/Tap-Team/timerapi/pkg/exception"
	"github.com/Tap-Team/timerapi/pkg/saga"
	"github.com/google/uuid"
)

const _PROVIDER = "internal/domain/usecase/countdowntimerusecase"

type TimerUpdater interface {
	CountdownTimer(ctx context.Context, timerId uuid.UUID) (*timermodel.CountdownTimer, error)
	UpdateTime(ctx context.Context, timerId uuid.UUID, endTime amidtime.DateTime) error
	UpdatePauseTime(ctx context.Context, timerId uuid.UUID, pauseTime amidtime.DateTime, isPaused bool) error
	TimerPause(ctx context.Context, timerId uuid.UUID) (*timermodel.TimerPause, error)
}

type EventSender interface {
	Send(event timerevent.TimerEvent)
}

type UseCase struct {
	timerService timerservice.TimerServiceClient
	updater      TimerUpdater
	sender       EventSender
}

func New(
	timerService timerservice.TimerServiceClient,
	updater TimerUpdater,
	sender EventSender,
) *UseCase {
	return &UseCase{
		timerService: timerService,
		updater:      updater,
		sender:       sender,
	}
}

func (uc *UseCase) Stop(ctx context.Context, timerId uuid.UUID, userId int64, pauseTime int64) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	var err error
	timer, err := uc.checkCountDownTimer(ctx, timerId, userId)
	if err != nil {
		return exception.Wrap(err, exception.NewCause("check timer", "Stop", _PROVIDER))
	}
	if timer.IsPaused {
		return exception.Wrap(timererror.ExceptionTimerIsPaused(), exception.NewCause("check timer not paused", "Stop", _PROVIDER))
	}

	saga := new(saga.Saga)
	defer saga.Rollback()

	ptime := amidtime.DateTime(time.Unix(pauseTime, 0))

	// stop timer in timer service
	err = uc.timerService.Stop(ctx, timerId)
	if err != nil {
		return exception.Wrap(err, exception.NewCause("stop in timer service", "Stop", _PROVIDER))
	}
	saga.Register(func() {
		uc.timerService.Start(ctx, timerId, timer.EndTime.Unix())
	})

	// set pause time in storage
	err = uc.updater.UpdatePauseTime(ctx, timerId, ptime, true)
	if err != nil {
		return exception.Wrap(err, exception.NewCause("update pause time in storage", "Stop", _PROVIDER))
	}
	saga.Register(func() {
		uc.updater.UpdatePauseTime(ctx, timerId, amidtime.DateTime{}, false)
	})

	uc.sender.Send(timerevent.NewStop(timerId, ptime))
	saga.OK()
	return nil
}

func (uc *UseCase) Start(ctx context.Context, timerId uuid.UUID, userId int64) (*timermodel.Timer, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	var err error
	timer, err := uc.checkCountDownTimer(ctx, timerId, userId)
	if err != nil {
		return nil, exception.Wrap(err, exception.NewCause("check timer", "Stop", _PROVIDER))
	}
	if !timer.IsPaused {
		return nil, exception.Wrap(timererror.ExceptionTimerIsPlaying(), exception.NewCause("check timer is paused", "Stop", _PROVIDER))
	}

	timeInPause := time.Since(timer.PauseTime.T())
	// count the time for which the timer was stopped
	endTime := amidtime.DateTime(timer.EndTime.T().Add(timeInPause))

	saga := new(saga.Saga)
	defer saga.Rollback()

	err = uc.updater.UpdateTime(ctx, timerId, endTime)
	if err != nil {
		return nil, exception.Wrap(err, exception.NewCause("update end time in storage", "Start", _PROVIDER))
	}
	saga.Register(func() { uc.updater.UpdateTime(ctx, timerId, timer.EndTime) })

	// update status in storage
	err = uc.updater.UpdatePauseTime(ctx, timerId, amidtime.DateTime{}, false)
	if err != nil {
		return nil, exception.Wrap(err, exception.NewCause("update pause time in storage", "Start", _PROVIDER))
	}
	saga.Register(func() { uc.updater.UpdatePauseTime(ctx, timerId, timer.PauseTime, true) })

	// start timer in timer service
	err = uc.timerService.Start(ctx, timerId, endTime.Unix())
	if err != nil {
		return nil, exception.Wrap(err, exception.NewCause("start timer in timer service", "Start", _PROVIDER))
	}
	saga.Register(func() { uc.timerService.Stop(ctx, timerId) })

	uc.sender.Send(timerevent.NewStart(timerId, endTime))
	saga.OK()

	// if all ok, change timer fields to according database and return timer
	t := &timer.Timer
	// set end time like in database
	t.EndTime = endTime
	// set pause time to zero and isPaused to false
	t.PauseTime = amidtime.DateTime{}
	t.IsPaused = false

	return t, nil
}

func (uc *UseCase) Reset(ctx context.Context, timerId uuid.UUID, userId int64) (*timermodel.Timer, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	var err error

	// check timer
	timer, err := uc.checkCountDownTimer(ctx, timerId, userId)
	if err != nil {
		return nil, exception.Wrap(err, exception.NewCause("check timer", "Reset", _PROVIDER))
	}

	// add timer duration to end time
	endTime := amidtime.DateTime(time.Now().Add(time.Second * time.Duration(timer.Duration)))

	// update time in database
	err = uc.updater.UpdateTime(ctx, timerId, endTime)
	if err != nil {
		return nil, exception.Wrap(err, exception.Wrap(err, exception.NewCause("update timer time", "Reset", _PROVIDER)))
	}
	// send reset event
	uc.sender.Send(timerevent.NewReset(timerId, endTime))

	t := &timer.Timer
	t.EndTime = endTime
	return t, nil
}

// check timer can be stopped or paused
func (uc *UseCase) checkCountDownTimer(ctx context.Context, timerId uuid.UUID, userId int64) (*timermodel.CountdownTimer, error) {
	timer, err := uc.updater.CountdownTimer(ctx, timerId)
	if err != nil {
		return nil, exception.Wrap(err, exception.NewCause("get timer by id", "checkAccess", _PROVIDER))
	}
	if timer.Creator != userId {
		return nil, exception.Wrap(timererror.ExceptionUserForbidden(), exception.NewCause("check creator", "checkAccess", _PROVIDER))
	}
	return timer, nil
}
