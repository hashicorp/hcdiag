package runner

import (
	"errors"

	"github.com/hashicorp/hcdiag/op"
)

var _ Runner = DoSync{}

// DoSync runs shell commands.
type DoSync struct {
	Runners []Runner `json:"runners"`
	// TODO(dcohen): should "Do/DoSync" accept redactions? My instinct is no.
	// Redactions []*redact.Redact `json:"redactions"`
}

// NewDoSync provides a runner for bin commands
func NewDoSync(runners []Runner) *DoSync {
	return &DoSync{
		Runners: runners,
	}
}

func (d DoSync) ID() string {
	return "DoSync"
}

// Run executes the Command
func (d DoSync) Run() []op.Op {
	opList := make([]op.Op, 0)

	for _, r := range d.Runners {
		ops := r.Run()
		// If any result op is not Success, abort and return all existing ops
		for _, o := range ops {
			if o.Status != op.Success {
				// Add an op for this failed DoSync at the end of the slice
				// TODO(dcohen) is this what returning a status for Do/DoSync blocks should look like?
				e := errors.New("DoSync runner failed")
				return append(opList, op.New(d.ID(), "TODO what is a failed DoSync result string?", op.Fail, e, Params(d)))
			}
			opList = append(opList, o)
		}
	}
	// Return runner ops, adding one at the end for our successful DoSync run
	return append(opList, op.New(d.ID(), "TODO what is a successful DoSync result string?", op.Success, nil, Params(d)))
}
