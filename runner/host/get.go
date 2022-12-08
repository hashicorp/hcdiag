// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package host

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/hcdiag/op"

	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = Get{}

type GetConfig struct {
	Path string
	// Redactions includes any redactions to apply to the output of the runner.
	Redactions []*redact.Redact
	// Timeout specifies the amount of time that the runner should be allowed to execute before cancellation.
	Timeout time.Duration
}

type Get struct {
	ctx context.Context

	Path string `json:"path"`
	// Redactions includes any redactions to apply to the output of the runner.
	Redactions []*redact.Redact `json:"redactions"`
	// Timeout specifies the amount of time that the runner should be allowed to execute before cancellation.
	Timeout runner.Timeout `json:"timeout"`
}

func NewGet(cfg GetConfig) (*Get, error) {
	return NewGetWithContext(context.Background(), cfg)
}

func NewGetWithContext(ctx context.Context, cfg GetConfig) (*Get, error) {
	return &Get{
		Path:       cfg.Path,
		Redactions: cfg.Redactions,
		Timeout:    runner.Timeout(cfg.Timeout),
	}, nil
}

func (g Get) ID() string {
	return "GET" + " " + g.Path
}

func (g Get) Run() op.Op {
	startTime := time.Now()

	if g.ctx == nil {
		g.ctx = context.Background()
	}

	runCtx := g.ctx
	var cancel context.CancelFunc
	resultChan := make(chan op.Op, 1)
	if 0 < g.Timeout {
		runCtx, cancel = context.WithTimeout(g.ctx, time.Duration(g.Timeout))
		defer cancel()
	}

	go func(ch chan op.Op) {
		o := g.run()
		o.Start = startTime
		ch <- o
	}(resultChan)

	select {
	case <-runCtx.Done():
		switch runCtx.Err() {
		case context.Canceled:
			return runner.CancelOp(g, runCtx.Err(), startTime)
		case context.DeadlineExceeded:
			return runner.TimeoutOp(g, runCtx.Err(), startTime)
		default:
			return op.New(g.ID(), nil, op.Unknown, runCtx.Err(), runner.Params(g), startTime, time.Now())
		}
	case o := <-resultChan:
		return o
	}
}

func (g Get) run() op.Op {
	cmd := strings.Join([]string{"curl -s", g.Path}, " ")
	// NOTE(mkcp): We will get JSON back from a lot of requests, so this can be improved
	format := "string"
	cmdCfg := runner.CommandConfig{
		Command:    cmd,
		Format:     format,
		Redactions: g.Redactions,
	}
	cmdRunner, err := runner.NewCommand(cmdCfg)
	if err != nil {
		return op.New(g.ID(), nil, op.Fail, err, runner.Params(g), time.Time{}, time.Now())
	}
	o := cmdRunner.Run()
	return op.New(g.ID(), o.Result, o.Status, o.Error, runner.Params(g), time.Time{}, time.Now())
}
