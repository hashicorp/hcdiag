package runner

import (
	"context"
	"time"
)

// Timeout is an embeddable struct that includes the structural elements required for an embedding runner
// to implement cancellation/timeouts. The embedding struct still needs to use this data to implement such
// functionality, but the intention is for this struct to help ensure that runners that support timeouts will use
// similar field names, which are reported consistently in result output.
type Timeout struct {
	// Context is the base context that the runner should use.
	Context context.Context `json:"-"`

	// Timeout is the maximum duration that the runner should run.
	Timeout time.Duration `json:"timeout"`
}
