package log

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJournalctlGetLogsCmd(t *testing.T) {
	testTable := []struct {
		desc    string
		name    string
		destDir string
		since   time.Time
		until   time.Time
		expect  string
	}{
		{
			desc:    "No since or until",
			name:    "nomad",
			destDir: "/testing/test",
			since:   time.Time{},
			until:   time.Time{},
			expect:  "journalctl -x -u nomad --no-pager > /testing/test/journald-nomad.log",
		},
		{
			desc:    "With since but not until",
			name:    "nomad",
			destDir: "/testing/test",
			since:   time.Date(1, 1, 1, 1, 1, 1, 1, &time.Location{}),
			until:   time.Time{},
			expect:  "journalctl -x -u nomad --since \"0001-01-01 01:01:01\" --no-pager > /testing/test/journald-nomad.log",
		},
		{
			desc:    "With until but not since",
			name:    "nomad",
			destDir: "/testing/test",
			since:   time.Time{},
			until:   time.Date(2, 1, 1, 1, 1, 1, 1, &time.Location{}),
			expect:  "journalctl -x -u nomad --until \"0002-01-01 01:01:01\" --no-pager > /testing/test/journald-nomad.log",
		},
		{
			desc:    "with since and until",
			name:    "nomad",
			destDir: "/testing/test",
			since:   time.Date(1, 1, 1, 1, 1, 1, 1, &time.Location{}),
			until:   time.Date(2, 1, 1, 1, 1, 1, 1, &time.Location{}),
			expect:  "journalctl -x -u nomad --since \"0001-01-01 01:01:01\" --until \"0002-01-01 01:01:01\" --no-pager > /testing/test/journald-nomad.log",
		},
	}

	for _, tc := range testTable {
		result := JournalctlGetLogsCmd(tc.name, tc.destDir, tc.since, tc.until)
		assert.Equal(t, tc.expect, result)
	}
}
