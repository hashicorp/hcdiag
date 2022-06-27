package log

import (
	"fmt"
	"time"

	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = Docker{}

// NewDocker returns a runner with an identifier and fully configured docker runner
func NewDocker(container, destDir string, since time.Time) *Docker {
	return &Docker{
		Container: container,
		DestDir:   destDir,
		Since:     since,
	}
}

// Docker allows logs to be retrieved for a docker container
type Docker struct {
	// Container is the name of the docker container to get logs from
	Container string
	// DestDir is the directory we will write the logs to
	DestDir string
	// Since marks the beginning of the time range to include logs
	Since time.Time
}

func (d Docker) ID() string {
	return "log/docker " + d.Container
}

// Run executes the runner
func (d Docker) Run() runner.Op {
	// Check that docker exists
	o := runner.NewSheller("docker version").Run()
	if o.Error != nil {
		return runner.New(d, o.Result, runner.Fail, DockerNotFoundError{
			container: d.Container,
			err:       o.Error,
		})
	}

	// Retrieve logs
	cmd := DockerLogCmd(d.Container, d.DestDir, d.Since)
	o = runner.NewSheller(cmd).Run()
	// NOTE(mkcp): If the container does not exist, docker will exit non-zero and it'll surface as a ShellExecError.
	//  The result actionably states that the container wasn't found. In the future we may want to scrub the result
	//  and only return an actionable error message
	if o.Error != nil {
		return runner.New(d, o.Result, o.Status, o.Error)
	}
	if o.Result == "" {
		return runner.New(d, o.Result, runner.Unknown, DockerNoLogsError{container: d.Container})
	}

	return runner.New(d, o.Result, runner.Success, nil)
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

type DockerNoLogsError struct {
	container string
}

func (e DockerNoLogsError) Error() string {
	return fmt.Sprintf("docker container found but results were empty, container=%s", e.container)
}
