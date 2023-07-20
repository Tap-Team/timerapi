package timereventstream

import (
	"sync"

	"github.com/Tap-Team/timerapi/internal/model/timerevent"
	"github.com/google/uuid"
)

type Stream interface {
	Subscribe(...uuid.UUID)
	Unsubscribe(...uuid.UUID)
	Stream() <-chan timerevent.TimerEvent
	Close()
}

type timers struct {
	*sync.RWMutex
	storage map[uuid.UUID]struct{}
}

type EventStream struct {
	id      uuid.UUID
	handler *EventHandler
	stream  chan timerevent.TimerEvent
	timers  *timers
}

func (es *EventStream) Subscribe(timerIds ...uuid.UUID) {
	es.handler.timerSubscribers.Lock()
	es.timers.Lock()
	var ok bool
	var m map[uuid.UUID]struct{}
	for _, timerId := range timerIds {
		// get subscribers of timer
		if m, ok = es.handler.timerSubscribers.storage[timerId]; !ok {
			// if not found, create new for timer
			m = make(map[uuid.UUID]struct{})
			es.handler.timerSubscribers.storage[timerId] = m
		}
		// set subscriber
		m[es.id] = struct{}{}
		es.timers.storage[timerId] = struct{}{}
	}
	es.timers.Unlock()
	es.handler.timerSubscribers.Unlock()
}

func (es *EventStream) Unsubscribe(timerIds ...uuid.UUID) {
	es.handler.timerSubscribers.Lock()
	es.timers.Lock()
	for _, timerId := range timerIds {
		// on each timer id delete es.id from timerSubscribers
		if m, ok := es.handler.timerSubscribers.storage[timerId]; ok {
			delete(m, es.id)
		}
		// delete timer from es.timers
		delete(es.timers.storage, timerId)
	}
	es.timers.Unlock()
	es.handler.timerSubscribers.Unlock()
}

func (es *EventStream) Stream() <-chan timerevent.TimerEvent {
	return es.stream
}

func (es *EventStream) Close() {
	// delete EventStream from stream storage
	es.handler.streamStorage.Lock()
	delete(es.handler.streamStorage.storage, es.id)
	es.handler.streamStorage.Unlock()

	// delete EventStream from every timer
	es.timers.RLock()
	es.handler.timerSubscribers.Lock()
	for timerId := range es.timers.storage {
		delete(es.handler.timerSubscribers.storage[timerId], es.id)
		delete(es.timers.storage, timerId)
	}
	es.handler.timerSubscribers.Unlock()
	es.timers.RUnlock()
	close(es.stream)
}

func (h *EventHandler) NewStream() interface {
	Subscribe(...uuid.UUID)
	Unsubscribe(...uuid.UUID)
	Stream() <-chan timerevent.TimerEvent
	Close()
} {
	stream := make(chan timerevent.TimerEvent, 20)
	es := &EventStream{
		id:      uuid.New(),
		handler: h,
		stream:  stream,
		timers: &timers{
			RWMutex: new(sync.RWMutex),
			storage: make(map[uuid.UUID]struct{}),
		},
	}
	h.streamStorage.Lock()
	h.streamStorage.storage[es.id] = es
	h.streamStorage.Unlock()
	return es
}
