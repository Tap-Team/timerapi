package timerfields

import "github.com/Tap-Team/timerapi/internal/errorutils/timererror"

type Type string

const (
	COUNTDOWN Type = "COUNTDOWN"
	DATE      Type = "DATE"
)

func (t Type) Validate() error {
	for _, tp := range []Type{COUNTDOWN, DATE} {
		if tp == t {
			return nil
		}
	}
	return timererror.ExceptionTypeNotFound()
}
