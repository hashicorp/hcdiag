package seeker

import (
	"reflect"
	"runtime"
)

var _ Runner = GoFuncSeeker{}

// GoFunc is any function that returns an interface and error
type GoFunc func() (interface{}, error)

// GoFuncSeeker runs go functions.
type GoFuncSeeker struct {
	f    GoFunc
	Name string
}

// NewGoFuncSeeker provides a Seeker for running go functions.
func NewGoFuncSeeker(id string, function GoFunc) *Seeker {
	return &Seeker{
		Identifier: id,
		Runner: GoFuncSeeker{
			f: function,
			// this "get the function name" reflection is a mouthful, but it's here to
			// show up in Results.json ... might potentially be helpful?
			Name: runtime.FuncForPC(reflect.ValueOf(function).Pointer()).Name(),
		},
	}
}

// Run executes the function
func (gf GoFuncSeeker) Run() (interface{}, Status, error) {
	result, err := gf.f()
	if err != nil {
		return result, Fail, err
	}
	return result, Success, nil
}
