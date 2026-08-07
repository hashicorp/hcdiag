// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicUi_Output(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "empty string",
			input:  "",
			expect: "\n",
		},
		{
			name:   "simple message",
			input:  "hello",
			expect: "hello\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			ui := &BasicUi{Writer: &out}
			ui.Output(tc.input)
			assert.Equal(t, tc.expect, out.String())
		})
	}
}

func TestBasicUi_Warn(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "empty string",
			input:  "",
			expect: "\n",
		},
		{
			name:   "warning message",
			input:  "watch out",
			expect: "watch out\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var errBuf bytes.Buffer
			ui := &BasicUi{ErrorWriter: &errBuf}
			ui.Warn(tc.input)
			assert.Equal(t, tc.expect, errBuf.String())
		})
	}
}

func TestBasicUi_WritersAreIndependent(t *testing.T) {
	// Output goes only to Writer; Warn/Error go only to ErrorWriter — they must not bleed into each other.
	var out, errBuf bytes.Buffer
	ui := &BasicUi{Writer: &out, ErrorWriter: &errBuf}

	ui.Output("to stdout")
	ui.Warn("to stderr warn")
	ui.Error("to stderr error")

	assert.Equal(t, "to stdout\n", out.String(), "Writer should only contain Output calls")
	assert.Equal(t, "to stderr warn\nto stderr error\n", errBuf.String(), "ErrorWriter should only contain Warn/Error calls")
}

func TestBasicUi_Error(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "empty string",
			input:  "",
			expect: "\n",
		},
		{
			name:   "error message",
			input:  "something failed",
			expect: "something failed\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var errBuf bytes.Buffer
			ui := &BasicUi{ErrorWriter: &errBuf}
			ui.Error(tc.input)
			assert.Equal(t, tc.expect, errBuf.String())
		})
	}
}
