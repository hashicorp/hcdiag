package log

import (
	"fmt"
	"time"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/hcdiag/runner"

	"github.com/hashicorp/go-hclog"
)

var _ runner.Runner = Journald{}

// JournaldTimeLayout custom go time layouts must match the reference time Jan 2 15:04:05 2006 MST
const JournaldTimeLayout = "2006-01-02 15:04:05"

type Journald struct {
	Service    string           `json:"service"`
	DestDir    string           `json:"destDir"`
	Since      time.Time        `json:"since"`
	Until      time.Time        `json:"until"`
	Redactions []*redact.Redact `json:"redactions"`
}

// NewJournald sets the defaults for the journald runner
func NewJournald(service, destDir string, since, until time.Time, redactions []*redact.Redact) *Journald {
	return &Journald{
		Service:    service,
		DestDir:    destDir,
		Since:      since,
		Until:      until,
		Redactions: redactions,
	}
}

func (j Journald) ID() string {
	return "journald"
}

// Run attempts to pull logs from journald via shell command, e.g.:
// journalctl -x -u {name} --since '3 days ago' --no-pager > {destDir}/journald-{name}.log
func (j Journald) Run() op.Op {
	startTime := time.Now()

	version := runner.NewSheller("journalctl --version", j.Redactions).Run()
	if version.Error != nil {
		return op.New(j.ID(), version.Result, op.Skip, JournaldNotFound{
			service: j.Service,
			err:     version.Error,
		},
			runner.Params(j), startTime, time.Now())
	}

	// Check if systemd has a unit with the provided name
	cmd := fmt.Sprintf("systemctl is-enabled %s", j.Service)
	enabled := runner.NewCommand(cmd, "string", j.Redactions).Run()
	if enabled.Error != nil {
		hclog.L().Debug("skipping journald", "service", j.Service, "output", enabled.Result, "error", enabled.Error)
		return op.New(j.ID(), enabled.Result, op.Skip, JournaldServiceNotEnabled{
			service: j.Service,
			command: cmd,
			result:  fmt.Sprintf("%s", enabled.Result),
			err:     enabled.Error,
		},
			runner.Params(j), startTime, time.Now())
	}

	// check if user is able to read messages
	cmd = fmt.Sprintf("journalctl -n0 -u %s 2>&1 | grep -A10 'not seeing messages from other users'", j.Service)
	permissions := runner.NewSheller(cmd, j.Redactions).Run()
	// permissions error detected
	if permissions.Error == nil {
		return op.New(j.ID(), permissions.Result, op.Fail, JournaldPermissionError{
			service: j.Service,
			command: cmd,
			result:  fmt.Sprintf("%s", permissions.Result),
			err:     permissions.Error,
		},
			runner.Params(j), startTime, time.Now())
	}

	cmd = j.LogsCmd()
	logs := runner.NewSheller(cmd, j.Redactions).Run()

	return op.New(j.ID(), logs.Result, logs.Status, logs.Error, runner.Params(j), startTime, time.Now())
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
