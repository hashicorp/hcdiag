// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package runner

import (
	"context"
	"os"
	"runtime"
	"testing"

	"github.com/hashicorp/hcdiag/op"

	"github.com/stretchr/testify/assert"
)

func TestShell(t *testing.T) {
	// only run on not-windows, and explicitly set SHELL env
	if runtime.GOOS == "windows" {
		return
	}
	curShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", curShell)
	os.Setenv("SHELL", "/bin/sh")

	// features pipe "|" and file redirection ">"
	c, err := NewShell(ShellConfig{
		Command: "echo hiii | grep hi > cooltestfile",
	})
	assert.NoError(t, err)
	defer os.Remove("cooltestfile")
	o := c.Run()
	assert.Equal(t, map[string]any{"shell": ""}, o.Result)
	assert.NoError(t, o.Error)

	bts, err := os.ReadFile("cooltestfile")
	assert.Equal(t, "hiii\n", string(bts))
	assert.NoError(t, err)
}

func TestShell_RunTimeout(t *testing.T) {
	t.Parallel()

	// Set to a short timeout, and sleep briefly to ensure it passes before we try to run the command
	ctx, cancelFunc := context.WithTimeout(context.Background(), 0)
	defer cancelFunc()

	sh, err := NewShellWithContext(ctx, ShellConfig{Command: "status-unknown-if-this-completes-lkshjajflks"})
	assert.NoError(t, err)

	result := sh.Run()
	assert.Equal(t, op.Timeout, result.Status)
	assert.ErrorIs(t, result.Error, context.DeadlineExceeded)
}

func TestShell_RunCancel(t *testing.T) {
	t.Parallel()

	// Set to a short timeout, and sleep briefly to ensure it passes before we try to run the command
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	sh, err := NewShellWithContext(ctx, ShellConfig{Command: "status-unknown-if-this-completes-lkshjajflks"})
	result := sh.Run()

	assert.NoError(t, err)
	assert.Equal(t, op.Canceled, result.Status)
	assert.ErrorIs(t, result.Error, context.Canceled)
}
