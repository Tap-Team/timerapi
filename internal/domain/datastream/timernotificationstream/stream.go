package timernotificationstream

import (
	"github.com/Tap-Team/timerapi/internal/model/notification"
	"github.com/google/uuid"
)

type ServiceStream struct {
	// pointer to handler
	handler *StreamHandler
	// unique stream id
	id uuid.UUID
	// chan to stream notification
	ch chan notification.NotificationSubscribers
}

func (h *StreamHandler) NewStream() interface {
	Stream() <-chan notification.NotificationSubscribers
	Close()
} {
	id := uuid.New()
	stream := &ServiceStream{handler: h, id: id, ch: make(chan notification.NotificationSubscribers, 100)}

	h.mu.Lock()
	h.serviceStreams[id] = stream
	h.mu.Unlock()
	return stream
}

func (s *ServiceStream) Close() {
	s.handler.mu.Lock()
	delete(s.handler.serviceStreams, s.id)
	s.handler.mu.Unlock()
}

func (s *ServiceStream) Stream() <-chan notification.NotificationSubscribers {
	return s.ch
}
