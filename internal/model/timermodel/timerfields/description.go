package timerfields

import (
	"database/sql/driver"

	"github.com/Tap-Team/timerapi/pkg/amidstr"
	"github.com/Tap-Team/timerapi/pkg/validate"
)

const (
	DescriptionField   = "Описание Таймера"
	DescriptionMinSize = 0
	DescriptionMaxSize = 1000
)

type Description string

func (d Description) Validate() error {
	return validate.StringValidate(string(d), DescriptionField, DescriptionMinSize, DescriptionMaxSize)
}

func (d *Description) Scan(src any) error {
	return amidstr.ScanNullString((*string)(d), src)
}

func (d *Description) UnmarshalJSON(data []byte) error {
	return amidstr.UnmarshalTrimString((*string)(d), data)
}

func (d Description) Value() (driver.Value, error) {
	if len(d) == 0 {
		return nil, nil
	}
	return string(d), nil
}
