package host

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcdiag/seeker"
	"github.com/stretchr/testify/assert"
)

type mockShellRunner struct {
	result interface{}
	status seeker.Status
	err    error
}

func (m mockShellRunner) Run() (interface{}, seeker.Status, error) {
	return m.result, m.status, m.err
}

var _ seeker.Runner = mockShellRunner{}

func TestFstab_Run(t *testing.T) {
	type response struct {
		result interface{}
		status seeker.Status
		err    error
	}

	testCases := []struct {
		name     string
		fstab    Fstab
		expected response
	}{
		{
			name: "Test Windows Does Not Run",
			fstab: Fstab{
				os: "windows",
			},
			expected: response{
				status: seeker.Success,
				err:    fmt.Errorf("Fstab.Run() not available on os, os=windows"),
			},
		},
		{
			name: "Test Darwin Does Not Run",
			fstab: Fstab{
				os: "darwin",
			},
			expected: response{
				status: seeker.Success,
				err:    fmt.Errorf("Fstab.Run() not available on os, os=darwin"),
			},
		},
		{
			name: "Test Successful Run",
			fstab: Fstab{
				os: "linux",
				sheller: &seeker.Seeker{
					Runner: mockShellRunner{
						result: "contents",
						status: seeker.Success,
						err:    nil,
					},
				},
			},
			expected: response{
				result: "contents",
				status: seeker.Success,
				err:    nil,
			},
		},
		{
			name: "Test Unsuccessful Run",
			fstab: Fstab{
				os: "linux",
				sheller: &seeker.Seeker{
					Runner: mockShellRunner{
						result: nil,
						status: seeker.Unknown,
						err:    fmt.Errorf("an error"),
					},
				},
			},
			expected: response{
				result: nil,
				status: seeker.Fail,
				err:    fmt.Errorf("an error"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expected := tc.expected
			result, status, err := tc.fstab.Run()
			if expected.result != nil {
				assert.Equal(t, expected.result, result)
			}
			assert.Equal(t, expected.status, status)
			assert.Equal(t, expected.err, err)
		})
	}
}
