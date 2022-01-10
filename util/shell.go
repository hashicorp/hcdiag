package util

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// ErrUnknownShell occurs when the $SHELL environment variable is empty or not set.
var ErrUnknownShell = errors.New("unable to determine shell from environment")

// ErrNoShellOnWindows occurs when ShellExec() is run on a Windows machine.
var ErrNoShellOnWindows = errors.New("shell{} is not supported on Windows. please use command{}")

// ShellExec runs a command in a subshell based on the $SHELL environment variable.
// Not available on Windows, this is mainly to enable pipes "|" and file redirection ">"
// e.g. if $SHELL is "/bin/zsh" this will run
// /bin/zsh -c "command"
func ShellExec(command string) (string, error) {

	if runtime.GOOS == "windows" {
		return "", ErrNoShellOnWindows
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		// perhaps default to shell = "sh" in future?
		// for now, just hard fail.
		return "", ErrUnknownShell
	}

	// sh, bash, zsh, fish, shells all have a -c flag to run a command in a single string argument.
	// e.g. zsh -c "cool command | grep stuff > somefile"
	// there may be edge cases
	bts, err := exec.Command(shell, "-c", command).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("exec.Command error: %s", err)
	}
	return string(bts), err
}
