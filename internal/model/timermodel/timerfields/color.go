package timerfields

import "github.com/Tap-Team/timerapi/internal/errorutils/timererror"

type Color string

const (
	DEFAULT Color = "DEFAULT"
	RED     Color = "RED"
	GREEN   Color = "GREEN"
	BLUE    Color = "BLUE"
	PURPLE  Color = "PURPLE"
	YELLOW  Color = "YELLOW"
)

func (c Color) Validate() error {
	for _, clr := range []Color{DEFAULT, RED, GREEN, BLUE, PURPLE, YELLOW} {
		if c == clr {
			return nil
		}
	}
	return timererror.ExceptionColorNotFound()
}
