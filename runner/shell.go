// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package runner

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/hcdiag/op"

	"github.com/hashicorp/hcdiag/util"
)

type ShellConfig struct {
	Command    string
	Redactions []*redact.Redact
	Timeout    time.Duration
}

// Shell runs shell commands in a real unix shell.
type Shell struct {
	ctx context.Context

	Command    string           `json:"command"`
	Shell      string           `json:"shell"`
	Redactions []*redact.Redact `json:"redactions"`
	Timeout    Timeout          `json:"timeout"`
}

// NewShell provides a runner for arbitrary shell code.
func NewShell(cfg ShellConfig) (*Shell, error) {
	return NewShellWithContext(context.Background(), cfg)
}

// NewShellWithContext provides a runner for arbitrary shell code.
func NewShellWithContext(ctx context.Context, cfg ShellConfig) (*Shell, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	timeout := cfg.Timeout
	if timeout < 0 {
		return nil, fmt.Errorf("timeout must be a nonnegative, timeout='%s'", timeout.String())
	}

	return &Shell{
		ctx:        ctx,
		Command:    cfg.Command,
		Redactions: cfg.Redactions,
		Timeout:    Timeout(timeout),
	}, nil
}

func (s Shell) ID() string {
	return s.Command
}

// Run ensures a shell exists and optimistically executes the given Command string
func (s Shell) Run() op.Op {
	startTime := time.Now()

	if s.ctx == nil {
		s.ctx = context.Background()
	}

	runCtx := s.ctx
	var cancel context.CancelFunc
	if 0 < s.Timeout {
		runCtx, cancel = context.WithTimeout(s.ctx, time.Duration(s.Timeout))
		defer cancel()
	}

	resChan := make(chan op.Op, 1)
	go func(resChan chan<- op.Op, start time.Time) {
		o := s.run()
		o.Start = start
		resChan <- o
	}(resChan, startTime)

	select {
	case <-runCtx.Done():
		switch runCtx.Err() {
		case context.Canceled:
			return op.NewCancel(s.ID(), runCtx.Err(), Params(s), startTime)
		case context.DeadlineExceeded:
			return op.NewTimeout(s.ID(), runCtx.Err(), Params(s), startTime)
		default:
			return op.New(s.ID(), nil, op.Unknown, runCtx.Err(), Params(s), startTime, time.Now())
		}
	case result := <-resChan:
		return result
	}
}

func (s Shell) run() op.Op {
	// Read the shell from the environment
	shell, err := util.GetShell()
	if err != nil {
		return op.New(s.ID(), nil, op.Fail, err, Params(s), time.Time{}, time.Now())
	}
	s.Shell = shell

	// Run the command
	args := []string{"-c", s.Command}
	bts, cmdErr := exec.Command(s.Shell, args...).CombinedOutput()
	// Store and redact the result before cmd error handling, so we can return it in error and success cases.
	redBts, redErr := redact.Bytes(bts, s.Redactions)
	// Fail run if unable to redact
	if redErr != nil {
		return op.New(s.ID(), nil, op.Fail, redErr, Params(s), time.Time{}, time.Now())
	}
	if cmdErr != nil {
		result := map[string]any{"shell": string(redBts)}
		return op.New(s.ID(), result, op.Unknown,
			ShellExecError{
				command: s.Command,
				err:     cmdErr,
			}, Params(s), time.Time{}, time.Now())
	}
	result := map[string]any{"shell": string(redBts)}
	return op.New(s.ID(), result, op.Success, nil, Params(s), time.Time{}, time.Now())
}

type ShellExecError struct {
	command string
	err     error
}

func (e ShellExecError) Error() string {
	return fmt.Sprintf("exec error, command=%s, error=%s", e.command, e.err.Error())
}

func (e ShellExecError) Unwrap() error {
	return e.err
}
