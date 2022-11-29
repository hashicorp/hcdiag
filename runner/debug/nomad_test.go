// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package debug

import (
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/hcdiag/redact"
	"github.com/stretchr/testify/assert"
)

func TestNomadCmdString(t *testing.T) {
	tcs := []struct {
		name            string
		cfg             NomadDebugConfig
		productDuration time.Duration
		productInterval time.Duration
		filterString    string
		expected        string
	}{
		{
			name:            "product config defaults should be used when no configuration is passed in",
			cfg:             NomadDebugConfig{},
			productDuration: 5 * time.Minute,
			productInterval: 45 * time.Second,
			expected:        "nomad debug -no-color -duration=5m0s -interval=45s -log-level=TRACE -max-nodes=10 -node-id=all -pprof-duration=1s -pprof-interval=250ms -server-id=all -stale=false -verbose=false -output=/tmp/hcdiag/NomadDebug123/NomadDebug",
		},
		{
			name: "config values should override product config defaults",
			cfg: NomadDebugConfig{
				Duration:      "3m",
				Interval:      "30s",
				LogLevel:      "INFO",
				MaxNodes:      15,
				NodeID:        "my_node",
				PprofDuration: "5s",
				PprofInterval: "500ms",
				ServerID:      "my_server",
				Stale:         true,
				Verbose:       true,
				Redactions:    []*redact.Redact{},
			},
			productDuration: 2 * time.Minute,
			productInterval: 20 * time.Second,
			expected:        "nomad debug -no-color -duration=3m -interval=30s -log-level=INFO -max-nodes=15 -node-id=my_node -pprof-duration=5s -pprof-interval=500ms -server-id=my_server -stale=true -verbose=true -output=/tmp/hcdiag/NomadDebug123/NomadDebug",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			d, err := NewNomadDebug(tc.cfg, "/tmp", tc.productDuration, tc.productInterval)
			assert.NoError(t, err)

			cmdString := nomadCmdString(*d, tc.filterString, "/tmp/hcdiag/NomadDebug123")
			matched, _ := regexp.MatchString(tc.expected, cmdString)
			assert.True(t, matched, "got: %v, expected: %v", cmdString, tc.expected)
		})
	}
}
