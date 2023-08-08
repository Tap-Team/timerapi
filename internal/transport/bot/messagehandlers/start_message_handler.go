package messagehandlers

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
	"github.com/SevereCloud/vksdk/v2/events"
)

const (
	startCommand = "start"
	startMessage = `Здравствуйте, приветствуем вас в нашем мини приложении, я бот который будет следить за вашими таймерами и уведомлять в случае если он будет удалён или окончит свою работу`
)

type startMessagePayload struct {
	Command string `json:"command"`
}

func sendStartMessage(ctx context.Context, vk *api.VK, peerId int) error {
	b := params.NewMessagesSendBuilder()
	b.Message(startMessage)
	b.PeerID(peerId)
	b.RandomID(rand.Int())
	_, err := vk.MessagesSend(b.Params)
	return err
}

func (m *mainHandler) startMessageHandler(ctx context.Context, obj events.MessageNewObject) {
	var payload startMessagePayload
	err := json.Unmarshal([]byte(obj.Message.Payload), &payload)
	if err != nil {
		log.Printf("failed unmarshal start message payload, %s", obj.Message.Payload)
		return
	}
	if payload.Command != startCommand {
		return
	}
	err = sendStartMessage(ctx, m.vk, obj.Message.PeerID)
	if err != nil {
		log.Printf("failed send start message, %s", err)
		return
	}
}
