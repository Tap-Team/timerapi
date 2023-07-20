package timernotificationstream

import (
	"context"
	"sync"
	"time"

	"github.com/Tap-Team/timerapi/internal/model/notification"
	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/internal/model/timermodel/timerfields"
	"github.com/Tap-Team/timerapi/internal/timerservice"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/Tap-Team/timerapi/pkg/exception"
	"github.com/google/uuid"
)

const _PROVIDER = "internal/domain/datastream/timerservicestream"

type TimerStorage interface {
	DeleteTimer(ctx context.Context, id uuid.UUID) error
	Timer(ctx context.Context, timerId uuid.UUID) (*timermodel.Timer, error)
	Unsubscribe(ctx context.Context, timerId uuid.UUID, userId int64) error
	UpdatePauseTime(ctx context.Context, timerId uuid.UUID, pauseTime amidtime.DateTime, isPaused bool) error
}

type SubscriberCacheStorage interface {
	DeleteTimer(ctx context.Context, id uuid.UUID) error
	TimerSubscribers(ctx context.Context, timerId uuid.UUID) (timermodel.Subscribers, error)
}

type NotificationStorage interface {
	InsertNotification(ctx context.Context, userId int64, notification notification.Notification) error
}

type StreamHandler struct {
	mu *sync.Mutex
	// map of user to stream
	subscribers    map[int64]map[uuid.UUID]*Stream
	externalStream chan notification.Notification

	timerservice        timerservice.TimerServiceClient
	timerStorage        TimerStorage
	subscriberStorage   SubscriberCacheStorage
	notificationStorage NotificationStorage
}

func New(
	timerservice timerservice.TimerServiceClient,
	timerStorage TimerStorage,
	subscriberStorage SubscriberCacheStorage,
	notificationStorage NotificationStorage,
) *StreamHandler {
	return &StreamHandler{
		timerservice:        timerservice,
		timerStorage:        timerStorage,
		subscriberStorage:   subscriberStorage,
		notificationStorage: notificationStorage,

		mu:             new(sync.Mutex),
		subscribers:    make(map[int64]map[uuid.UUID]*Stream),
		externalStream: make(chan notification.Notification, 1024),
	}
}

func (sh *StreamHandler) Send(notification notification.Notification) {
	sh.externalStream <- notification
}

func (sh *StreamHandler) Start(ctx context.Context) error {
	stream, err := sh.timerservice.TimerTick(ctx)
	if err != nil {
		return exception.Wrap(err, exception.NewCause("create service stream chan", "Start", _PROVIDER))
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case timerIds, ok := <-stream:
			if !ok {
				continue
			}
			for _, timerId := range timerIds {
				go sh.timerExpired(ctx, timerId)
			}
		case n, ok := <-sh.externalStream:
			if !ok {
				continue
			}
			switch n.Type() {
			case notification.Delete:
				go sh.timerDelete(ctx, n.Timer())
			}
		}
	}
}

// send notification for every subscriber
// if subscriber offline save notification in storage
func (sh *StreamHandler) notification(ctx context.Context, notification notification.Notification) {
	timerSubscribers, err := sh.subscriberStorage.TimerSubscribers(ctx, notification.TimerId())
	if err != nil {
		return
	}
	// in range send to every stream subscriber if of expired timer
	for userId := range timerSubscribers {
		// if subscriber online send timerId to stream
		if streamp, ok := sh.subscribers[userId]; ok {
			for _, stream := range streamp {
				stream.ch <- notification
			}
			// if subscriber offline, save notification in storage
		} else {
			sh.notificationStorage.InsertNotification(ctx, userId, notification)
		}
	}
}

func (sh *StreamHandler) timerDelete(ctx context.Context, timer timermodel.Timer) {
	// create context with timeout
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	// send notification for every subscriber
	sh.notification(ctx, notification.NewDelete(timer))
	// delete timer from storage
	sh.subscriberStorage.DeleteTimer(ctx, timer.ID)
}

func (sh *StreamHandler) timerExpired(ctx context.Context, timerId uuid.UUID) {
	// create context with timeout
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	timer, err := sh.timerStorage.Timer(ctx, timerId)
	if err != nil {
		return
	}

	// send notification for every subscriber
	sh.notification(ctx, notification.NewExpired(*timer))

	// clear timer from storage (delete or reset time)
	sh.clearExpiredTimer(ctx, *timer)
}

func (sh *StreamHandler) clearExpiredTimer(ctx context.Context, timer timermodel.Timer) {
	// depending on the type delete or clear timer
	switch timer.Type {
	case timerfields.DATE:
		wg := new(sync.WaitGroup)
		wg.Add(2)
		// delete timer from storage
		go func() { sh.timerStorage.DeleteTimer(ctx, timer.ID); wg.Done() }()
		// delete timer from subsriber storage with them subscribers
		go func() { sh.subscriberStorage.DeleteTimer(ctx, timer.ID); wg.Done() }()
		wg.Wait()
	case timerfields.COUNTDOWN:
		/*
				to reset timer we need 2 things
				1. Stop timer
				2. Set pause timer on timerStartTime
			Example:
			StartTime 00:00
			EndTime 02:00

			When timer expired we set Pause Time to 00:00

			Imagine in this day we start this timer in 12:00

			When we start stopped timer we set endTime = endTime + time.Since(timer.PauseTime)
			endTime = 02:00 + (12:00 - 00:00) = 14:00

			check timer duration not changed

			02:00 - 00:00 = 2
			endTime - launchTime
			14:00 - 12:00 = 2
			2 = 2
		*/
		pauseTime := amidtime.DateTime(time.Unix(timer.EndTime.Unix()-timer.Duration, 0))
		sh.timerStorage.UpdatePauseTime(ctx, timer.ID, pauseTime, true)
	}
}
