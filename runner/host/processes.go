// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package host

import (
	"context"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/runner"
	"github.com/mitchellh/go-ps"
)

var _ runner.Runner = &Process{}

type ProcessConfig struct {
	// Redactions includes any redactions to apply to the output of the runner.
	Redactions []*redact.Redact
	// Timeout specifies the amount of time that the runner should be allowed to execute before cancellation.
	Timeout time.Duration
}

// Process represents a single OS Process
type Process struct {
	ctx context.Context

	// Redactions includes any redactions to apply to the output of the runner.
	Redactions []*redact.Redact `json:"redactions"`
	// Timeout specifies the amount of time that the runner should be allowed to execute before cancellation.
	Timeout runner.Timeout `json:"timeout"`
}

func NewProcess(cfg ProcessConfig) *Process {
	return NewProcessWithContext(context.Background(), cfg)
}

func NewProcessWithContext(ctx context.Context, cfg ProcessConfig) *Process {
	return &Process{
		ctx:        ctx,
		Redactions: cfg.Redactions,
		Timeout:    runner.Timeout(cfg.Timeout),
	}
}

// Proc represents the process data we're collecting and returning
type Proc struct {
	Name string `json:"name"`
	PID  int    `json:"pid"`
	PPID int    `json:"ppid"`
}

func (p Process) ID() string {
	return "process"
}

func (p Process) Run() op.Op {
	startTime := time.Now()

	if p.ctx == nil {
		p.ctx = context.Background()
	}

	resChan := make(chan op.Op, 1)
	runCtx := p.ctx
	var cancel context.CancelFunc
	if 0 < p.Timeout {
		runCtx, cancel = context.WithTimeout(p.ctx, time.Duration(p.Timeout))
		defer cancel()
	}

	go func(resChan chan op.Op) {
		o := p.run()
		o.Start = startTime
		resChan <- o
	}(resChan)

	select {
	case <-runCtx.Done():
		switch runCtx.Err() {
		case context.Canceled:
			return runner.CancelOp(p, runCtx.Err(), startTime)
		case context.DeadlineExceeded:
			return runner.TimeoutOp(p, runCtx.Err(), startTime)
		default:
			return op.New(p.ID(), nil, op.Unknown, runCtx.Err(), runner.Params(p), startTime, time.Now())
		}
	case o := <-resChan:
		return o
	}
}

func (p Process) run() op.Op {
	var procs []Proc

	psProcs, err := ps.Processes()
	if err != nil {
		hclog.L().Trace("runner/host.Process.Run()", "error", err)
		results := map[string]any{"procs": psProcs}
		return op.New(p.ID(), results, op.Fail, err, runner.Params(p), time.Time{}, time.Now())
	}

	procs, err = p.procs(psProcs)
	results := map[string]any{
		"procs":      procs,
		"proc_count": len(procs),
	}
	if err != nil {
		hclog.L().Trace("runner/host.Process.Run()", "error", err)
		return op.New(p.ID(), results, op.Fail, err, runner.Params(p), time.Time{}, time.Now())
	}

	return op.New(p.ID(), results, op.Success, nil, runner.Params(p), time.Time{}, time.Now())
}

func (p Process) procs(psProcs []ps.Process) ([]Proc, error) {
	var result []Proc

	for _, psProc := range psProcs {
		executable, err := redact.String(psProc.Executable(), p.Redactions)
		if err != nil {
			return result, err
		}
		proc := Proc{
			Name: executable,
			PID:  psProc.Pid(),
			PPID: psProc.PPid(),
		}
		result = append(result, proc)
	}

	return result, nil
}
