// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package host

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/hcdiag/redact"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/stretchr/testify/require"
)

func TestDisk_partitions(t *testing.T) {
	testCases := []struct {
		name       string
		disk       Disk
		partitions []disk.PartitionStat
		expected   []Partition
		expectErr  bool
	}{
		{
			name: "Test Conversion without Redactions",
			disk: Disk{},
			partitions: []disk.PartitionStat{
				{
					Device:     "device1",
					Fstype:     "fstype1",
					Mountpoint: "/mnt/1/",
					Opts: []string{
						"opt1",
						"opt2",
						"opt3",
					},
				},
			},
			expected: []Partition{
				{
					Device:     "device1",
					Fstype:     "fstype1",
					Mountpoint: "/mnt/1/",
					Opts: []string{
						"opt1",
						"opt2",
						"opt3",
					},
				},
			},
		},
		{
			name: "Test Conversion with Redactions",
			disk: Disk{
				Redactions: createRedactionSlice(t, redact.Config{Matcher: "1"}),
			},
			partitions: []disk.PartitionStat{
				{
					Device:     "device1",
					Fstype:     "fstype1",
					Mountpoint: "/mnt/1/",
					Opts: []string{
						"opt1",
						"opt2",
						"opt3",
					},
				},
			},
			expected: []Partition{
				{
					Device:     fmt.Sprintf("device%s", redact.DefaultReplace),
					Fstype:     fmt.Sprintf("fstype%s", redact.DefaultReplace),
					Mountpoint: fmt.Sprintf("/mnt/%s/", redact.DefaultReplace),
					Opts: []string{
						fmt.Sprintf("opt%s", redact.DefaultReplace),
						"opt2",
						"opt3",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			partitions, err := tc.disk.partitions(tc.partitions)
			if tc.expectErr {
				require.Error(t, err, "an error was expected, but was not returned")
			} else {
				require.NoError(t, err)
				require.True(t, reflect.DeepEqual(tc.expected, partitions))
			}
		})
	}
}
