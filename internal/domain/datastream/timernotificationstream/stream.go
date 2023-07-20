package timernotificationstream

import (
	"github.com/Tap-Team/timerapi/internal/model/notification"
	"github.com/google/uuid"
)

type Stream struct {
	handler *StreamHandler
	id      uuid.UUID
	userId  int64
	ch      chan notification.Notification
}

func (sh *StreamHandler) NewStream(userId int64) interface {
	Stream() <-chan notification.Notification
	Close()
} {
	id := uuid.New()
	stream := &Stream{handler: sh, id: id, userId: userId, ch: make(chan notification.Notification, 10)}
	sh.mu.Lock()
	if _, ok := sh.subscribers[userId]; !ok {
		sh.subscribers[userId] = make(map[uuid.UUID]*Stream)
	}
	sh.subscribers[userId][id] = stream
	sh.mu.Unlock()
	return stream
}

func (s *Stream) Close() {
	s.handler.mu.Lock()
	delete(s.handler.subscribers[s.userId], s.id)
	if len(s.handler.subscribers[s.userId]) == 0 {
		delete(s.handler.subscribers, s.userId)
	}
	close(s.ch)
	s.handler.mu.Unlock()
}

func (s *Stream) Stream() <-chan notification.Notification {
	return s.ch
}
