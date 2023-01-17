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

var _ runner.Runner = FSTab{}

// FSTabConfig takes each parameter to configure an FSTab runner implementation.
type FSTabConfig struct {
	OS         string           `json:"os"`
	Timeout    runner.Timeout   `json:"timeout"`
	Redactions []*redact.Redact `json:"redactions"`
}

// FSTab accepts
type FSTab struct {
	OS      string         `json:"os"`
	Shell   runner.Runner  `json:"shell"`
	Timeout runner.Timeout `json:"timeout"`
	ctx     context.Context
}

// NewFSTab takes FSTabConfig and returns a runnable FSTab.
func NewFSTab(cfg FSTabConfig) (*FSTab, error) {
	return NewFSTabWithContext(context.Background(), cfg)
}

// NewFSTabWithContext takes a Context and FSTabConfig and returns a runnable FSTab.
func NewFSTabWithContext(ctx context.Context, cfg FSTabConfig) (*FSTab, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	shell, err := runner.NewShellWithContext(ctx, runner.ShellConfig{
		Command:    "cat /etc/fstab",
		Redactions: cfg.Redactions,
		Timeout:    time.Duration(cfg.Timeout),
	})
	if err != nil {
		return nil, err
	}
	return &FSTab{
		OS:    cfg.OS,
		Shell: shell,
	}, nil
}

// ID returns the runner ID for FSTab.
func (r FSTab) ID() string {
	return "/etc/fstab"
}

// Run executes the FSTab Runner and returns an op.Op result.
func (r FSTab) Run() op.Op {
	startTime := time.Now()

	// Only Linux is supported currently; Windows is unsupported, and Darwin doesn't use /etc/fstab by default.
	if r.OS != "linux" {
		return op.New(r.ID(), nil, op.Skip, fmt.Errorf("FSTab.Run() not available on os, os=%s", r.OS), runner.Params(r), startTime, time.Now())
	}
	o := r.Shell.Run()
	return op.New(r.ID(), o.Result, o.Status, o.Error, runner.Params(r), startTime, time.Now())
}
