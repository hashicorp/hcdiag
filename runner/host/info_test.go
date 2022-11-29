// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package host

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcdiag/redact"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/stretchr/testify/require"
)

func TestInfo_infoStat(t *testing.T) {
	testCases := []struct {
		name      string
		info      Info
		inputInfo host.InfoStat
		expected  InfoStat
		expectErr bool
	}{
		{
			name: "Test No Redactions",
			info: Info{},
			inputInfo: host.InfoStat{
				Hostname:             "host-1",
				Uptime:               1,
				BootTime:             1,
				Procs:                100,
				OS:                   "rhel",
				Platform:             "linux",
				PlatformFamily:       "linux",
				PlatformVersion:      "8.0",
				KernelVersion:        "5.0",
				KernelArch:           "amd64",
				VirtualizationSystem: "virtual-system",
				VirtualizationRole:   "virtual-role",
				HostID:               "12345",
			},
			expected: InfoStat{
				Hostname:             "host-1",
				Uptime:               1,
				BootTime:             1,
				Procs:                100,
				OS:                   "rhel",
				Platform:             "linux",
				PlatformFamily:       "linux",
				PlatformVersion:      "8.0",
				KernelVersion:        "5.0",
				KernelArch:           "amd64",
				VirtualizationSystem: "virtual-system",
				VirtualizationRole:   "virtual-role",
				HostID:               "12345",
			},
		},
		{
			name: "Test Redactions",
			info: Info{
				Redactions: createRedactionSlice(t, redact.Config{Matcher: "12345"}),
			},
			inputInfo: host.InfoStat{
				Hostname:             "host-1",
				Uptime:               12345,
				BootTime:             1,
				Procs:                100,
				OS:                   "rhel",
				Platform:             "linux",
				PlatformFamily:       "linux",
				PlatformVersion:      "8.0",
				KernelVersion:        "5.0",
				KernelArch:           "amd64",
				VirtualizationSystem: "virtual-system",
				VirtualizationRole:   "virtual-role",
				HostID:               "12345",
			},
			expected: InfoStat{
				Hostname:             "host-1",
				Uptime:               12345, // Non-strings are not redacted
				BootTime:             1,
				Procs:                100,
				OS:                   "rhel",
				Platform:             "linux",
				PlatformFamily:       "linux",
				PlatformVersion:      "8.0",
				KernelVersion:        "5.0",
				KernelArch:           "amd64",
				VirtualizationSystem: "virtual-system",
				VirtualizationRole:   "virtual-role",
				HostID:               "<REDACTED>",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info, err := tc.info.infoStat(&tc.inputInfo)
			if tc.expectErr {
				require.Error(t, err, "an error was expected, but was not returned")
			} else {
				require.NoError(t, err)
				require.True(t, reflect.DeepEqual(tc.expected, info),
					"result did not match the expected result:\nactual=%#v\nexpected=%#v", info, tc.expected)
			}
		})
	}
}
