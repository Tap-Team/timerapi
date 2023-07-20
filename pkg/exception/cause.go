package exception

import "fmt"

type Cause interface {
	error
	Action() string
	Method() string
	Pkg() string
}

type AmidCause struct {
	action string
	method string
	pkg    string
}

func NewCause(action string, method string, pkg string) Cause {
	return &AmidCause{action: action, method: method, pkg: pkg}
}

func (a *AmidCause) Error() string {
	return fmt.Sprintf("Action %s of Method %s in %s package", a.action, a.method, a.pkg)
}

func (a *AmidCause) Action() string {
	return a.action
}

func (e *AmidCause) Method() string {
	return e.method
}

func (e *AmidCause) Pkg() string {
	return e.pkg
}

func (e *AmidCause) Is(target error) bool {
	err, ok := target.(Cause)
	if !ok {
		return false
	}
	if e.action != err.Action() {
		return false
	}
	if e.method != err.Method() {
		return false
	}
	if e.pkg != err.Pkg() {
		return false
	}
	return true
}
