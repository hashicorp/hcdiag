package log

import (
	"fmt"
	"time"

	"github.com/hashicorp/hcdiag/seeker"

	"github.com/hashicorp/go-hclog"
)

var _ seeker.Runner = Journald{}

// JournaldTimeLayout custom go time layouts must match the reference time Jan 2 15:04:05 2006 MST
const JournaldTimeLayout = "2006-01-02 15:04:05"

type Journald struct {
	service string
	destDir string
	since   time.Time
	until   time.Time
}

// NewJournald sets the defaults for the journald runner and returns a seeker
func NewJournald(service, destDir string, since, until time.Time) *seeker.Seeker {
	return &seeker.Seeker{
		Identifier: "journald",
		Runner: Journald{
			service: service,
			destDir: destDir,
			since:   since,
			until:   until,
		},
	}
}

// Run attempts to pull logs from journald via shell command, e.g.:
// journalctl -x -u {name} --since '3 days ago' --no-pager > {destDir}/journald-{name}.log
func (j Journald) Run() (interface{}, seeker.Status, error) {
	// Check if systemd has a unit with the provided name, return a nil pointer
	cmd := fmt.Sprintf("systemctl is-enabled %s", j.service) // TODO(gulducat): another command?
	out, err := seeker.NewCommander(cmd, "string").Run()
	if err != nil {
		hclog.L().Debug("skipping journald", "service", j.service, "output", out, "error", err)
		return nil, seeker.Fail, JournaldServiceNotEnabled{
			service: j.service,
			command: cmd,
			result:  fmt.Sprintf("%s", out),
			err:     err,
		}
	}

	// check if user is able to read messages
	cmd = fmt.Sprintf("journalctl -n0 -u %s 2>&1 | grep -A10 'not seeing messages from other users'", j.service)
	out, err = seeker.NewSheller(cmd).Run()
	if err != nil {
		return out, seeker.Fail, err
	}
	// permissions error detected
	if err == nil {
		return nil, seeker.Fail, JournaldPermissionError{
			service: j.service,
			command: cmd,
			result:  fmt.Sprintf("%s", out),
			err:     err,
		}
	}

	cmd = j.LogsCmd()
	s := seeker.NewSheller(cmd)
	s.Identifier = "journald"

	return s.Runner.Run()
}

// LogsCmd arranges the params into a runnable command string.
func (j Journald) LogsCmd() string {
	// Write since and until flags with formatted time if either has a non-zero value
	// Flag strings handle their own trailing whitespace to avoid having extra spaces when the flag is disabled.
	var sinceFlag, untilFlag string
	if !j.since.IsZero() {
		t := j.since.Format(JournaldTimeLayout)
		sinceFlag = fmt.Sprintf("--since \"%s\" ", t)
	}
	if !j.until.IsZero() {
		t := j.until.Format(JournaldTimeLayout)
		untilFlag = fmt.Sprintf("--until \"%s\" ", t)
	}

	// Add our write destination
	dest := fmt.Sprintf("%s/journald-%s.log", j.destDir, j.service)

	// Compose the service name with flags and write to dest
	cmd := fmt.Sprintf("journalctl -x -u %s %s%s--no-pager > %s", j.service, sinceFlag, untilFlag, dest)
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
