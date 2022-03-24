package seeker

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-hclog"
)

// JournaldTimeLayout custom go time layouts must match the reference time Jan 2 15:04:05 2006 MST
const JournaldTimeLayout = "2006-01-02 15:04:05"

// JournaldGetter attempts to pull logs from journald via shell command, e.g.:
// journalctl -x -u {name} --since '3 days ago' --no-pager > {destDir}/journald-{name}.log
func JournaldGetter(name, destDir string, since, until time.Time) *Seeker {
	// if systemd does not exist or have a unit with the provided name, return a nil pointer
	cmd := fmt.Sprintf("systemctl is-enabled %s", name) // TODO(gulducat): another command?
	out, err := NewCommander(cmd, "string").Run()
	if err != nil {
		hclog.L().Debug("skipping journald", "name", name, "output", out, "error", err)
		return nil
	}

	// check if user is able to read messages
	cmd = fmt.Sprintf("journalctl -n0 -u %s 2>&1 | grep -A10 'not seeing messages from other users'", name)
	out, err = NewSheller(cmd).Run()
	if err == nil { // no error, our sad magic string was found
		hclog.L().Error("journalctl -u "+name,
			"message", "unable to read logs, is your user allowed?  try sudo?",
			"output", out,
		)
		return nil
	}

	cmd = JournalctlGetLogsCmd(name, destDir, since, until)
	seeker := NewSheller(cmd)
	seeker.Identifier = "journald"

	return seeker
}

// JournalctlGetLogsCmd takes the service name, destination to write logs to, two timestamps for the time range, and
//  arranges them into a runnable command string.
func JournalctlGetLogsCmd(name, destDir string, since, until time.Time) string {
	// Write since and until flags with formatted time if either has a non-zero value
	// Flag strings handle their own trailing whitespace to avoid having extra spaces when the flag is disabled.
	var sinceFlag, untilFlag string
	if !since.IsZero() {
		t := since.Format(JournaldTimeLayout)
		sinceFlag = fmt.Sprintf("--since \"%s\" ", t)
	}
	if !until.IsZero() {
		t := until.Format(JournaldTimeLayout)
		untilFlag = fmt.Sprintf("--until \"%s\" ", t)
	}

	// Add our write destination
	dest := fmt.Sprintf("%s/journald-%s.log", destDir, name)

	// Compose the service name with flags and write to dest
	cmd := fmt.Sprintf("journalctl -x -u %s %s%s--no-pager > %s", name, sinceFlag, untilFlag, dest)
	return cmd
}
