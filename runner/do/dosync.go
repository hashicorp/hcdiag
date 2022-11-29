// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package do

import (
	"fmt"
	"time"

	"github.com/hashicorp/hcdiag/runner"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
)

var _ runner.Runner = Sync{}

// Sync wraps a collection of runners and executes them in order, returning all of their Ops keyed by their ID(). If one
// of the runners has a status other than Success, subsequent runners will not be executed and the do-sync will return
// a status.Fail.
type Sync struct {
	Runners     []runner.Runner `json:"runners"`
	Label       string          `json:"label"`
	Description string          `json:"description"`
	log         hclog.Logger
}

// NewSync initializes a DoSync runner.
func NewSync(l hclog.Logger, label, description string, runners []runner.Runner) *Sync {
	return &Sync{
		Label:       label,
		Description: description,
		Runners:     runners,
		log:         l,
	}
}

func (d Sync) ID() string {
	return "do-sync " + d.Label
}

// Run executes the Command
func (d Sync) Run() op.Op {
	startTime := time.Now()
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
			}, runner.Params(d), startTime, time.Now())
		}
	}
	// Return runner ops, adding one at the end for our successful DoSync run
	return op.New(d.ID(), results, op.Success, nil, runner.Params(d), startTime, time.Now())
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
