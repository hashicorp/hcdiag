// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package host

import (
	"context"
	"time"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = OS{}

type OSConfig struct {
	// OS is the operating system family of the host.
	OS string
	// Redactions includes any redactions to apply to the output of the runner.
	Redactions []*redact.Redact
	// Timeout specifies the amount of time that the runner should be allowed to execute before cancellation.
	Timeout time.Duration
}

type OS struct {
	ctx context.Context

	// OS is the operating system family of the host.
	OS string `json:"os"`
	// Command is the command that will execute to gather OS details.
	Command string `json:"command"`
	// Redactions includes any redactions to apply to the output of the runner.
	Redactions []*redact.Redact `json:"redactions"`
	// Timeout specifies the amount of time that the runner should be allowed to execute before cancellation.
	Timeout runner.Timeout `json:"timeout"`
}

func NewOS(cfg OSConfig) *OS {
	return NewOSWithContext(context.Background(), cfg)
}

func NewOSWithContext(ctx context.Context, cfg OSConfig) *OS {
	os := cfg.OS
	osCmd := "uname -v"
	if os == "windows" {
		osCmd = "systeminfo"
	}

	return &OS{
		ctx:        ctx,
		OS:         os,
		Command:    osCmd,
		Redactions: cfg.Redactions,
		Timeout:    runner.Timeout(cfg.Timeout),
	}
}

func (o OS) ID() string {
	return o.Command
}

// Run calls the given OS utility to get information on the operating system
func (o OS) Run() op.Op {
	startTime := time.Now()
	cmdCfg := runner.CommandConfig{
		Command:    o.Command,
		Format:     "string",
		Redactions: o.Redactions,
		Timeout:    time.Duration(o.Timeout),
	}
	cmdRunner, err := runner.NewCommandWithContext(o.ctx, cmdCfg)
	if err != nil {
		return op.New(o.ID(), nil, op.Fail, err, runner.Params(o), startTime, time.Now())
	}
	// NOTE(mkcp): This runner can be made consistent between multiple operating systems if we parse the output of
	// systeminfo to match uname's scope of concerns.
	c := cmdRunner.Run()
	return op.New(o.ID(), c.Result, c.Status, c.Error, runner.Params(o), startTime, time.Now())
}
