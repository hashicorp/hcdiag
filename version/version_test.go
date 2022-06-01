package version

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVersion(t *testing.T) {
	vi := GetVersion()
	assert.Equal(t, Version, vi.Version)
	assert.Equal(t, Prerelease, vi.Prerelease)
}

func TestVersionInfo_VersionNumber(t *testing.T) {
	testCases := []struct {
		name string
		vi   VersionInfo
	}{
		{
			name: "Test only Version",
			vi: VersionInfo{
				Version: "0.0.0",
			},
		},
		{
			name: "Test Prerelease",
			vi: VersionInfo{
				Version:    "0.0.0",
				Prerelease: "test",
			},
		},
		{
			name: "Test Metadata",
			vi: VersionInfo{
				Version:  "0.0.0",
				Metadata: "buildinfo",
			},
		},
		{
			name: "Test All",
			vi: VersionInfo{
				Version:    "0.0.0",
				Prerelease: "test",
				Metadata:   "buildinfo",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vn := tc.vi.VersionNumber()
			assert.Contains(t, vn, tc.vi.Version)
			if tc.vi.Prerelease != "" {
				assert.Contains(t, vn, fmt.Sprintf("-%s", tc.vi.Prerelease))
			}
			if tc.vi.Metadata != "" {
				assert.Contains(t, vn, fmt.Sprintf("+%s", tc.vi.Metadata))
			}
		})
	}
}
