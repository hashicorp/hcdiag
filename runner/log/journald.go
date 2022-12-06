// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package log

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/util"

	"github.com/hashicorp/hcdiag/runner"

	"github.com/hashicorp/go-hclog"
)

var _ runner.Runner = Journald{}

// JournaldTimeLayout custom go time layouts must match the reference time Jan 2 15:04:05 2006 MST
const JournaldTimeLayout = "2006-01-02 15:04:05"

type JournaldConfig struct {
	Service string
	DestDir string
	Since   time.Time
	Until   time.Time
	// Redactions includes any redactions to apply to the output of the runner.
	Redactions []*redact.Redact
	// Timeout specifies the amount of time that the runner should be allowed to execute before cancellation.
	Timeout time.Duration
}

type Journald struct {
	ctx context.Context

	Service string    `json:"service"`
	DestDir string    `json:"destDir"`
	Since   time.Time `json:"since"`
	Until   time.Time `json:"until"`
	// Redactions includes any redactions to apply to the output of the runner.
	Redactions []*redact.Redact `json:"redactions"`
	// Timeout specifies the amount of time that the runner should be allowed to execute before cancellation.
	Timeout runner.Timeout `json:"timeout"`
}

// NewJournald sets the defaults for the journald runner
func NewJournald(cfg JournaldConfig) *Journald {
	return NewJournaldWithContext(context.Background(), cfg)
}

// NewJournaldWithContext sets the defaults for the journald runner, including a context.
func NewJournaldWithContext(ctx context.Context, cfg JournaldConfig) *Journald {
	return &Journald{
		ctx:        ctx,
		Service:    cfg.Service,
		DestDir:    cfg.DestDir,
		Since:      cfg.Since,
		Until:      cfg.Until,
		Redactions: cfg.Redactions,
		Timeout:    runner.Timeout(cfg.Timeout),
	}
}

func (j Journald) ID() string {
	return "journald"
}

// Run attempts to pull logs from journald via shell command, e.g.:
// journalctl -x -u {name} --since '3 days ago' --no-pager > {destDir}/journald-{name}.log
func (j Journald) Run() op.Op {
	startTime := time.Now()

	if j.ctx == nil {
		j.ctx = context.Background()
	}

	runCtx := j.ctx
	var cancel context.CancelFunc
	resultChan := make(chan op.Op, 1)
	if 0 < j.Timeout {
		runCtx, cancel = context.WithTimeout(j.ctx, time.Duration(j.Timeout))
		defer cancel()
	}

	go func(ch chan op.Op) {
		o := j.run()
		o.Start = startTime
		ch <- o
	}(resultChan)

	select {
	case <-runCtx.Done():
		switch runCtx.Err() {
		case context.Canceled:
			return runner.CancelOp(j, runCtx.Err(), startTime)
		case context.DeadlineExceeded:
			return runner.TimeoutOp(j, runCtx.Err(), startTime)
		default:
			return op.New(j.ID(), nil, op.Unknown, runCtx.Err(), runner.Params(j), startTime, time.Now())
		}
	case o := <-resultChan:
		return o
	}
}

func (j Journald) run() op.Op {
	s, err := runner.NewShell(runner.ShellConfig{
		Command:    "journalctl --version",
		Redactions: j.Redactions,
	})
	if err != nil {
		return op.New(j.ID(), map[string]any{}, op.Fail, err, runner.Params(j), time.Time{}, time.Now())
	}
	o := s.Run()
	if o.Error != nil {
		return op.New(j.ID(), o.Result, op.Skip, JournaldNotFound{
			service: j.Service,
			err:     o.Error,
		},
			runner.Params(j), time.Time{}, time.Now())
	}

	// Ensure the destination directory exists
	err = util.EnsureDirectory(j.DestDir)
	if err != nil {
		return op.New(j.ID(), nil, op.Fail, err, runner.Params(j), time.Time{}, time.Now())
	}

	// Check if systemd has a unit with the provided name
	cmd := fmt.Sprintf("systemctl is-enabled %s", j.Service)
	cmdCfg := runner.CommandConfig{
		Command:    cmd,
		Format:     "string",
		Redactions: j.Redactions,
	}
	cmdRunner, err := runner.NewCommand(cmdCfg)
	if err != nil {
		return op.New(j.ID(), nil, op.Fail, err, runner.Params(j), time.Time{}, time.Now())
	}

	enabled := cmdRunner.Run()
	if enabled.Error != nil {
		hclog.L().Debug("skipping journald", "service", j.Service, "output", enabled.Result, "error", enabled.Error)
		return op.New(j.ID(), enabled.Result, op.Skip, JournaldServiceNotEnabled{
			service: j.Service,
			command: cmd,
			result:  fmt.Sprintf("%s", enabled.Result),
			err:     enabled.Error,
		},
			runner.Params(j), time.Time{}, time.Now())
	}

	// check if user is able to read messages
	sMessages, err := runner.NewShell(runner.ShellConfig{
		Command:    fmt.Sprintf("journalctl -n0 -u %s 2>&1 | grep -A10 'not seeing messages from other users'", j.Service),
		Redactions: j.Redactions,
	})
	if err != nil {
		return op.New(j.ID(), map[string]any{}, op.Fail, err, runner.Params(j), time.Time{}, time.Now())
	}
	permissions := sMessages.Run()
	// permissions error detected
	if permissions.Error == nil {
		return op.New(j.ID(), permissions.Result, op.Fail, JournaldPermissionError{
			service: j.Service,
			command: cmd,
			result:  fmt.Sprintf("%s", permissions.Result),
			err:     permissions.Error,
		},
			runner.Params(j), time.Time{}, time.Now())
	}

	sLogs, err := runner.NewShell(runner.ShellConfig{
		Command:    j.LogsCmd(),
		Redactions: j.Redactions,
	})
	if err != nil {
		return op.New(j.ID(), map[string]any{}, op.Fail, err, runner.Params(j), time.Time{}, time.Now())
	}
	logs := sLogs.Run()
	return op.New(j.ID(), logs.Result, logs.Status, logs.Error, runner.Params(j), time.Time{}, time.Now())
}

// LogsCmd arranges the params into a runnable command string.
func (j Journald) LogsCmd() string {
	// Write since and until flags with formatted time if either has a non-zero value
	// Flag strings handle their own trailing whitespace to avoid having extra spaces when the flag is disabled.
	var sinceFlag, untilFlag string
	if !j.Since.IsZero() {
		t := j.Since.Format(JournaldTimeLayout)
		sinceFlag = fmt.Sprintf("--since \"%s\" ", t)
	}
	if !j.Until.IsZero() {
		t := j.Until.Format(JournaldTimeLayout)
		untilFlag = fmt.Sprintf("--until \"%s\" ", t)
	}

	// Add our write destination
	dest := fmt.Sprintf("%s/journald-%s.log", j.DestDir, j.Service)

	// Compose the service name with flags and write to dest
	cmd := fmt.Sprintf("journalctl -x -u %s %s%s--no-pager > %s", j.Service, sinceFlag, untilFlag, dest)
	return cmd
}

type JournaldServiceNotEnabled struct {
	service string
	command string
	result  string
	err     error
}

func (e JournaldServiceNotEnabled) Error() string {
	return fmt.Sprintf("service not enabled in sysctl, service=%s, command=%s, result=%s, error=%s", e.service, e.command, e.result, e.err)
}

func (e JournaldServiceNotEnabled) Unwrap() error {
	return e.err
}

type JournaldPermissionError struct {
	service string
	command string
	result  string
	err     error
}

func (e JournaldPermissionError) Error() string {
	return fmt.Sprintf("unable to read logs, is your user allowed? try sudo, service=%s, command=%s, result=%s, error=%s", e.service, e.command, e.result, e.err)
}

func (e JournaldPermissionError) Unwrap() error {
	return e.err
}

type JournaldNotFound struct {
	service string
	err     error
}

func (e JournaldNotFound) Error() string {
	return fmt.Sprintf("journald not found on this system, service=%s, error=%s", e.service, e.err)
}

func (e JournaldNotFound) Unwrap() error {
	return e.err
}
