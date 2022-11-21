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

var _ runner.Runner = EtcHosts{}

type EtcHostsConfig struct {
	OS         string
	Redactions []*redact.Redact
	Timeout    runner.Timeout
}

type EtcHosts struct {
	ctx        context.Context
	OS         string           `json:"os"`
	Redactions []*redact.Redact `json:"redactions"`
	Timeout    runner.Timeout   `json:"timeout"`
}

func NewEtcHosts(cfg EtcHostsConfig) *EtcHosts {
	return NewEtcHostsWithContext(context.Background(), cfg)
}

func NewEtcHostsWithContext(ctx context.Context, cfg EtcHostsConfig) *EtcHosts {
	os := cfg.OS
	if os == "" {
		os = runtime.GOOS
	}
	return &EtcHosts{
		ctx:        ctx,
		OS:         os,
		Redactions: cfg.Redactions,
		Timeout:    cfg.Timeout,
	}
}

func (r EtcHosts) ID() string {
	return "/etc/hosts"
}

func (r EtcHosts) Run() op.Op {
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

	go func(ch chan op.Op) {
		o := r.run()
		o.Start = startTime
		ch <- o
	}(resultChan)

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

func (r EtcHosts) run() op.Op {
	// Not compatible with windows
	if r.OS == "windows" {
		err := fmt.Errorf(" EtcHosts.Run() not available on os, os=%s", r.OS)
		return op.New(r.ID(), nil, op.Skip, err, runner.Params(r), time.Time{}, time.Now())
	}

	s, err := runner.NewShell(runner.ShellConfig{
		Command:    "cat /etc/hosts",
		Redactions: r.Redactions,
		Timeout:    time.Duration(r.Timeout),
	})
	if err != nil {
		return op.New(r.ID(), map[string]any{}, op.Fail, err, runner.Params(r), time.Time{}, time.Now())
	}

	o := s.Run()
	if o.Error != nil {
		return o
	}
	return op.New(r.ID(), o.Result, op.Success, nil, runner.Params(r), time.Time{}, time.Now())
}
