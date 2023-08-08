package botnotification

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"github.com/SevereCloud/vksdk/v2/api/params"
	"github.com/Tap-Team/timerapi/internal/model/notification"
)

type User int64

func (u User) Send(ctx context.Context, sender MessageSender, n notification.Notification) {
	msg, err := message(n)
	if err != nil {
		return
	}
	b := params.NewMessagesSendBuilder()
	b.WithContext(ctx)
	b.UserID(int(u))
	b.Message(msg)
	b.RandomID(rand.Int())
	sender.MessagesSend(b.Params)
}

func message(n notification.Notification) (string, error) {
	switch n.Type() {
	case notification.Delete:
		return deleteMessage(n), nil
	case notification.Expired:
		return expiredMessage(n), nil
	default:
		return "", errors.New("wrong notification type")
	}
}

func deleteMessage(n notification.Notification) string {
	name := n.Timer().Name
	if len(name) == 0 {
		name = "Без названия"
	}
	return fmt.Sprintf(`Здравствуйте, уведомляю вас о том что таймер %s был удалён`, name)
}

func expiredMessage(n notification.Notification) string {
	name := n.Timer().Name
	if len(name) == 0 {
		name = "Без названия"
	}
	return fmt.Sprintf(`Здравствуйте, уведомляю о том что таймер %s истёк`, name)
}
