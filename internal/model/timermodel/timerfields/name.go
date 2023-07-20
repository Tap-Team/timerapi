package timerfields

import (
	"database/sql/driver"

	"github.com/Tap-Team/timerapi/pkg/amidstr"
	"github.com/Tap-Team/timerapi/pkg/validate"
)

const (
	NameField   = "Название Таймера"
	NameMinSize = 0
	NameMaxSize = 60
)

type Name string

func (n Name) Validate() error {
	return validate.StringValidate(string(n), NameField, NameMinSize, NameMaxSize)
}

func (n *Name) UnmarshalJSON(data []byte) error {
	return amidstr.UnmarshalTrimString((*string)(n), data)
}
func (n *Name) Scan(src any) error {
	return amidstr.ScanNullString((*string)(n), src)
}

func (n Name) Value() (driver.Value, error) {
	if len(n) == 0 {
		return nil, nil
	}
	return string(n), nil
}
