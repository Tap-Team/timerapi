package exception

import (
	"fmt"
	"net/http"
	"strings"
)

// HttpCode interface for send custom http code from server
type HttpError interface {
	HttpCode() int
}

// code interface for error classification
type CodeableError interface {
	Code() string
}

// typed error for error classification
type TypedError interface {
	Type() string
}

type Exception interface {
	error
	fmt.Stringer
	HttpError
	CodeableError
	TypedError
}

// main error struct of the project
// users dont see exceptions, transport layer map it to ErrorResponse
type AmidException struct {
	err      error
	causes   []Cause
	code     string
	etype    string
	httpCode int
}

func New(httpcode int, etype string, code string) Exception {
	return &AmidException{code: code, etype: etype, httpCode: httpcode}
}
func Error(err error, httpcode int, etype string, code string) Exception {
	return &AmidException{err: err, code: code, etype: etype, httpCode: httpcode}
}

// create new exception with internal status codes
func NewInternal(err error) Exception {
	return Error(err, http.StatusInternalServerError, "common", "internal")
}

type CodeTypedError interface {
	CodeableError
	TypedError
}

func MakeCode(e CodeTypedError) string {
	return e.Type() + "_" + e.Code()
}

// Stringer implementation,
// return info of cause: err, code, httpcode for rest transport and list of causes if not empty
func (e *AmidException) String() string {
	causes := new(strings.Builder)
	for i, cause := range e.causes {
		causes.WriteString(fmt.Sprintf("\n%d. %s", i+1, cause))
	}
	return fmt.Sprintf(
		"Err is %s\nCauses: %s\nCode %s, HttpCode %d",
		e.err,
		causes,
		MakeCode(e),
		e.httpCode,
	)
}

// error implementation
func (e *AmidException) Error() string {
	// if e.err == nil {
	// 	return fmt.Sprintf("Code %s, HttpCode %d", MakeCode(e), e.httpCode)
	// }
	return e.String()
}

// TypedError implementation
func (e *AmidException) Type() string {
	return e.etype
}

// CodeableError implementation
func (e *AmidException) Code() string {
	return e.code
}

// HttpError implementation
func (e *AmidException) HttpCode() int {
	return e.httpCode
}

// Compare all fields
func (e *AmidException) Is(target error) bool {
	err, ok := target.(Exception)
	if !ok {
		return false
	}
	if e.code != err.Code() {
		return false
	}
	if e.etype != err.Type() {
		return false
	}
	if e.httpCode != err.HttpCode() {
		return false
	}
	return true
}

func (e *AmidException) Unwrap() error {
	return e.err
}

func Wrap(err error, cause Cause) Exception {
	switch err := err.(type) {
	case *AmidException:
		err.causes = append(err.causes, cause)
		return err
	default:
		return Wrap(NewInternal(err), cause)
	}
}
