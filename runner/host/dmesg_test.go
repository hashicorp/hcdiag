// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package host

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcdiag/op"

	"github.com/stretchr/testify/assert"
)

func TestDMesg_Run(t *testing.T) {
	type response struct {
		result    map[string]any
		status    op.Status
		expectErr bool
	}

	testCases := []struct {
		name     string
		dmesg    DMesg
		expected response
	}{
		{
			name: "Test Windows Does Not Run",
			dmesg: DMesg{
				OS: "windows",
			},
			expected: response{
				status:    op.Skip,
				expectErr: true,
			},
		},
		{
			name: "Test Successful linux Run",
			dmesg: DMesg{
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
			dmesg: DMesg{
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
			dmesg: DMesg{
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
			dmesg: DMesg{
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
			o := tc.dmesg.Run()
			assert.Equal(t, expected.result, o.Result)
			assert.Equal(t, expected.status, o.Status)
			if tc.expected.expectErr {
				assert.Error(t, o.Error)
			}
		})
	}
}

func TestNewDMesg(t *testing.T) {
	testCases := []struct {
		name     string
		os       string
		expected DMesg
	}{
		{
			name:     "Test Linux",
			os:       "linux",
			expected: DMesg{OS: "linux"},
		},
		{
			name:     "Test Darwin",
			os:       "darwin",
			expected: DMesg{OS: "darwin"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dmesg, err := NewDMesg(DMesgConfig{OS: tc.os})
			assert.NoError(t, err)
			assert.Equal(t, tc.expected.OS, dmesg.OS)
		})
	}
}
