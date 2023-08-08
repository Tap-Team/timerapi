package botnotification

import (
	"context"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/Tap-Team/timerapi/internal/model/notification"
)

type MessageSender interface {
	MessagesSend(params api.Params) (response int, err error)
}

type NotificationStream interface {
	NewStream() interface {
		Stream() <-chan notification.NotificationSubscribers
		Close()
	}
}

type NotificationBot interface {
	Run(context.Context)
}

type notificationBot struct {
	sender             MessageSender
	notificationStream NotificationStream
}

func New(sender MessageSender, notificationStream NotificationStream) NotificationBot {
	return &notificationBot{sender: sender, notificationStream: notificationStream}
}

func (b *notificationBot) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stream := b.notificationStream.NewStream()
	defer stream.Close()
Loop:
	for {
		select {
		case <-ctx.Done():
			break Loop
		case n, ok := <-stream.Stream():
			if !ok {
				break Loop
			}
			go sendNotification(ctx, b.sender, n)
		}
	}
}

func sendNotification(ctx context.Context, sender MessageSender, n notification.NotificationSubscribers) {
	for _, userId := range n.Subscribers() {
		user := User(userId)
		user.Send(ctx, sender, n)
	}
}
