package runner

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSheller(t *testing.T) {
	// only run on not-windows, and explicitly set SHELL env
	if runtime.GOOS == "windows" {
		return
	}
	curShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", curShell)
	os.Setenv("SHELL", "/bin/sh")

	// features pipe "|" and file redirection ">"
	c := NewSheller("echo hiii | grep hi > cooltestfile", nil)
	defer os.Remove("cooltestfile")
	o := c.Run()
	assert.Equal(t, "", o[0].Result)
	assert.NoError(t, o[0].Error)

	bts, err := os.ReadFile("cooltestfile")
	assert.Equal(t, "hiii\n", string(bts))
	assert.NoError(t, err)
}
