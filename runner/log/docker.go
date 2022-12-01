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
)

var _ runner.Runner = Docker{}

type DockerConfig struct {
	// Container is the name of the docker container to get logs from
	Container string
	// DestDir is the directory we will write the logs to
	DestDir string
	// Since marks the beginning of the time range to include logs
	Since time.Time
	// Redactions includes any redactions to apply to the output of the runner.
	Redactions []*redact.Redact
	// Timeout specifies the amount of time that the runner should be allowed to execute before cancellation.
	Timeout time.Duration
}

// Docker allows logs to be retrieved for a docker container
type Docker struct {
	ctx context.Context

	// Container is the name of the docker container to get logs from
	Container string `json:"container"`
	// DestDir is the directory we will write the logs to
	DestDir string `json:"destDir"`
	// Since marks the beginning of the time range to include logs
	Since time.Time `json:"since"`
	// Redactions includes any redactions to apply to the output of the runner.
	Redactions []*redact.Redact `json:"redactions"`
	// Timeout specifies the amount of time that the runner should be allowed to execute before cancellation.
	Timeout runner.Timeout `json:"timeout"`
}

// NewDocker returns a runner with an identifier and fully configured docker runner
func NewDocker(cfg DockerConfig) *Docker {
	return NewDockerWithContext(context.Background(), cfg)
}

// NewDockerWithContext returns a runner with an identifier and fully configured docker runner, which includes a provided context.
func NewDockerWithContext(ctx context.Context, cfg DockerConfig) *Docker {
	return &Docker{
		ctx:        ctx,
		Container:  cfg.Container,
		DestDir:    cfg.DestDir,
		Since:      cfg.Since,
		Redactions: cfg.Redactions,
		Timeout:    runner.Timeout(cfg.Timeout),
	}
}

func (d Docker) ID() string {
	return "log/docker " + d.Container
}

// Run executes the runner
func (d Docker) Run() op.Op {
	startTime := time.Now()

	if d.ctx == nil {
		d.ctx = context.Background()
	}

	runCtx := d.ctx
	var cancel context.CancelFunc
	resultChan := make(chan op.Op, 1)
	if 0 < d.Timeout {
		runCtx, cancel = context.WithTimeout(d.ctx, time.Duration(d.Timeout))
		defer cancel()
	}

	go func(ch chan op.Op) {
		o := d.run()
		o.Start = startTime
		ch <- o
	}(resultChan)

	select {
	case <-runCtx.Done():
		switch runCtx.Err() {
		case context.Canceled:
			return runner.CancelOp(d, runCtx.Err(), startTime)
		case context.DeadlineExceeded:
			return runner.TimeoutOp(d, runCtx.Err(), startTime)
		default:
			return op.New(d.ID(), nil, op.Unknown, runCtx.Err(), runner.Params(d), startTime, time.Now())
		}
	case o := <-resultChan:
		return o
	}
}

func (d Docker) run() op.Op {
	// Check that docker exists
	version, err := runner.NewShell(runner.ShellConfig{
		Command:    "docker version",
		Redactions: d.Redactions,
	})
	if err != nil {
		return op.New(d.ID(), map[string]any{}, op.Fail, err, runner.Params(d), time.Time{}, time.Now())
	}
	versionOp := version.Run()
	if versionOp.Error != nil {
		return op.New(d.ID(), versionOp.Result, op.Skip, DockerNotFoundError{
			container: d.Container,
			err:       versionOp.Error,
		},
			runner.Params(d), time.Time{}, time.Now())
	}

	// Check whether the container can be found on the system
	ok, err := d.containerExists()
	if err != nil {
		return op.New(d.ID(), map[string]any{}, op.Fail, err, runner.Params(d), time.Time{}, time.Now())
	}
	if !ok {
		return op.New(d.ID(), map[string]any{}, op.Skip, ContainerNotFoundError{
			container: d.Container,
		},
			runner.Params(d), time.Time{}, time.Now())
	}

	// Ensure the destination directory exists
	err = util.EnsureDirectory(d.DestDir)
	if err != nil {
		return op.New(d.ID(), nil, op.Fail, err, runner.Params(d), time.Time{}, time.Now())
	}

	// Retrieve logs
	logShell, err := runner.NewShell(runner.ShellConfig{
		Command:    DockerLogCmd(d.Container, d.DestDir, d.Since),
		Redactions: d.Redactions,
	})
	if err != nil {
		return op.New(d.ID(), map[string]any{}, op.Fail, err, runner.Params(d), time.Time{}, time.Now())
	}
	logOp := logShell.Run()
	if logOp.Error != nil {
		return op.New(d.ID(), logOp.Result, logOp.Status, logOp.Error, runner.Params(d), time.Time{}, time.Now())
	}

	return op.New(d.ID(), logOp.Result, op.Success, nil, runner.Params(d), time.Time{}, time.Now())
}

func DockerLogCmd(container, destDir string, since time.Time) string {
	var sinceFlag string
	if !since.IsZero() {
		sinceFlag = fmt.Sprintf(" --since %s", since.Format(time.RFC3339))
	}

	// Add our write destination
	dest := fmt.Sprintf("%s/docker-%s.log", destDir, container)

	// Compose the service name with flags and write to dest
	cmd := fmt.Sprintf("docker logs --timestamps%s %s > %s", sinceFlag, container, dest)
	return cmd
}

func (d Docker) containerExists() (bool, error) {
	// attempt to inspect the container by name, to ensure it exists
	s, err := runner.NewShell(runner.ShellConfig{
		Command:    fmt.Sprintf("docker container inspect %s > /dev/null 2>&1", d.Container),
		Redactions: d.Redactions,
	})
	if err != nil {
		return false, err
	}
	o := s.Run()
	if o.Error != nil {
		return false, o.Error
	}
	return true, nil
}

var _ error = DockerNotFoundError{}

type DockerNotFoundError struct {
	container string
	err       error
}

func (e DockerNotFoundError) Error() string {
	return fmt.Sprintf("docker not found, container=%s, error=%s", e.container, e.err)
}

func (e DockerNotFoundError) Unwrap() error {
	return e.err
}

var _ error = DockerNoLogsError{}

type DockerNoLogsError struct {
	container string
}

func (e DockerNoLogsError) Error() string {
	return fmt.Sprintf("docker container found but results were empty, container=%s", e.container)
}

var _ error = ContainerNotFoundError{}

type ContainerNotFoundError struct {
	container string
}

func (e ContainerNotFoundError) Error() string {
	return fmt.Sprintf("docker container not found, container=%s", e.container)
}
