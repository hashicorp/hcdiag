package op

import (
	"fmt"
	"path/filepath"
)

// Status describes the result of a op run
type Status string

const (
	// Success means all systems green
	Success Status = "success"
	// Fail means that we detected a known error and can conclusively say that the op did not complete.
	Fail Status = "fail"
	// Unknown means that we detected an error and the result is indeterminate (e.g. some side effect like disk or
	//   network may or may not have completed) or we don't recognize the error. If we don't recognize the error that's
	//   a signal to improve the error handling to account for it.
	Unknown Status = "unknown"
)

// Op seeks information via its Runner then stores the results.
type Op struct {
	Runner     Runner      `json:"runner"`
	Identifier string      `json:"-"`
	Result     interface{} `json:"result"`
	ErrString  string      `json:"error"` // this simplifies json marshaling
	Error      error       `json:"-"`
	Status     Status      `json:"status"`
}

// Runner runs things to get information.
type Runner interface {
	Run() (interface{}, Status, error)
}

// Run calls a Runner's Run() method and writes the results and any errors on the op struct
func (s *Op) Run() (interface{}, error) {
	result, stat, err := s.Runner.Run()
	s.Result = result
	s.Error = err
	s.Status = stat

	if err != nil {
		s.ErrString = fmt.Sprintf("%s", err)
		return s.Result, s.Error
	}
	return result, err
}

// Exclude takes a slice of matcher strings and a slice of seekers. If any of the op identifiers match the exclude
// according to filepath.Match() then it will not be present in the returned op slice.
// NOTE(mkcp): This is precisely identical to Select() except we flip the match check. Maybe we can perform both rounds
//  of filtering in one pass one rather than iterating over all the seekers several times. Not likely to be a huge speed
//  increase though... we're not even remotely bottlenecked on op filtering.
func Exclude(excludes []string, seekers []*Op) ([]*Op, error) {
	newSeekers := make([]*Op, 0)
	for _, s := range seekers {
		// Set our match flag if we get a hit for any of the matchers on this op
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

		// Add the op back to our set if we have not matched an exclude
		if !match {
			newSeekers = append(newSeekers, s)
		}
	}
	return newSeekers, nil
}

// Select takes a slice of matcher strings and a slice of seekers. The only seekers returned will be those
// matching the given select strings according to filepath.Match()
func Select(selects []string, seekers []*Op) ([]*Op, error) {
	newSeekers := make([]*Op, 0)
	for _, s := range seekers {
		// Set our match flag if we get a hit for any of the matchers on this op
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

		// Only include the op if we've matched it
		if match {
			newSeekers = append(newSeekers, s)
		}
	}
	return newSeekers, nil
}

// StatusCounts takes a slice of op references and returns a map containing sums of each Status
func StatusCounts(seekers []*Op) (map[Status]int, error) {
	statuses := make(map[Status]int)
	for _, s := range seekers {
		if s.Status == "" {
			return nil, fmt.Errorf("unable to build Statuses map, op not run: op=%s", s.Identifier)
		}
		statuses[s.Status]++
	}
	return statuses, nil
}
