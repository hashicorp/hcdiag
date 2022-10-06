package runner

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/hcdiag/op"
)

// Runner runs things to get information.
type Runner interface {
	ID() string
	Run() op.Op
}

// Exclude takes a slice of matcher strings and a slice of ops. If any of the runner identifiers match the exclude
// according to filepath.Match() then it will not be present in the returned runner slice.
// NOTE(mkcp): This is precisely identical to Select() except we flip the match check. Maybe we can perform both rounds
// of filtering in one pass one rather than iterating over all the ops several times. Not likely to be a huge speed
// increase though... we're not even remotely bottlenecked on runner filtering.
func Exclude(excludes []string, runners []Runner) ([]Runner, error) {
	newRunners := make([]Runner, 0)
	for _, r := range runners {
		// Set our match flag if we get a hit for any of the matchers on this runner
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

		// Add the runner back to our set if we have not matched an exclude
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
		// Set our match flag if we get a hit for any of the matchers on this runner
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

// Params takes a Runner and returns a map of its public fields
func Params(r Runner) map[string]interface{} {
	var inInterface map[string]interface{}
	inrec, err := json.Marshal(&r)
	if err != nil {
		hclog.L().Error("runner.Params failed to serialize params", "runner", r, "error", err)
	}
	_ = json.Unmarshal(inrec, &inInterface)
	return inInterface
}
