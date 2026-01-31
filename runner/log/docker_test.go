// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package log

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDockerLogCmd(t *testing.T) {
	testTable := []struct {
		desc    string
		name    string
		destDir string
		since   time.Time
		until   time.Time
		expect  string
	}{
		{
			desc:    "DockerLogCmd() with since",
			name:    "with-since",
			destDir: "/testing/test",
			since:   time.Date(1, 1, 1, 1, 1, 1, 1, &time.Location{}),
			expect: "docker logs --timestamps --since 0001-01-01T01:01:01Z with-since " +
				"> /testing/test/docker-with-since.log",
		},
		{
			desc:    "DockerLogCmd() without since",
			name:    "no-since",
			destDir: "/testing/test",
			since:   time.Time{},
			expect:  "docker logs --timestamps no-since > /testing/test/docker-no-since.log",
		},
		{
			desc:    "DockerLogCmd() with until does nothing",
			name:    "ignore-until",
			destDir: "/testing/test",
			until:   time.Date(1, 1, 1, 1, 1, 1, 1, &time.Location{}),
			expect:  "docker logs --timestamps ignore-until > /testing/test/docker-ignore-until.log",
		},
	}

	for _, tc := range testTable {
		result := DockerLogCmd(tc.name, tc.destDir, tc.since)
		assert.Equal(t, tc.expect, result)
	}
}

func TestDocker_Run(t *testing.T) {
	t.Skip("TestDocker_Run not implemented yet")
}
