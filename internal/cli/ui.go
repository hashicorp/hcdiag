// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cli

import (
	"fmt"
	"io"
)

// Ui is the interface for interacting with a CLI from a command.
type Ui interface {
	Output(string)
	Warn(string)
	Error(string)
}

// BasicUi is a Ui implementation that writes output to Writer and warnings/errors to ErrorWriter.
type BasicUi struct {
	Writer      io.Writer
	ErrorWriter io.Writer
}

func (u *BasicUi) Output(s string) {
	_, _ = fmt.Fprintln(u.Writer, s)
}

func (u *BasicUi) Warn(s string) {
	_, _ = fmt.Fprintln(u.ErrorWriter, s)
}

func (u *BasicUi) Error(s string) {
	_, _ = fmt.Fprintln(u.ErrorWriter, s)
}
