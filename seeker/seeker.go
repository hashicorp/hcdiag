package seeker

import (
	"fmt"
)

// Seeker seeks information via its Runner then stores the results.
type Seeker struct {
	Runner      Runner      `json:"runner"`
	Identifier  string      `json:"-"`
	Result      interface{} `json:"result"`
	ErrString   string      `json:"error"` // this simplifies json marshaling
	Error       error       `json:"-"`
}

// Runner runs things to get information.
type Runner interface {
	Run() (interface{}, error)
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

// CountInSets iterates over each set of seekers and sums their counts
func CountInSets(sets map[string][]*Seeker) int {
	var count int
	for _, s := range sets  {
		count = count + len(s)
	}
	return count
}
