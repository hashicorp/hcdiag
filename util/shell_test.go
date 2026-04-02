// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package util

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
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

	// no $SHELL ? no shell for you.
	t.Setenv("SHELL", "")
	_, err := GetShell()
	assert.ErrorIs(t, err, ErrUnknownShell)

	// happy path
	t.Setenv("SHELL", "/bin/sh")
	shell, err := GetShell()
	assert.Equal(t, "/bin/sh", shell)
	assert.NoError(t, err)
}
