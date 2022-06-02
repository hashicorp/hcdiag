package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVersion(t *testing.T) {
	vi := GetVersion()
	assert.Equal(t, version, vi.Version)
	assert.Equal(t, prerelease, vi.Prerelease)
}

func TestVersionInfo_SemanticVersion(t *testing.T) {
	testCases := []struct {
		name     string
		v        Version
		expected string
	}{
		{
			name: "Test only Version",
			v: Version{
				Version: "0.0.1",
			},
			expected: "0.0.1",
		},
		{
			name: "Test Prerelease",
			v: Version{
				Version:    "0.0.1",
				Prerelease: "test",
			},
			expected: "0.0.1-test",
		},
		{
			name: "Test Metadata",
			v: Version{
				Version:  "0.0.1",
				Metadata: "buildinfo",
			},
			expected: "0.0.1+buildinfo",
		},
		{
			name: "Test All",
			v: Version{
				Version:    "0.0.1",
				Prerelease: "test",
				Metadata:   "buildinfo",
			},
			expected: "0.0.1-test+buildinfo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vn := tc.v.SemanticVersion()
			assert.Equal(t, vn, tc.expected)
		})
	}
}
