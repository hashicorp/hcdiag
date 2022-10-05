package op

import (
	"fmt"
)

// Status describes the result of an operation.
type Status string

const (
	// Success means all systems green
	Success Status = "success"

	// Fail means that we detected a known error and can conclusively say that the runner did not complete.
	Fail Status = "fail"

	// Unknown means that we detected an error and the result is indeterminate (e.g. some side effect like disk or
	//   network may or may not have completed) or we don't recognize the error. If we don't recognize the error that's
	//   a signal to improve the error handling to account for it.
	Unknown Status = "unknown"

	// Skip means that this Op was intentionally not run
	Skip Status = "skip"
)

// Op seeks information via its Runner then stores the results.
type Op struct {
	Identifier string                 `json:"-"`
	Result     map[string]any         `json:"result"`
	ErrString  string                 `json:"error"` // this simplifies json marshaling
	Error      error                  `json:"-"`
	Status     Status                 `json:"status"`
	Params     map[string]interface{} `json:"params,omitempty"`
}

// New takes a runner its results, serializing it into an immutable Op struct.
func New(id string, result map[string]any, status Status, err error, params map[string]any) Op {
	// We store the error directly to make JSON serialization easier
	var message string
	if err != nil {
		message = err.Error()
	}
	return Op{
		Identifier: id,
		Result:     result,
		Error:      err,
		ErrString:  message,
		Status:     status,
		Params:     params,
	}
}

// StatusCounts takes a slice of op references and returns a map containing sums of each Status
func StatusCounts(ops map[string]Op) (map[Status]int, error) {
	statuses := make(map[Status]int)
	for _, o := range ops {
		if o.Status == "" {
			return nil, fmt.Errorf("unable to build Statuses map, op not run: op=%s", o.Identifier)
		}
		statuses[o.Status]++
	}
	return statuses, nil
}
