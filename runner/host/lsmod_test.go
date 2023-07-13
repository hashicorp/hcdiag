// Copyright (c) HashiCorp, Inc.
package host

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcdiag/op"

	"github.com/hashicorp/hcdiag/runner"
	"github.com/stretchr/testify/assert"
)

type sampleShellRunner struct {
	result map[string]any
	status op.Status
	err    error
}

func (m sampleShellRunner) Run() op.Op {
	return op.Op{
		Result: m.result,
		Status: m.status,
		Error:  m.err,
	}
}

func (m sampleShellRunner) ID() string {
	return ""
}

var _ runner.Runner = sampleShellRunner{}

// Running a first test case at once
func TestLsmod_Run(t *testing.T) {
	type response struct {
		result    map[string]any
		status    op.Status
		expectErr bool
	}

	testCases := []struct {
		name     string
		lsmod    Lsmod
		expected response
	}{
		{
			name: "Testing if Windows Does Not Run",
			lsmod: Lsmod{
				OS: "windows",
			},
			expected: response{
				status:    op.Skip,
				expectErr: true,
			},
		},
		{
			name: "Testing if Darwin Does Not Run",
			lsmod: Lsmod{
				OS: "darwin",
			},
			expected: response{
				status:    op.Skip,
				expectErr: true,
			},
		},
		{
			name: "Testing a Successful Run",
			lsmod: Lsmod{
				OS: "linux",
				Shell: &sampleShellRunner{
					result: map[string]any{"shell": "contents"},
					status: op.Success,
					err:    nil,
				},
			},
			expected: response{
				result:    map[string]any{"shell": "contents"},
				status:    op.Success,
				expectErr: false,
			},
		},
		{
			name: "Testing an Unsuccessful Linux Run",
			lsmod: Lsmod{
				OS: "linux",
				Shell: &sampleShellRunner{
					status: op.Unknown,
					err:    fmt.Errorf("an error"),
				},
			},
			expected: response{
				status:    op.Unknown,
				expectErr: true,
			},
		},
	}

	//Running all the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expected := tc.expected
			o := tc.lsmod.Run()
			assert.Equal(t, expected.result, o.Result)
			assert.Equal(t, expected.status, o.Status)
			if tc.expected.expectErr {
				assert.Error(t, o.Error)
			}
		})
	}
}

// Giving a sample test to be run.
func TestNewLsmod(t *testing.T) {
	testCases := []struct {
		name     string
		os       string
		expected Lsmod
	}{
		{
			name:     "Test for Linux",
			os:       "linux",
			expected: Lsmod{OS: "linux"},
		},
		{
			name:     "Test for windows",
			os:       "windows",
			expected: Lsmod{OS: "windows"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lsmod, err := NewLsmod(LsmodConfig{OS: tc.os})
			assert.NoError(t, err)
			assert.Equal(t, tc.expected.OS, lsmod.OS)
		})
	}
}
