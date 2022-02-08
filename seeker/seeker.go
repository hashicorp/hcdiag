package seeker

import (
	"fmt"
	"path/filepath"
)

type Status string

const (
	Success Status = "success"
	Fail    Status = "fail"
	Unknown Status = "unknown"
)

// Seeker seeks information via its Runner then stores the results.
type Seeker struct {
	Runner     Runner      `json:"runner"`
	Identifier string      `json:"-"`
	Result     interface{} `json:"result"`
	Status     Status      `json:"status"`
	ErrString  string      `json:"error"` // this simplifies json marshaling
	Error      error       `json:"-"`

	callbacks []func() error
}

// Runner runs things to get information.
type Runner interface {
	Run() (interface{}, error)
}

func (s *Seeker) Run() (interface{}, error) {
	s.Status = Unknown

	s.Result, s.Error = s.Runner.Run()
	if s.setStatus() == Fail {
		return s.Result, s.Error
	}

	// callbacks may include result validation or other post-Run() steps.
	for _, callback := range s.callbacks {
		s.Error = callback()
		if s.setStatus() == Fail {
			break // stop at first failure
		}
	}

	return s.Result, s.Error
}

func (s *Seeker) setStatus() Status {
	if s.Error == nil {
		s.Status = Success
	} else {
		s.Status = Fail
		s.ErrString = s.Error.Error()
	}
	return s.Status
}

func (s *Seeker) AddCallback(c func() error) {
	s.callbacks = append(s.callbacks, c)
}

// Exclude takes a slice of matcher strings and a slice of seekers. If any of the seeker identifiers match the exclude
// according to filepath.Match() then it will not be present in the returned seeker slice.
// NOTE(mkcp): This is precisely identical to Select() except we flip the match check. Maybe we can perform both rounds
//  of filtering in one pass one rather than iterating over all the seekers several times. Not likely to be a huge speed
//  increase though... we're not even remotely bottlenecked on seeker filtering.
func Exclude(excludes []string, seekers []*Seeker) ([]*Seeker, error) {
	newSeekers := make([]*Seeker, 0)
	for _, s := range seekers {
		// Set our match flag if we get a hit for any of the matchers on this seeker
		var match bool
		var err error
		for _, matcher := range excludes {
			match, err = filepath.Match(matcher, s.Identifier)
			if err != nil {
				return newSeekers, fmt.Errorf("filter error: '%s' for '%s'", err, matcher)
			}
			if match {
				break
			}
		}

		// Add the seeker back to our set if we have not matched an exclude
		if !match {
			newSeekers = append(newSeekers, s)
		}
	}
	return newSeekers, nil
}

// Select takes a slice of matcher strings and a slice of seekers. The only seekers returned will be those
// matching the given select strings according to filepath.Match()
func Select(selects []string, seekers []*Seeker) ([]*Seeker, error) {
	newSeekers := make([]*Seeker, 0)
	for _, s := range seekers {
		// Set our match flag if we get a hit for any of the matchers on this seeker
		var match bool
		var err error
		for _, matcher := range selects {
			match, err = filepath.Match(matcher, s.Identifier)
			if err != nil {
				return newSeekers, fmt.Errorf("filter error: '%s' for '%s'", err, matcher)
			}
			if match {
				break
			}
		}

		// Only include the seeker if we've matched it
		if match {
			newSeekers = append(newSeekers, s)
		}
	}
	return newSeekers, nil
}
