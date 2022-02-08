package seeker

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-hclog"
)

var (
	ErrJournaldNotSeeingMessages error = errors.New("unable to read logs using journalctl; is your user allowed?  try: `sudo -i !!`")
)

// JournaldGetter attempts to pull logs from journald via shell command, e.g.:
// journalctl -xeu {name} --since '1 day ago' > {destDir}/journald-{name}.log
func JournaldGetter(name, destDir string) *Seeker {
	// if systemd does not exist or have a unit with the provided name, return a nil pointer
	cmd := fmt.Sprintf("systemctl is-active %s", name) // TODO(gulducat): another command?
	out, err := NewCommander(cmd, "string").Run()
	if err != nil {
		hclog.L().Warn("journald name not found", "name", name, "output", out, "error", err)
		return nil
	}

	/*** OPTION 1: just don't return a seeker if pre-flight check fails. ***/
	// check if user is able to read messages
	cmd = fmt.Sprintf("journalctl -n0 -u %s 2>&1 | grep -A10 'not seeing messages'", name)
	out, err = NewSheller(cmd).Run()
	if err == nil { // no error, our sad magic string was found
		Logger.Warn(out.(string))
		hclog.L().Error(ErrJournaldNotSeeingMessages.Error())
		// return nil
	}

	/*** OPTION 2: check *during* seeker runs, with a validation callback. ***/
	// try to get the logs from journald
	// ">" sends STDOUT to a log file in our bundle
	// the rest (STDERR) will land in the Seeker's Results for ResultNotContains() callback to check later.
	// TODO(gulducat): custom --since ?  other bits?
	cmd = fmt.Sprintf("journalctl -xeu %s --since '1 day ago' > %s/journald-%s.log", name, destDir, name)
	seeker := NewSheller(cmd)
	seeker.Identifier = name + " journald" // TODO(gulducat): is this sensible?

	// validate that the we were actually able to get the logs
	seeker.AddCallback(ResultNotContains(seeker,
		"not seeing messages from other users", // TODO(gulducat): is this magic string reliably consistent?
		// if this ^ is in the Results output,
		// seeker will end up with this error:
		ErrJournaldNotSeeingMessages,
	))

	return seeker
}
