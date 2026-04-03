// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package util

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetShellWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		return
	}

	_, err := GetShell()
	assert.ErrorIs(t, err, ErrNoShellOnWindows)
}

func TestGetShell(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}

	curShell := os.Getenv("SHELL")
	t.Cleanup(func() {
		assert.NoError(t, os.Setenv("SHELL", curShell))
	})

	// no $SHELL ? no shell for you.
	require.NoError(t, os.Setenv("SHELL", ""))
	_, err := GetShell()
	assert.ErrorIs(t, err, ErrUnknownShell)

	// happy path
	require.NoError(t, os.Setenv("SHELL", "/bin/sh"))
	shell, err := GetShell()
	assert.Equal(t, "/bin/sh", shell)
	assert.NoError(t, err)
}
