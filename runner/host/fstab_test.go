package host

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcdiag/op"

	"github.com/hashicorp/hcdiag/runner"
	"github.com/stretchr/testify/assert"
)

type mockShellRunner struct {
	result map[string]any
	status op.Status
	err    error
}

func (m mockShellRunner) Run() op.Op {
	return op.Op{
		Result: m.result,
		Status: m.status,
		Error:  m.err,
	}
}

func (m mockShellRunner) ID() string {
	return ""
}

var _ runner.Runner = mockShellRunner{}

func TestFSTab_Run(t *testing.T) {
	type response struct {
		result    map[string]any
		status    op.Status
		expectErr bool
	}

	testCases := []struct {
		name     string
		fstab    FSTab
		expected response
	}{
		{
			name: "Test Windows Does Not Run",
			fstab: FSTab{
				OS: "windows",
			},
			expected: response{
				status:    op.Skip,
				expectErr: true,
			},
		},
		{
			name: "Test Darwin Does Not Run",
			fstab: FSTab{
				OS: "darwin",
			},
			expected: response{
				status:    op.Skip,
				expectErr: true,
			},
		},
		{
			name: "Test Successful Run",
			fstab: FSTab{
				OS: "linux",
				Sheller: &mockShellRunner{
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
			name: "Test Unsuccessful Run",
			fstab: FSTab{
				OS: "linux",
				Sheller: &mockShellRunner{
					status: op.Unknown,
					err:    fmt.Errorf("an error"),
				},
			},
			expected: response{
				status:    op.Fail,
				expectErr: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expected := tc.expected
			o := tc.fstab.Run()
			assert.Equal(t, expected.result, o.Result)
			assert.Equal(t, expected.status, o.Status)
			if tc.expected.expectErr {
				assert.Error(t, o.Error)
			}
		})
	}
}

func TestNewFSTab(t *testing.T) {
	testCases := []struct {
		name     string
		os       string
		expected FSTab
	}{
		{
			name:     "Test Linux",
			os:       "linux",
			expected: FSTab{OS: "linux"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fstab := NewFSTab(tc.os, nil)
			assert.Equal(t, tc.expected.OS, fstab.OS)
		})
	}
}
