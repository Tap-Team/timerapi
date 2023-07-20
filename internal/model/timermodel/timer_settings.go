package timermodel

import (
	"github.com/Tap-Team/timerapi/internal/model/timermodel/timerfields"
	"github.com/Tap-Team/timerapi/pkg/validate"
)

type TimerSettings struct {
	Name        timerfields.Name        `json:"name"`
	Description timerfields.Description `json:"description"`
	Color       timerfields.Color       `json:"color"`
	WithMusic   bool                    `json:"withMusic"`
}

func NewTimerSettings(
	name timerfields.Name,
	description timerfields.Description,
	color timerfields.Color,
	withMusic bool,
) *TimerSettings {
	return &TimerSettings{
		Name:        name,
		Description: description,
		Color:       color,
		WithMusic:   withMusic,
	}
}

func (t *TimerSettings) ValidatableVariables() []validate.Validatable {
	return []validate.Validatable{t.Color, t.Name, t.Description}
}

func (t *TimerSettings) Validate() error {
	return validate.ValidateFields(t)
}
