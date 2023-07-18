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

var _ runner.Runner = Lsof{}

// LsofConfig takes each parameter to configure an lsof runner implementation.
type LsofConfig struct {
	OS         string           `json:"os"`
	Timeout    runner.Timeout   `json:"timeout"`
	Redactions []*redact.Redact `json:"redactions"`
}

type Lsof struct {
	OS      string         `json:"os"`
	Shell   runner.Runner  `json:"shell"`
	Timeout runner.Timeout `json:"timeout"`
	ctx     context.Context
}

// NewLsof takes LsofConfig and returns a runnable lsof.
func NewLsof(cfg LsofConfig) (*Lsof, error) {
	return NewLsofWithContext(context.Background(), cfg)
}

// NewLsofWithContext takes a Context and LsofConfig and returns a runnable lsof.
func NewLsofWithContext(ctx context.Context, cfg LsofConfig) (*Lsof, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	shell, err := runner.NewShellWithContext(ctx, runner.ShellConfig{
		Command:    "lsof",
		Redactions: cfg.Redactions,
		Timeout:    time.Duration(cfg.Timeout),
	})
	if err != nil {
		return nil, err
	}
	return &Lsof{
		OS:    cfg.OS,
		Shell: shell,
		ctx:   ctx,
	}, nil
}

// ID returns the runner ID for lsof.
func (r Lsof) ID() string {
	return "lsof"
}

// Run executes the lsof Runner and returns an op.Op result.
func (r Lsof) Run() op.Op {
	startTime := time.Now()

	// Only Linux and Darwin is supported currently; lsof is not natively available on windows.
	if r.OS != "linux" && r.OS != "darwin" {
		return op.New(r.ID(), nil, op.Skip, fmt.Errorf("Lsof.Run() not available on os, os=%s", r.OS), runner.Params(r), startTime, time.Now())
	}
	o := r.Shell.Run()
	return op.New(r.ID(), o.Result, o.Status, o.Error, runner.Params(r), startTime, time.Now())
}
