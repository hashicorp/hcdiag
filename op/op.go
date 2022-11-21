package op

import (
	"fmt"
	"time"
)

// Status describes the result of an operation.
type Status string

const (
	// Success means all systems green
	Success Status = "success"

	// Fail means that we detected a known error and can conclusively say that the runner did not complete.
	Fail Status = "fail"

	// Unknown means that we detected an error and the result is indeterminate (e.g. some side effect like disk or
	// network may or may not have completed) or we don't recognize the error. If we don't recognize the error that's
	// a signal to improve the error handling to account for it.
	Unknown Status = "unknown"

	// Skip means that this Op was intentionally not run
	Skip Status = "skip"

	// Timeout means that an operation timed out during execution.
	Timeout Status = "timeout"

	// Canceled means that an operation was canceled during execution.
	Canceled Status = "canceled"
)

// Op seeks information via its Runner then stores the results.
type Op struct {
	Identifier string                 `json:"-"`
	Result     map[string]any         `json:"result"`
	ErrString  string                 `json:"error"` // this simplifies json marshaling
	Error      error                  `json:"-"`
	Status     Status                 `json:"status"`
	Params     map[string]interface{} `json:"params,omitempty"`
	Start      time.Time              `json:"start"`
	End        time.Time              `json:"end"`
}

// New takes a runner its results, serializing it into an immutable Op struct.
func New(id string, result map[string]any, status Status, err error, params map[string]any, start time.Time, end time.Time) Op {
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
		Start:      start,
		End:        end,
	}
}

func NewCancel(id string, err error, params map[string]any, start time.Time) Op {
	return New(id, map[string]any{}, Canceled, err, params, start, time.Now())
}

func NewTimeout(id string, err error, params map[string]any, start time.Time) Op {
	return New(id, map[string]any{}, Timeout, err, params, start, time.Now())
}

// StatusCounts takes a slice of op references and returns a map containing sums of each Status
func StatusCounts(ops map[string]Op) (map[Status]int, error) {
	// copy our input into a new map that conforms to our Walk input type
	m := make(map[string]any, len(ops))
	for k, v := range ops {
		m[k] = v
	}

	statuses := WalkStatuses(m)

	if val, ok := statuses[""]; ok {
		return nil, fmt.Errorf("op.StatusCounts received ops that did not have a status, count=%d", val)
	}
	return statuses, nil
}

// WalkStatuses performs a depth-first search of a tree of results and returns a map of status counts
func WalkStatuses(results map[string]any) map[Status]int {
	statuses := make(map[Status]int)

	var walkStatuses func(map[Status]int, map[string]any)
	walkStatuses = func(accumulator map[Status]int, results map[string]any) {
		for _, res := range results {
			switch res := res.(type) {

			case map[string]any:
				walkStatuses(statuses, res)

			case Op:
				// Increment and recur
				statuses[res.Status]++
				walkStatuses(statuses, res.Result)

			// End of branch reached
			default:
				continue
			}
		}
	}
	walkStatuses(statuses, results)
	return statuses
}
