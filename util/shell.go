// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package util

import (
	"errors"
	"os"
	"runtime"
)

// ErrUnknownShell occurs when the $SHELL environment variable is empty or not set.
var ErrUnknownShell = errors.New("unable to determine shell from environment. please set the $SHELL environment variable")

// ErrNoShellOnWindows occurs when GetShell() is run on a Windows machine.
var ErrNoShellOnWindows = errors.New("shell{} is not supported on Windows. please use command{}")

// GetShell returns the value of the $SHELL environment variable,
// or an error if on Windows or the variable is not set.
func GetShell() (string, error) {
	if runtime.GOOS == "windows" {
		return "", ErrNoShellOnWindows
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		// perhaps default to shell = "sh" in future?
		// for now, just hard fail.
		return "", ErrUnknownShell
	}

	return shell, nil
}
