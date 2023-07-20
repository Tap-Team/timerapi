package timermodel

import (
	"time"

	"github.com/Tap-Team/timerapi/internal/model/timermodel/timerfields"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/Tap-Team/timerapi/pkg/validate"
	"github.com/google/uuid"
)

type Timer struct {
	ID          uuid.UUID               `json:"id"`
	UTC         int16                   `json:"utc"`
	Creator     int64                   `json:"creator"`
	EndTime     amidtime.DateTime       `json:"endTime"`
	PauseTime   amidtime.DateTime       `json:"pauseTime"`
	Type        timerfields.Type        `json:"type"`
	Name        timerfields.Name        `json:"name"`
	Description timerfields.Description `json:"description"`
	Color       timerfields.Color       `json:"color"`
	WithMusic   bool                    `json:"withMusic"`
	Duration    int64                   `json:"duration"`
	IsPaused    bool                    `json:"isPaused,omitempty"`
}

func NewTimer(
	id uuid.UUID,
	utc int16,
	creator int64,
	endTime, pauseTime amidtime.DateTime,
	timerType timerfields.Type,
	name timerfields.Name,
	description timerfields.Description,
	color timerfields.Color,
	withMusic bool,
	duration int64,
	isPaused bool,
) *Timer {
	return &Timer{
		ID:          id,
		UTC:         utc,
		Creator:     creator,
		EndTime:     endTime,
		PauseTime:   pauseTime,
		Type:        timerType,
		Name:        name,
		Description: description,
		Color:       color,
		WithMusic:   withMusic,
		Duration:    duration,
		IsPaused:    isPaused,
	}
}

func (t *Timer) ValidatableVariables() []validate.Validatable {
	return []validate.Validatable{t.Description, t.Name, t.Color, t.Type}
}

func (t *Timer) CreateTimer() *CreateTimer {
	startTime := time.Unix(t.EndTime.T().Unix()-t.Duration, 0)
	return NewCreateTimer(t.ID, t.UTC, amidtime.DateTime(startTime), t.EndTime, t.Type, t.Name, t.Description, t.Color, t.WithMusic)
}

func (t *Timer) Is(target *Timer) (string, bool) {
	if t.ID != target.ID {
		return "id", false
	}
	if t.UTC != target.UTC {
		return "utc", false
	}
	if t.Creator != target.Creator {
		return "creator", false
	}
	if t.EndTime.T().Unix() != target.EndTime.T().Unix() {
		return "endtime", false
	}
	if t.Type != target.Type {
		return "type", false
	}
	if t.Name != target.Name {
		return "name", false
	}
	if t.Description != target.Description {
		return "description", false
	}
	if t.Color != target.Color {
		return "color", false
	}
	if t.WithMusic != target.WithMusic {
		return "withmusic", false
	}
	if t.Duration != target.Duration {
		return "duration", false
	}
	if t.IsPaused != target.IsPaused {
		return "is_paused", false
	}
	return "", true
}
