package seeker

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
)

// JournaldGetter attempts to pull logs from journald via shell command, e.g.:
// journalctl -x -u {name} --since '3 days ago' --no-pager > {destDir}/journald-{name}.log
func JournaldGetter(name, destDir string) *Seeker {
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

	// TODO(gulducat): custom --since
	cmd = fmt.Sprintf("journalctl -x -u %s --since '3 days ago' --no-pager > %s/journald-%s.log", name, destDir, name)
	seeker := NewSheller(cmd)
	seeker.Identifier = "journald"
	return seeker
}
