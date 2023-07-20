package saga

import (
	"sync"
)

type Saga struct {
	mu    sync.Mutex
	queue []func()
	ok    bool
}

func (s *Saga) Rollback() {
	s.mu.Lock()
	if s.ok {
		return
	}
	for len(s.queue) != 0 {
		s.queue[len(s.queue)-1]()
		s.queue = s.queue[:len(s.queue)-1]
	}
	s.mu.Unlock()
}

func (s *Saga) OK() {
	s.mu.Lock()
	s.ok = true
	s.queue = []func(){}
	s.mu.Unlock()
}

func (s *Saga) Register(f func()) {
	s.mu.Lock()
	s.queue = append(s.queue, f)
	s.mu.Unlock()
}
