package timermodel

import (
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/google/uuid"
)

type TimerSubscribers struct {
	ID          uuid.UUID
	EndTime     amidtime.DateTime
	Subscribers []int64
}

func NewTimerSubscribers(id uuid.UUID, endTime amidtime.DateTime, subscribers []int64) *TimerSubscribers {
	return &TimerSubscribers{
		ID:          id,
		EndTime:     endTime,
		Subscribers: subscribers,
	}
}
