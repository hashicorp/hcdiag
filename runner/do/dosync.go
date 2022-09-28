package do

import (
	"fmt"

	"github.com/hashicorp/hcdiag/runner"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
)

var _ runner.Runner = Sync{}

// Sync runs shell commands.
type Sync struct {
	Runners     []runner.Runner `json:"runners"`
	Label       string          `json:"label"`
	Description string          `json:"description"`
	log         hclog.Logger
}

// NewSync provides a runner for bin commands
func NewSync(l hclog.Logger, label, description string, runners []runner.Runner) *Sync {
	return &Sync{
		Label:       label,
		Description: description,
		Runners:     runners,
		log:         l,
	}
}

func (d Sync) ID() string {
	return "do-sync"
}

// Run executes the Command
func (d Sync) Run() op.Op {
	results := make(map[string]any, 0)

	for _, r := range d.Runners {
		d.log.Info("running operation", "runner", r.ID())
		o := r.Run()
		// If any result op is not Success, abort and return all existing ops
		if o.Status != op.Success {
			// Add an op for this failed DoSync at the end of the slice
			results[o.Identifier] = o
			return op.New(d.ID(), results, op.Fail, ChildRunnerError{
				Parent: d.ID(),
				Child:  o.Identifier,
				err:    o.Error,
			}, runner.Params(d))
		}
	}
	// Return runner ops, adding one at the end for our successful DoSync run
	return op.New(d.ID(), results, op.Success, nil, runner.Params(d))
}

type ChildRunnerError struct {
	Parent string
	Child  string
	err    error
}

func (e ChildRunnerError) Error() string {
	return fmt.Sprintf("error in child runner, parent=%s, child=%s, err=%s", e.Parent, e.Child, e.Unwrap().Error())
}

func (e ChildRunnerError) Unwrap() error {
	return e.err
}
