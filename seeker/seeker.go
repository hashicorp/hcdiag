package seeker

import (
	"fmt"
	"path/filepath"
)

// Status describes the result of a seeker run
type Status string

const (
	// Success means all systems green
	Success Status = "success"
	// Fail means that we detected a known error and can conclusively say that the seeker did not complete.
	Fail Status = "fail"
	// Unknown means that we detected an error and the result is indeterminate (e.g. some side effect like disk or
	//   network may or may not have completed) or we don't recognize the error. If we don't recognize the error that's
	//   a signal to improve the error handling to account for it.
	Unknown Status = "unknown"
)

// Seeker seeks information via its Runner then stores the results.
type Seeker struct {
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

// Run calls a Runner's Run() method and writes the results and any errors on the seeker struct
func (s *Seeker) Run() (interface{}, error) {
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

// StatusCounts takes a slice of seeker references and returns a map containing sums of each Status
func StatusCounts(seekers []*Seeker) (map[Status]int, error) {
	statuses := make(map[Status]int)
	for _, s := range seekers {
		if s.Status == "" {
			return nil, fmt.Errorf("unable to build Statuses map, seeker not run: seeker=%s", s.Identifier)
		}
		statuses[s.Status]++
	}
	return statuses, nil
}
