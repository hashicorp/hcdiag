package util

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShellExecWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		return
	}

	_, err := ShellExec("anything")
	assert.ErrorIs(t, err, ErrNoShellOnWindows)
}

func TestShellExec(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	curShell := os.Getenv("SHELL")
	defer func() {
		os.Setenv("SHELL", curShell)
	}()

	var out string
	var err error

	// no $SHELL ? no shell for you.
	os.Setenv("SHELL", "")
	_, err = ShellExec("anything")
	assert.ErrorIs(t, err, ErrUnknownShell)

	os.Setenv("SHELL", "/bin/sh")

	// if a pipe "|" works, other shell features will work too.
	out, err = ShellExec("echo hi | grep hi")
	assert.Equal(t, "hi\n", out)
	assert.NoError(t, err)

	// command's output should be surfaced on error.
	out, err = ShellExec("not-a-real-command")
	assert.Contains(t, out, "not found")
	assert.Error(t, err)
}
