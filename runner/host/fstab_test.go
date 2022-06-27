package host

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcdiag/runner"
	"github.com/stretchr/testify/assert"
)

type mockShellRunner struct {
	result interface{}
	status runner.Status
	err    error
}

func (m mockShellRunner) Run() runner.Op {
	return runner.Op{
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
		result    interface{}
		status    runner.Status
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
				os: "windows",
			},
			expected: response{
				status:    runner.Success,
				expectErr: true,
			},
		},
		{
			name: "Test Darwin Does Not Run",
			fstab: FSTab{
				os: "darwin",
			},
			expected: response{
				status:    runner.Success,
				expectErr: true,
			},
		},
		{
			name: "Test Successful Run",
			fstab: FSTab{
				os: "linux",
				sheller: &mockShellRunner{
					result: "contents",
					status: runner.Success,
					err:    nil,
				},
			},
			expected: response{
				result:    "contents",
				status:    runner.Success,
				expectErr: false,
			},
		},
		{
			name: "Test Unsuccessful Run",
			fstab: FSTab{
				os: "linux",
				sheller: &mockShellRunner{
					result: nil,
					status: runner.Unknown,
					err:    fmt.Errorf("an error"),
				},
			},
			expected: response{
				result:    nil,
				status:    runner.Fail,
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
			expected: FSTab{os: "linux"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fstab := NewFSTab(tc.os)
			assert.Equal(t, tc.expected.os, fstab.os)
		})
	}
}
