// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package log

import (
	"testing"
	"time"

	"github.com/hashicorp/hcdiag/redact"
	"github.com/stretchr/testify/assert"
)

func TestJournalctlGetLogsCmd(t *testing.T) {
	testTable := []struct {
		desc   string
		expect string
		cfg    JournaldConfig
	}{
		{
			desc: "No since or until",
			cfg: JournaldConfig{
				Service:    "nomad",
				DestDir:    "/testing/test",
				Since:      time.Time{},
				Until:      time.Time{},
				Redactions: make([]*redact.Redact, 0),
			},
			expect: "journalctl -x -u nomad --no-pager > /testing/test/journald-nomad.log",
		},
		{
			desc: "With since but not until",
			cfg: JournaldConfig{
				Service:    "nomad",
				DestDir:    "/testing/test",
				Since:      time.Date(1, 1, 1, 1, 1, 1, 1, &time.Location{}),
				Until:      time.Time{},
				Redactions: []*redact.Redact{},
			},
			expect: "journalctl -x -u nomad --since \"0001-01-01 01:01:01\" --no-pager > /testing/test/journald-nomad.log",
		},
		{
			desc: "With until but not since",
			cfg: JournaldConfig{
				Service:    "nomad",
				DestDir:    "/testing/test",
				Since:      time.Time{},
				Until:      time.Date(2, 1, 1, 1, 1, 1, 1, &time.Location{}),
				Redactions: nil,
			},
			expect: "journalctl -x -u nomad --until \"0002-01-01 01:01:01\" --no-pager > /testing/test/journald-nomad.log",
		},
		{
			desc: "with since and until",
			cfg: JournaldConfig{
				Service:    "nomad",
				DestDir:    "/testing/test",
				Since:      time.Date(1, 1, 1, 1, 1, 1, 1, &time.Location{}),
				Until:      time.Date(2, 1, 1, 1, 1, 1, 1, &time.Location{}),
				Redactions: nil,
			},
			expect: "journalctl -x -u nomad --since \"0001-01-01 01:01:01\" --until \"0002-01-01 01:01:01\" --no-pager > /testing/test/journald-nomad.log",
		},
	}

	for _, tc := range testTable {
		j := NewJournald(tc.cfg)
		result := j.LogsCmd()
		assert.Equal(t, tc.expect, result)
	}
}
