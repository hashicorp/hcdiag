package host

import (
	"testing"

	"github.com/hashicorp/hcdiag/op"

	"github.com/stretchr/testify/assert"
)

func TestFSTab_Run(t *testing.T) {
	type response struct {
		result    map[string]any
		status    op.Status
		expectErr bool
	}

	testCases := []struct {
		name     string
		cfg      FSTabConfig
		expected response
	}{
		{
			name: "Test Windows Does Not Run",
			cfg:  FSTabConfig{OS: "windows"},
			expected: response{
				status:    op.Skip,
				expectErr: true,
			},
		},
		{
			name: "Test Darwin Does Not Run",
			cfg:  FSTabConfig{OS: "darwin"},
			expected: response{
				status:    op.Skip,
				expectErr: true,
			},
		},
		{
			name: "Test Successful Run",
			cfg:  FSTabConfig{OS: "linux"},
			expected: response{
				result:    map[string]any{"shell": ""},
				status:    op.Success,
				expectErr: false,
			},
		},
		{
			name: "Test Unsuccessful Run",
			cfg:  FSTabConfig{OS: "linux"},
			expected: response{
				status:    op.Fail,
				expectErr: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expected := tc.expected
			r := NewFSTab(tc.cfg)
			o := r.Run()
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
		cfg      FSTabConfig
		expected FSTab
	}{
		{
			name:     "Test Linux",
			cfg:      FSTabConfig{OS: "linux"},
			expected: FSTab{OS: "linux"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fstab := NewFSTab(tc.cfg)
			assert.Equal(t, tc.expected.OS, fstab.OS)
		})
	}
}
