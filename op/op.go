package op

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcdiag/util"
)

// Status describes the result of a Runner's Run
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
	Identifier string                 `json:"-"`
	Result     interface{}            `json:"result"`
	ErrString  string                 `json:"error"` // this simplifies json marshaling
	Error      error                  `json:"-"`
	Status     Status                 `json:"status"`
	Params     map[string]interface{} `json:"params"`
}

// New takes a runner its results, serializing it into an immutable Op struct.
func New(runner Runner, result interface{}, status Status, err error) Op {
	// We store the error directly to make JSON serialization easier
	var message string
	if err != nil {
		message = err.Error()
	}
	return Op{
		Identifier: runner.ID(),
		Result:     result,
		Error:      err,
		ErrString:  message,
		Status:     status,
		Params:     util.RunnerParams(runner),
	}
}

// Runner runs things to get information.
type Runner interface {
	ID() string
	Run() Op
}

// Exclude takes a slice of matcher strings and a slice of ops. If any of the op identifiers match the exclude
// according to filepath.Match() then it will not be present in the returned op slice.
// NOTE(mkcp): This is precisely identical to Select() except we flip the match check. Maybe we can perform both rounds
//  of filtering in one pass one rather than iterating over all the ops several times. Not likely to be a huge speed
//  increase though... we're not even remotely bottlenecked on op filtering.
func Exclude(excludes []string, runners []Runner) ([]Runner, error) {
	newRunners := make([]Runner, 0)
	for _, r := range runners {
		// Set our match flag if we get a hit for any of the matchers on this op
		var match bool
		var err error
		for _, matcher := range excludes {
			match, err = filepath.Match(matcher, r.ID())
			if err != nil {
				return newRunners, fmt.Errorf("filter error: '%s' for '%s'", err, matcher)
			}
			if match {
				break
			}
		}

		// Add the op back to our set if we have not matched an exclude
		if !match {
			newRunners = append(newRunners, r)
		}
	}
	return newRunners, nil
}

// Select takes a slice of matcher strings and a slice of ops. The only ops returned will be those
// matching the given select strings according to filepath.Match()
func Select(selects []string, runners []Runner) ([]Runner, error) {
	newRunners := make([]Runner, 0)
	for _, r := range runners {
		// Set our match flag if we get a hit for any of the matchers on this op
		var match bool
		var err error
		for _, matcher := range selects {
			match, err = filepath.Match(matcher, r.ID())
			if err != nil {
				return newRunners, fmt.Errorf("filter error: '%s' for '%s'", err, matcher)
			}
			if match {
				break
			}
		}

		// Only include the runner if we've matched it
		if match {
			newRunners = append(newRunners, r)
		}
	}
	return newRunners, nil
}

// StatusCounts takes a slice of op references and returns a map containing sums of each Status
func StatusCounts(ops map[string]Op) (map[Status]int, error) {
	statuses := make(map[Status]int)
	for _, op := range ops {
		if op.Status == "" {
			return nil, fmt.Errorf("unable to build Statuses map, op not run: op=%s", op.Identifier)
		}
		statuses[op.Status]++
	}
	return statuses, nil
}
