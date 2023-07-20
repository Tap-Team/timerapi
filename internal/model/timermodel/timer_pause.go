package timermodel

import (
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/google/uuid"
)

type TimerPause struct {
	ID        uuid.UUID         `json:"id"`
	IsPaused  bool              `json:"isPaused"`
	PauseTime amidtime.DateTime `json:"pauseTime"`
}

func NewTimerPause(
	id uuid.UUID,
	isPaused bool,
	pauseTime amidtime.DateTime,
) *TimerPause {
	return &TimerPause{ID: id, IsPaused: isPaused, PauseTime: pauseTime}
}

type CountdownTimer struct {
	Timer
	PauseTime amidtime.DateTime `json:"pauseTime"`
}

func NewCountDownTimer(timer Timer, isPaused bool, pauseTime amidtime.DateTime) *CountdownTimer {
	return &CountdownTimer{Timer: timer, PauseTime: pauseTime}
}
