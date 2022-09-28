package do

import (
	"errors"

	"github.com/hashicorp/hcdiag/runner"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
)

var _ runner.Runner = Sync{}

// Sync runs shell commands.
type Sync struct {
	Runners []runner.Runner `json:"runners"`
	log     hclog.Logger
	// TODO(dcohen): should "Do/DoSync" accept redactions? My instinct is no.
	// Redactions []*redact.Redact `json:"redactions"`
}

// NewSync provides a runner for bin commands
func NewSync(l hclog.Logger, runners []runner.Runner) *Sync {
	return &Sync{
		Runners: runners,
		log:     l,
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
			// TODO(dcohen) is this what returning a status for Do/DoSync blocks should look like?
			e := errors.New("do-sync runner failed")
			results[o.Identifier] = o
			return op.New(d.ID(), results, op.Fail, e, runner.Params(d))
		}
	}
	// Return runner ops, adding one at the end for our successful DoSync run
	return op.New(d.ID(), results, op.Success, nil, runner.Params(d))
}
