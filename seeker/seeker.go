package seeker

import (
	"fmt"
	"regexp"
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

// Exclude takes a slice of matcher strings and a slice of seekers. If any of the seeker identifiers match the exclude
// exactly then it will not be present in the returned seeker slice.
// NOTE(mkcp): This is precisely identical to Select() except we flip the match check. Maybe we can perform both rounds
//  of filtering in one pass one rather than iterating over all the seekers several times. Not likely to be a huge speed
//  increase though... we're not even remotely bottlenecked on seeker filtering.
func Exclude(excludes []string, seekers []*Seeker) []*Seeker {
	newSeekers := make([]*Seeker, 0)
	for _, s := range seekers {
		// Set our match flag if we get a hit for any of the matchers on this seeker
		var match bool
		for _, matcher := range excludes {
			// TODO(gulducat): capture and log possible err here?
			if m, _ := regexp.MatchString(matcher, s.Identifier); m {
				match = true
				break
			}
		}

		// Add the seeker back to our set if we have not matched an exclude
		if !match {
			newSeekers = append(newSeekers, s)
		}
	}
	return newSeekers
}

// Select takes a slice of matcher strings and a slice of seekers. The only seekers returned will be those exactly
// matching the given select strings.
func Select(selects []string, seekers []*Seeker) []*Seeker {
	newSeekers := make([]*Seeker, 0)
	for _, s := range seekers {
		// Set our match flag if we get a hit for any of the matchers on this seeker
		var match bool
		for _, matcher := range selects {
			// TODO(gulducat): capture and log possible err here?
			if m, _ := regexp.MatchString(matcher, s.Identifier); m {
				match = true
				break
			}
		}

		// Only include the seeker if we've matched it
		if match {
			newSeekers = append(newSeekers, s)
		}
	}
	return newSeekers
}
