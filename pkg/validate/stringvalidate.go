package validate

import (
	"fmt"
	"net/http"
	"strings"
)

type WrongLenError struct {
	name             string
	min, max, actual int
}

func (v *WrongLenError) Error() string {
	return fmt.Sprintf("wrong len of parameter %s min %d max %d actual %d", v.name, v.min, v.max, v.actual)
}
func (v *WrongLenError) HttpCode() int {
	return http.StatusBadRequest
}
func (v *WrongLenError) Code() string {
	return "wrong_len"
}
func (v *WrongLenError) Type() string {
	return "common"
}

// ${param} должна находится в диапазоне от ${min} до ${max}
func (v *WrongLenError) Replace(target string) string {
	replacer := strings.NewReplacer("${param}", v.name, "${min}", fmt.Sprint(v.min), "${max}", fmt.Sprint(v.max))
	return replacer.Replace(target)
}

func StringValidate(s string, name string, min int, max int) error {
	l := len([]rune(s))
	if l < min || l > max {
		return &WrongLenError{name, min, max, l}
	}
	return nil
}
