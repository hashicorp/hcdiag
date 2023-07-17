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

var _ runner.Runner = DMesg{}

// DMesgConfig takes each parameter to configure an DMesg runner implementation.
type DMesgConfig struct {
	OS         string           `json:"os"`
	Timeout    runner.Timeout   `json:"timeout"`
	Redactions []*redact.Redact `json:"redactions"`
}

// DMesg accepts
type DMesg struct {
	OS      string         `json:"os"`
	Shell   runner.Runner  `json:"shell"`
	Timeout runner.Timeout `json:"timeout"`
	ctx     context.Context
}

// NewDMesg takes DMesgConfig and returns a runnable DMesg.
func NewDMesg(cfg DMesgConfig) (*DMesg, error) {
	return NewDMesgWithContext(context.Background(), cfg)
}

// NewDMesgWithContext takes a Context and DMesgConfig and returns a runnable DMesg.
func NewDMesgWithContext(ctx context.Context, cfg DMesgConfig) (*DMesg, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	shell, err := runner.NewShellWithContext(ctx, runner.ShellConfig{
		Command:    "dmesg -T",
		Redactions: cfg.Redactions,
		Timeout:    time.Duration(cfg.Timeout),
	})
	if err != nil {
		return nil, err
	}
	return &DMesg{
		OS:    cfg.OS,
		Shell: shell,
		ctx:   ctx,
	}, nil
}

// ID returns the runner ID for DMesg.
func (r DMesg) ID() string {
	return "dmesg -T"
}

// Run executes the DMesg Runner and returns an op.Op result.
func (r DMesg) Run() op.Op {
	startTime := time.Now()

	// Only Linux is supported currently; Windows is unsupported, and Darwin doesn't use dmesg -T (uses dmesg) by default.
	if r.OS != "linux" {
		return op.New(r.ID(), nil, op.Skip, fmt.Errorf("DMesg.Run() not available on os, os=%s", r.OS), runner.Params(r), startTime, time.Now())
	}
	o := r.Shell.Run()
	return op.New(r.ID(), o.Result, o.Status, o.Error, runner.Params(r), startTime, time.Now())
}
