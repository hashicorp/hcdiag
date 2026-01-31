// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package host

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = IPTables{}

type IPTablesConfig struct {
	OS         string
	Redactions []*redact.Redact
	Timeout    runner.Timeout
}

type IPTables struct {
	ctx context.Context

	OS         string           `json:"os"`
	Commands   []string         `json:"commands"`
	Redactions []*redact.Redact `json:"redactions"`
	Timeout    runner.Timeout   `json:"timeout"`
}

// NewIPTables returns a runner configured to run several iptables commands
func NewIPTables(cfg IPTablesConfig) *IPTables {
	return NewIPTablesWithContext(context.Background(), cfg)
}

// NewIPTablesWithContext returns a runner configured to run several iptables commands
func NewIPTablesWithContext(ctx context.Context, cfg IPTablesConfig) *IPTables {
	commands := []string{
		"iptables -L -n -v",
		"iptables -L -n -v -t nat",
		"iptables -L -n -v -t mangle",
	}
	return &IPTables{
		ctx:        ctx,
		OS:         cfg.OS,
		Commands:   commands,
		Redactions: cfg.Redactions,
		Timeout:    cfg.Timeout,
	}
}

func (r IPTables) ID() string {
	return "iptables"
}

func (r IPTables) Run() op.Op {
	startTime := time.Now()

	if r.ctx == nil {
		r.ctx = context.Background()
	}

	runCtx := r.ctx
	var cancel context.CancelFunc
	resultChan := make(chan op.Op, 1)
	if 0 < r.Timeout {
		runCtx, cancel = context.WithTimeout(r.ctx, time.Duration(r.Timeout))
		defer cancel()
	}

	go func(ctx context.Context, ch chan op.Op) {
		o := r.run(ctx)
		o.Start = startTime
		ch <- o
	}(runCtx, resultChan)

	select {
	case <-runCtx.Done():
		switch runCtx.Err() {
		case context.Canceled:
			return runner.CancelOp(r, runCtx.Err(), startTime)
		case context.DeadlineExceeded:
			return runner.TimeoutOp(r, runCtx.Err(), startTime)
		default:
			return op.New(r.ID(), nil, op.Unknown, runCtx.Err(), runner.Params(r), startTime, time.Now())
		}
	case o := <-resultChan:
		return o
	}
}

func (r IPTables) run(ctx context.Context) op.Op {
	if r.OS != "linux" {
		return op.New(r.ID(), nil, op.Skip, fmt.Errorf("os not linux, skipping, os=%s", runtime.GOOS), runner.Params(r), time.Time{}, time.Now())
	}
	result := make(map[string]any)
	for _, c := range r.Commands {
		cmdCfg := runner.CommandConfig{
			Command:    c,
			Format:     "string",
			Redactions: r.Redactions,
			Timeout:    time.Duration(r.Timeout),
		}
		cmdRunner, err := runner.NewCommandWithContext(ctx, cmdCfg)
		if err != nil {
			return op.New(r.ID(), nil, op.Fail, err, runner.Params(r), time.Time{}, time.Now())
		}
		o := cmdRunner.Run()
		result[c] = o.Result
	}
	return op.New(r.ID(), result, op.Success, nil, runner.Params(r), time.Time{}, time.Now())
}
