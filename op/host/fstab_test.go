package host

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcdiag/op"
	"github.com/stretchr/testify/assert"
)

type mockShellRunner struct {
	result interface{}
	status op.Status
	err    error
}

func (m mockShellRunner) Run() (interface{}, op.Status, error) {
	return m.result, m.status, m.err
}

var _ op.Runner = mockShellRunner{}

func TestFSTab_Run(t *testing.T) {
	type response struct {
		result    interface{}
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
				os: "windows",
			},
			expected: response{
				status:    op.Success,
				expectErr: true,
			},
		},
		{
			name: "Test Darwin Does Not Run",
			fstab: FSTab{
				os: "darwin",
			},
			expected: response{
				status:    op.Success,
				expectErr: true,
			},
		},
		{
			name: "Test Successful Run",
			fstab: FSTab{
				os: "linux",
				sheller: &op.Op{
					Runner: mockShellRunner{
						result: "contents",
						status: op.Success,
						err:    nil,
					},
				},
			},
			expected: response{
				result:    "contents",
				status:    op.Success,
				expectErr: false,
			},
		},
		{
			name: "Test Unsuccessful Run",
			fstab: FSTab{
				os: "linux",
				sheller: &op.Op{
					Runner: mockShellRunner{
						result: nil,
						status: op.Unknown,
						err:    fmt.Errorf("an error"),
					},
				},
			},
			expected: response{
				result:    nil,
				status:    op.Fail,
				expectErr: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expected := tc.expected
			result, status, err := tc.fstab.Run()
			assert.Equal(t, expected.result, result)
			assert.Equal(t, expected.status, status)
			if tc.expected.expectErr {
				assert.Error(t, err)
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
			runner := fstab.Runner.(FSTab)
			assert.Equal(t, tc.expected.os, runner.os)
		})
	}
}
