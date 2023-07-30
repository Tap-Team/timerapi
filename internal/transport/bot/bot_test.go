package bot_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/Tap-Team/timerapi/internal/model/notification"
	"github.com/Tap-Team/timerapi/internal/testdatamodule"
	"github.com/Tap-Team/timerapi/internal/transport/bot"
	"github.com/golang/mock/gomock"
)

type FakeNotificationStream chan notification.NotificationSubscribers

func (fn FakeNotificationStream) NewStream() interface {
	Stream() <-chan notification.NotificationSubscribers
	Close()
} {
	return stream{fn: fn}
}

type stream struct {
	fn FakeNotificationStream
}

func (s stream) Close() {}
func (s stream) Stream() <-chan notification.NotificationSubscribers {
	return s.fn
}

func RandomNotificationSubscribers(n notification.Notification) notification.NotificationSubscribers {
	l := rand.Intn(100)
	subscribers := make([]int64, 0, l)
	for i := 0; i < l; i++ {
		subscribers = append(subscribers, rand.Int63())
	}
	return notification.NewWithSubscribers(n, subscribers)
}

func TestSend(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ctrl := gomock.NewController(t)

	sender := bot.NewMockMessageSender(ctrl)
	fakeStream := make(FakeNotificationStream, 1)

	bot := bot.New(sender, fakeStream)
	timer := *testdatamodule.RandomTimer()
	var ntion notification.Notification
	if rand.Int63()%2 == 0 {
		ntion = notification.NewDelete(timer)
	} else {
		ntion = notification.NewExpired(timer)
	}
	timerSubs := RandomNotificationSubscribers(ntion)

	sender.EXPECT().MessagesSend(gomock.Any()).Return(0, nil).Times(len(timerSubs.Subscribers()))

	fakeStream <- timerSubs

	bot.Run(ctx)
}
