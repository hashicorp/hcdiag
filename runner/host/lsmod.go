// Copyright (c) HashiCorp, Inc.
package host

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = Lsmod{}

// LsmodConfig takes each parameter to configure an lsmod runner implementation.
type LsmodConfig struct {
	OS         string           `json:"os"`
	Timeout    runner.Timeout   `json:"timeout"`
	Redactions []*redact.Redact `json:"redactions"`
}

// lsmod accepts
type Lsmod struct {
	OS      string         `json:"os"`
	Shell   runner.Runner  `json:"shell"`
	Timeout runner.Timeout `json:"timeout"`
	ctx     context.Context
}

// NewLsmod takes LsmodConfig and returns a runnable lsmod.
func NewLsmod(cfg LsmodConfig) (*Lsmod, error) {
	return NewLsmodWithContext(context.Background(), cfg)
}

// NewlsmodWithContext takes a Context and lsmodConfig and returns a runnable lsmod.
func NewLsmodWithContext(ctx context.Context, cfg LsmodConfig) (*Lsmod, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	shell, err := runner.NewShellWithContext(ctx, runner.ShellConfig{
		Command:    "lsmod",
		Redactions: cfg.Redactions,
		Timeout:    time.Duration(cfg.Timeout),
	})
	if err != nil {
		return nil, err
	}
	return &Lsmod{
		OS:    cfg.OS,
		Shell: shell,
		ctx:   ctx,
	}, nil
}

// ID returns the runner ID for lsmod.
func (r Lsmod) ID() string {
	return "lsmod"
}

// Run executes the lsmod Runner and returns an op.Op result.
func (r Lsmod) Run() op.Op {
	startTime := time.Now()

	// Only Linux is supported currently; Windows is unsupported, and Darwin doesn't use lsmod by default.
	if r.OS != "linux" {
		return op.New(r.ID(), nil, op.Skip, fmt.Errorf("Lsmod.Run() not available on os, os=%s", r.OS), runner.Params(r), startTime, time.Now())
	}
	o := r.Shell.Run()
	return op.New(r.ID(), o.Result, o.Status, o.Error, runner.Params(r), startTime, time.Now())
}
