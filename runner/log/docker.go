// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package log

import (
	"fmt"
	"time"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = Docker{}

// Docker allows logs to be retrieved for a docker container
type Docker struct {
	// Container is the name of the docker container to get logs from
	Container string `json:"container"`
	// DestDir is the directory we will write the logs to
	DestDir string `json:"destDir"`
	// Since marks the beginning of the time range to include logs
	Since      time.Time        `json:"since"`
	Redactions []*redact.Redact `json:"redactions"`
}

// NewDocker returns a runner with an identifier and fully configured docker runner
func NewDocker(container, destDir string, since time.Time, redactions []*redact.Redact) *Docker {
	return &Docker{
		Container:  container,
		DestDir:    destDir,
		Since:      since,
		Redactions: redactions,
	}
}

func (d Docker) ID() string {
	return "log/docker " + d.Container
}

// Run executes the runner
func (d Docker) Run() op.Op {
	startTime := time.Now()

	// Check that docker exists
	version, err := runner.NewShell(runner.ShellConfig{
		Command:    "docker version",
		Redactions: d.Redactions,
	})
	if err != nil {
		return op.New(d.ID(), map[string]any{}, op.Fail, err, runner.Params(d), startTime, time.Now())
	}
	versionOp := version.Run()
	if versionOp.Error != nil {
		return op.New(d.ID(), versionOp.Result, op.Skip, DockerNotFoundError{
			container: d.Container,
			err:       versionOp.Error,
		},
			runner.Params(d), startTime, time.Now())
	}

	// Check whether the container can be found on the system
	ok, err := d.containerExists()
	if err != nil {
		return op.New(d.ID(), map[string]any{}, op.Fail, err, runner.Params(d), startTime, time.Now())
	}
	if !ok {
		return op.New(d.ID(), map[string]any{}, op.Skip, ContainerNotFoundError{
			container: d.Container,
		},
			runner.Params(d), startTime, time.Now())
	}

	// Retrieve logs
	logShell, err := runner.NewShell(runner.ShellConfig{
		Command:    DockerLogCmd(d.Container, d.DestDir, d.Since),
		Redactions: d.Redactions,
	})
	if err != nil {
		return op.New(d.ID(), map[string]any{}, op.Fail, err, runner.Params(d), startTime, time.Now())
	}
	logOp := logShell.Run()
	if logOp.Error != nil {
		return op.New(d.ID(), logOp.Result, logOp.Status, logOp.Error, runner.Params(d), startTime, time.Now())
	}

	return op.New(d.ID(), logOp.Result, op.Success, nil, runner.Params(d), startTime, time.Now())
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
