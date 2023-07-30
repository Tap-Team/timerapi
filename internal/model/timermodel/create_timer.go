package timermodel

import (
	"time"

	"github.com/Tap-Team/timerapi/internal/errorutils/timererror"
	"github.com/Tap-Team/timerapi/internal/model/timermodel/timerfields"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/Tap-Team/timerapi/pkg/validate"
	"github.com/google/uuid"
)

const MIN_TIMER_DURATION = 1

type CreateTimer struct {
	ID          uuid.UUID               `json:"id"`
	UTC         int16                   `json:"utc"`
	StartTime   amidtime.DateTime       `json:"startTime"`
	EndTime     amidtime.DateTime       `json:"endTime"`
	Type        timerfields.Type        `json:"type"`
	Name        timerfields.Name        `json:"name"`
	Description timerfields.Description `json:"description"`
	Color       timerfields.Color       `json:"color"`
	WithMusic   bool                    `json:"withMusic"`
}

func NewCreateTimer(
	id uuid.UUID,
	utc int16,
	startTime, endTime amidtime.DateTime,
	timerType timerfields.Type,
	name timerfields.Name,
	description timerfields.Description,
	color timerfields.Color,
	withMusic bool,
) *CreateTimer {
	return &CreateTimer{
		ID:          id,
		UTC:         utc,
		StartTime:   startTime,
		EndTime:     endTime,
		Type:        timerType,
		Name:        name,
		Description: description,
		Color:       color,
		WithMusic:   withMusic,
	}
}

func (t *CreateTimer) ValidatableVariables() []validate.Validatable {
	return []validate.Validatable{t.Description, t.Name, t.Color, t.Type}
}

func (t *CreateTimer) Validate() error {
	err := validate.ValidateFields(t)
	if err != nil {
		return err
	}
	if t.StartTime.T().After(t.EndTime.T()) {
		return timererror.ExceptionWrongTimerTime()
	}
	if t.EndTime.Unix()-time.Now().Unix() < MIN_TIMER_DURATION {
		return timererror.ExceptionWrongTimerTime()
	}
	if t.ID == uuid.Nil {
		return timererror.ExceptionNilID()
	}
	return nil
}

func (t CreateTimer) DefaultDuration() int64 {
	return t.EndTime.Unix() - t.StartTime.Unix()
}
