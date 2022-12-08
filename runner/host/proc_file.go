// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package host

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = ProcFile{}

type ProcFileConfig struct {
	OS         string
	Redactions []*redact.Redact
	Timeout    runner.Timeout
}

type ProcFile struct {
	ctx context.Context

	OS         string           `json:"os"`
	Commands   []string         `json:"commands"`
	Redactions []*redact.Redact `json:"redactions"`
	Timeout    runner.Timeout   `json:"timeout"`
}

func NewProcFile(cfg ProcFileConfig) *ProcFile {
	return NewProcFileWithContext(context.Background(), cfg)
}

func NewProcFileWithContext(ctx context.Context, cfg ProcFileConfig) *ProcFile {
	commands := []string{
		"cat /proc/cpuinfo",
		"cat /proc/loadavg",
		"cat /proc/version",
		"cat /proc/vmstat",
	}
	return &ProcFile{
		ctx:        ctx,
		OS:         cfg.OS,
		Commands:   commands,
		Redactions: cfg.Redactions,
		Timeout:    cfg.Timeout,
	}
}

func (p ProcFile) ID() string {
	return "/proc/ files"
}

func (p ProcFile) Run() op.Op {
	startTime := time.Now()

	var runCtx context.Context
	runCtx = p.ctx
	var runCancelFunc context.CancelFunc
	if p.Timeout > 0 {
		runCtx, runCancelFunc = context.WithTimeout(p.ctx, time.Duration(p.Timeout))
		defer runCancelFunc()
	}

	resultsChannel := make(chan op.Op, 1)
	go func(results chan<- op.Op) {
		o := p.run()
		o.Start = startTime
		resultsChannel <- o
	}(resultsChannel)
	select {
	case <-runCtx.Done():
		switch runCtx.Err() {
		case context.Canceled:
			return op.NewCancel(p.ID(), runCtx.Err(), runner.Params(p), startTime)
		case context.DeadlineExceeded:
			return op.NewTimeout(p.ID(), runCtx.Err(), runner.Params(p), startTime)
		default:
			return op.New(p.ID(), nil, op.Unknown, runCtx.Err(), runner.Params(p), startTime, time.Now())
		}
	case result := <-resultsChannel:
		return result
	}
}

func (p ProcFile) run() op.Op {
	result := make(map[string]any)
	if p.OS != "linux" {
		return op.New(p.ID(), nil, op.Skip, fmt.Errorf("os not linux, skipping, os=%s", p.OS), runner.Params(p), time.Time{}, time.Now())
	}
	for _, c := range p.Commands {
		shell, err := runner.NewShell(runner.ShellConfig{
			Command:    c,
			Redactions: p.Redactions,
		})
		if err != nil {
			return op.New(p.ID(), map[string]any{}, op.Fail, err, runner.Params(p), time.Time{}, time.Now())
		}
		o := shell.Run()
		if o.Error != nil {
			return op.New(p.ID(), o.Result, op.Fail, o.Error, runner.Params(p), time.Time{}, time.Now())
		}
		result[o.Identifier] = o.Result
	}
	return op.New(p.ID(), result, op.Success, nil, runner.Params(p), time.Time{}, time.Now())
}
