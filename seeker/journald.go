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

	sinceFmt := since.Format(JournaldTimeLayout)
	untilFmt := until.Format(JournaldTimeLayout)
	cmd = fmt.Sprintf("journalctl -x -u %s --since \"%s\" --until \"%s\" --no-pager > %s/journald-%s.log", name, sinceFmt, untilFmt, destDir, name)
	seeker := NewSheller(cmd)
	seeker.Identifier = "journald"

	return seeker
}
