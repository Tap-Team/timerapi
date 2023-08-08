package bot

import (
	"context"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/Tap-Team/timerapi/internal/transport/bot/botnotification"
	"github.com/Tap-Team/timerapi/internal/transport/bot/messagehandlers"
)

type manager struct {
	vk *api.VK
}

type Manager interface {
	RunMessageHandlers()
	RunNotificationBot(ctx context.Context, nstream botnotification.NotificationStream)
}

func NewManager(vk *api.VK) Manager {
	return &manager{vk: vk}
}

// blocking function, if you not need blocking of code run in new goroutine: go Manager.RunNotificationBot
func (m *manager) RunNotificationBot(ctx context.Context, nstream botnotification.NotificationStream) {
	nbot := botnotification.New(m.vk, nstream)
	nbot.Run(ctx)
}

// blocking function, if you not need blocking of code run in new goroutine: go Manager.RunMessageHandlers
func (m *manager) RunMessageHandlers() {
	handler := messagehandlers.NewMain(m.vk)
	handler.Handle()
}
