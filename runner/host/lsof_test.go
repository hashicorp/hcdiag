// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package host

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcdiag/op"
	"github.com/stretchr/testify/assert"
)

func TestLsof_Run(t *testing.T) {
	type response struct {
		result    map[string]any
		status    op.Status
		expectErr bool
	}

	testCases := []struct {
		name     string
		lsof     Lsof
		expected response
	}{
		{
			name: "Test Windows Does Not Run",
			lsof: Lsof{
				OS: "windows",
			},
			expected: response{
				status:    op.Skip,
				expectErr: true,
			},
		},
		{
			name: "Test Successful linux Run",
			lsof: Lsof{
				OS: "linux",
				Shell: &mockShellRunner{
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
			name: "Test Unsuccessful linux Run",
			lsof: Lsof{
				OS: "linux",
				Shell: &mockShellRunner{
					status: op.Unknown,
					err:    fmt.Errorf("an error"),
				},
			},
			expected: response{
				status:    op.Unknown,
				expectErr: true,
			},
		},
		{
			name: "Test Successful darwin Run",
			lsof: Lsof{
				OS: "darwin",
				Shell: &mockShellRunner{
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
			name: "Test Unsuccessful darwin Run",
			lsof: Lsof{
				OS: "darwin",
				Shell: &mockShellRunner{
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expected := tc.expected
			o := tc.lsof.Run()
			assert.Equal(t, expected.result, o.Result)
			assert.Equal(t, expected.status, o.Status)
			if tc.expected.expectErr {
				assert.Error(t, o.Error)
			}
		})
	}
}

func TestNewLsof(t *testing.T) {
	testCases := []struct {
		name     string
		os       string
		expected Lsof
	}{
		{
			name:     "Test Linux",
			os:       "linux",
			expected: Lsof{OS: "linux"},
		},
		{
			name:     "Test Darwin",
			os:       "darwin",
			expected: Lsof{OS: "darwin"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lsof, err := NewLsof(LsofConfig{OS: tc.os})
			assert.NoError(t, err)
			assert.Equal(t, tc.expected.OS, lsof.OS)
		})
	}
}
