package timerevent

import (
	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/google/uuid"
)

type EventType string

const (
	Update      EventType = "event_update"
	Stop        EventType = "event_stop"
	Start       EventType = "event_start"
	Subscribe   EventType = "event_subscribe"
	Unsubscribe EventType = "event_unsubscribe"
	Reset       EventType = "event_reset"
)

type TimerEvent interface {
	Type() EventType
	TimerId() uuid.UUID
}

type Event struct {
	Etype EventType `json:"type"`
	Id    uuid.UUID `json:"timerId"`
}

func (t *Event) TimerId() uuid.UUID {
	return t.Id
}

func (t *Event) Type() EventType {
	return t.Etype
}

type StopEvent struct {
	Event
	PauseTime amidtime.DateTime `json:"pauseTime"`
}

func NewStop(timerId uuid.UUID, pauseTime amidtime.DateTime) TimerEvent {
	return &StopEvent{Event: Event{Etype: Stop, Id: timerId}, PauseTime: pauseTime}
}

type StartEvent struct {
	Event
	EndTime amidtime.DateTime `json:"endTime"`
}

func NewStart(
	timerId uuid.UUID,
	endTime amidtime.DateTime,
) TimerEvent {
	return &StartEvent{
		Event:   Event{Etype: Start, Id: timerId},
		EndTime: endTime,
	}
}

type ResetEvent struct {
	Event
	EndTime amidtime.DateTime `json:"endTime"`
}

func NewReset(timerId uuid.UUID, endTime amidtime.DateTime) TimerEvent {
	return &ResetEvent{Event: Event{Etype: Reset, Id: timerId}, EndTime: endTime}
}

type UpdateEvent struct {
	Event
	timermodel.TimerSettings
}

func NewUpdate(timerId uuid.UUID, settings timermodel.TimerSettings) TimerEvent {
	return &UpdateEvent{
		Event{
			Etype: Update,
			Id:    timerId,
		},
		settings,
	}
}

// event which send client to server
// add or remove timer from hot update
type SubscribeEvent struct {
	Type     EventType   `json:"type"`
	TimerIds []uuid.UUID `json:"timerIds"`
}

func NewSubscribe(timerIds ...uuid.UUID) *SubscribeEvent {
	return &SubscribeEvent{Type: Subscribe, TimerIds: timerIds}
}

func NewUnsubscribe(timerIds ...uuid.UUID) *SubscribeEvent {
	return &SubscribeEvent{Type: Unsubscribe, TimerIds: timerIds}
}
