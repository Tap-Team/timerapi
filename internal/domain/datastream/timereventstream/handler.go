package timereventstream

import (
	"sync"

	"github.com/Tap-Team/timerapi/internal/model/timerevent"
	"github.com/google/uuid"
)

// storage for EventStream, key is EventStream id (uuid), value is EventStream pointer (pointer)
type streamStorage struct {
	*sync.RWMutex
	storage map[uuid.UUID]*EventStream
}

// storage for Timer Subscribers, key is timer id (uuid), value is map of EventStream id (uuid) to empty struct
type subscribersStorage struct {
	*sync.RWMutex
	storage map[uuid.UUID]map[uuid.UUID]struct{}
}

type EventHandler struct {
	// sync map of EventStream uuid to EventStream
	streamStorage *streamStorage
	// sync map of timerId to subscribers EventStream ids
	timerSubscribers *subscribersStorage
}

func New() *EventHandler {
	return &EventHandler{
		streamStorage: &streamStorage{
			RWMutex: new(sync.RWMutex),
			storage: make(map[uuid.UUID]*EventStream),
		},
		timerSubscribers: &subscribersStorage{
			RWMutex: new(sync.RWMutex),
			storage: make(map[uuid.UUID]map[uuid.UUID]struct{}),
		},
	}
}

func (h *EventHandler) Send(event timerevent.TimerEvent) {
	// get event stream which subscribe on timer
	subscribers := make([]uuid.UUID, 0)
	h.timerSubscribers.RLock()
	for esId := range h.timerSubscribers.storage[event.TimerId()] {
		subscribers = append(subscribers, esId)
	}
	h.timerSubscribers.RUnlock()

	// send event to subscribers
	h.streamStorage.RLock()
	for _, esId := range subscribers {
		// get event stream ch timer
		es, ok := h.streamStorage.storage[esId]
		if ok {
			es.stream <- event
		}
	}
	h.streamStorage.RUnlock()
}
