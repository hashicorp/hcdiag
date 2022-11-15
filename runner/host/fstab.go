package host

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = FSTab{}

type FSTab struct {
	OS    string        `json:"os"`
	Shell runner.Runner `json:"shell"`

	ctx        context.Context
	Timeout    runner.Timeout   `json:"timeout"`
	Redactions []*redact.Redact `json:"redactions"`
}

type FSTabConfig struct {
	OS         string
	Timeout    runner.Timeout
	Redactions []*redact.Redact
}

func NewFSTab(cfg FSTabConfig) *FSTab {
	return NewFSTabWithContext(nil, cfg)
}

func NewFSTabWithContext(ctx context.Context, cfg FSTabConfig) *FSTab {
	if ctx == nil {
		ctx = context.Background()
	}

	return &FSTab{
		ctx:        ctx,
		OS:         cfg.OS,
		Timeout:    cfg.Timeout,
		Redactions: cfg.Redactions,
	}
}

func (r FSTab) ID() string {
	return "/etc/fstab"
}

func (r FSTab) Run() op.Op {
	startTime := time.Now()

	var cancel context.CancelFunc
	var runCtx context.Context
	runCtx = r.ctx
	if r.Timeout > 0 {
		runCtx, cancel = context.WithTimeout(r.ctx, time.Duration(r.Timeout))
		defer cancel()
	}

	res := make(chan op.Op, 1)
	go func(results chan<- op.Op) {
		o := r.run()
		o.Start = startTime
		results <- o
	}(res)
	select {
	case <-runCtx.Done():
		switch runCtx.Err() {
		case context.Canceled:
			return op.New(r.ID(), nil, op.Canceled, r.ctx.Err(), runner.Params(r), startTime, time.Now())
		case context.DeadlineExceeded:
			return op.New(r.ID(), nil, op.Timeout, r.ctx.Err(), runner.Params(r), startTime, time.Now())
		default:
			return op.New(r.ID(), nil, op.Unknown, r.ctx.Err(), runner.Params(r), startTime, time.Now())
		}

	case o := <-res:
		return o
	}
}

func (r FSTab) run() op.Op {
	// Only Linux is supported currently; Windows is unsupported, and Darwin doesn't use /etc/fstab by default.
	if r.OS != "linux" {
		return op.New(r.ID(), nil, op.Skip, fmt.Errorf("FSTab.Run() not available on os, os=%s", r.OS), runner.Params(r), time.Time{}, time.Now())
	}

	shell := runner.NewShell("cat /etc/fstab", r.Redactions)
	o := shell.Run()
	if o.Error != nil {
		return op.New(r.ID(), o.Result, op.Fail, o.Error, runner.Params(r), time.Time{}, time.Now())
	}
	return op.New(r.ID(), o.Result, op.Success, nil, runner.Params(r), time.Time{}, time.Now())
}
