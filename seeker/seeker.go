package seeker

import (
	"fmt"
	"path/filepath"
)

// Seeker seeks information via its Runner then stores the results.
type Seeker struct {
	Runner     Runner      `json:"runner"`
	Identifier string      `json:"-"`
	Result     interface{} `json:"result"`
	ErrString  string      `json:"error"` // this simplifies json marshaling
	Error      error       `json:"-"`
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

// Filter returns a subset of seekers that either do ("select") or do not ("exclude")
// match any of a list of matchers.  method must be either "select" or "exclude"
// and matchers will be run through filepath.Match() with each seeker's identifier.
// If method="select" then *only* seekers that *do* match will be returned,
// if method="exclude" then seekers that *do not* match will *not* be returned.
func Filter(method string, matchers []string, seekers []*Seeker) ([]*Seeker, error) {
	newSeekers := make([]*Seeker, 0)
	if method != "select" && method != "exclude" {
		return newSeekers, fmt.Errorf("filter error: method must be either 'select' or 'exclude', got: '%s'", method)
	}
	for _, s := range seekers {
		// Set our match flag if we get a hit for any of the matchers on this seeker
		var match bool
		var err error
		for _, matcher := range matchers {
			match, err = filepath.Match(matcher, s.Identifier)
			if err != nil {
				return newSeekers, fmt.Errorf("filter error: '%s' for '%s'", err, matcher)
			}
			if match {
				// Found a match, stop this inner loop
				break
			}
		}
		if (match && method == "select") || (!match && method == "exclude") {
			newSeekers = append(newSeekers, s)
		}
	}
	return newSeekers, nil
}
