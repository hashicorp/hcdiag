package seeker

import (
	"fmt"
)

// Runner runs things to get information.
type Runner interface {
	Run() (interface{}, error)
}

// Seeker seeks information via its Runner then stores the results.
type Seeker struct {
	Runner      Runner      `json:"runner"`
	Identifier  string      `json:"-"`
	Result      interface{} `json:"result"`
	ErrString   string      `json:"error"` // this simplifies json marshaling
	Error       error       `json:"-"`
}

func (s *Seeker) Run() (result interface{}, err error) {
	result, err = s.Runner.Run()
	s.Result = result
	s.Error = err

	if err != nil {
		s.ErrString = fmt.Sprintf("%s", err)
		return s.Result, s.Error
	}
	return result, err
}
