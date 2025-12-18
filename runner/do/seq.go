// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package do

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/hcdiag/runner"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
)

var _ runner.Runner = Seq{}

type SeqConfig struct {
	Runners     []runner.Runner
	Label       string
	Description string
	Timeout     runner.Timeout
	Logger      hclog.Logger
}

// Seq wraps a collection of runners and executes them in order, returning all of their Ops keyed by their ID(). If one
// of the runners has a status other than Success, subsequent runners will not be executed and the Seq will return
// a status.Fail.
type Seq struct {
	Runners     []runner.Runner `json:"runners"`
	Label       string          `json:"label"`
	Description string          `json:"description"`
	Timeout     runner.Timeout  `json:"timeout"`
	log         hclog.Logger
	ctx         context.Context
}

// NewSeq initializes a Seq runner.
func NewSeq(cfg SeqConfig) *Seq {
	return NewSeqWithContext(context.Background(), cfg)
}

func NewSeqWithContext(ctx context.Context, cfg SeqConfig) *Seq {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Seq{
		ctx:         ctx,
		Label:       cfg.Label,
		Description: cfg.Description,
		Runners:     cfg.Runners,
		Timeout:     cfg.Timeout,
		log:         cfg.Logger,
	}
}

func (s Seq) ID() string {
	return "seq " + s.Label
}

// Run executes the Command
func (s Seq) Run() op.Op {
	startTime := time.Now()

	if s.ctx == nil {
		s.ctx = context.Background()
	}

	resChan := make(chan op.Op)

	runCtx := s.ctx
	var cancel context.CancelFunc
	if 0 < s.Timeout {
		runCtx, cancel = context.WithTimeout(s.ctx, time.Duration(s.Timeout))
		defer cancel()
	}

	go func(resChan chan<- op.Op, start time.Time) {
		o := s.run()
		o.Start = start
		resChan <- o
	}(resChan, startTime)

	select {
	case <-runCtx.Done():
		switch runCtx.Err() {
		case context.Canceled:
			return op.NewCancel(s.ID(), runCtx.Err(), runner.Params(s), startTime)
		case context.DeadlineExceeded:
			return op.NewTimeout(s.ID(), runCtx.Err(), runner.Params(s), startTime)
		default:
			return op.New(s.ID(), nil, op.Unknown, runCtx.Err(), runner.Params(s), startTime, time.Now())
		}
	case result := <-resChan:
		return result
	}
}

func (s Seq) run() op.Op {
	results := make(map[string]any, 0)
	for _, r := range s.Runners {
		s.log.Info("running operation", "runner", r.ID())
		o := r.Run()
		// If any result op is not Success, abort and return all existing ops
		if o.Status != op.Success {
			// Add an op for this failed Seq at the end of the slice
			results[o.Identifier] = o
			return op.New(s.ID(), results, op.Fail, ChildRunnerError{
				Parent: s.ID(),
				Child:  o.Identifier,
				err:    o.Error,
			}, runner.Params(s), time.Time{}, time.Now())
		}
	}
	// Return runner ops, adding one at the end for our successful Seq run
	return op.New(s.ID(), results, op.Success, nil, runner.Params(s), time.Time{}, time.Now())
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
