package messagehandlers

import (
	"log"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/longpoll-bot"
)

type mainHandler struct {
	vk *api.VK
}

func NewMain(vk *api.VK) *mainHandler {
	return &mainHandler{vk: vk}
}

func (m *mainHandler) Handle() {
	// get information about the group
	group, err := m.vk.GroupsGetByID(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Initializing Long Poll
	lp, err := longpoll.NewLongPoll(m.vk, group[0].ID)
	if err != nil {
		log.Fatal(err)
	}
	lp.MessageNew(m.startMessageHandler)

	lp.Run()
}
