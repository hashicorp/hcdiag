// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package debug

import (
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/hcdiag/redact"
	"github.com/stretchr/testify/assert"
)

func TestVaultCmdString(t *testing.T) {
	tcs := []struct {
		name            string
		cfg             VaultDebugConfig
		productDuration time.Duration
		productInterval time.Duration
		filterString    string
		expected        string
	}{
		{
			name: "product config defaults should be used when no configuration is passed in (duration and interval)",
			cfg: VaultDebugConfig{
				Compress:        "false",
				LogFormat:       "standard",
				MetricsInterval: "10s",
				Targets:         []string{},
				Redactions:      []*redact.Redact{},
			},
			productDuration: 5 * time.Minute,
			productInterval: 45 * time.Second,
			expected:        "vault debug -compress=false -duration=5m0s -interval=45s -log-format=standard -metrics-interval=10s -output=/tmp/hcdiag/VaultDebug123/VaultDebug*",
		},
		{
			name: "config values should override product config defaults (compression, duration, and interval)",
			cfg: VaultDebugConfig{
				Compress:        "false",
				Duration:        "3m",
				Interval:        "30s",
				LogFormat:       "standard",
				MetricsInterval: "10s",
				Targets:         []string{},
				Redactions:      []*redact.Redact{},
			},
			productDuration: 2 * time.Minute,
			productInterval: 20 * time.Second,
			expected:        "vault debug -compress=false -duration=3m -interval=30s -log-format=standard -metrics-interval=10s -output=/tmp/hcdiag/VaultDebug123/VaultDebug*",
		},
		{
			name: "Internal defaults should be used when not present in configuration (compress, logformat)",
			cfg: VaultDebugConfig{
				Duration:        "3m",
				Interval:        "30s",
				MetricsInterval: "10s",
				Targets:         []string{},
				Redactions:      []*redact.Redact{},
			},
			productDuration: 2 * time.Minute,
			productInterval: 20 * time.Second,
			expected:        "vault debug -compress=true -duration=3m -interval=30s -log-format=standard -metrics-interval=10s -output=/tmp/hcdiag/VaultDebug123/VaultDebug*.tar.gz",
		},
		{
			name:            "default config for a vaultDebug runner should make the resulting -output end with .tar.gz",
			cfg:             VaultDebugConfig{},
			productDuration: 2 * time.Minute,
			productInterval: 30 * time.Second,
			expected:        "vault debug -compress=true -duration=2m0s -interval=30s -log-format=standard -metrics-interval=10s -output=/tmp/hcdiag/VaultDebug123/VaultDebug*.tar.gz",
		},
		{
			name:            "an empty config should produce a valid VaultDebug command",
			cfg:             VaultDebugConfig{},
			productDuration: 2 * time.Minute,
			productInterval: 30 * time.Second,
			expected:        "vault debug -compress=true -duration=2m0s -interval=30s -log-format=standard -metrics-interval=10s -output=/tmp/hcdiag/VaultDebug123/VaultDebug*.tar.gz",
		},
		{
			name: "a new VaultDebug (with options) should have correct vault debug command",
			cfg: VaultDebugConfig{
				Compress:        "false",
				Duration:        "2m",
				Interval:        "30s",
				LogFormat:       "standard",
				MetricsInterval: "10s",
				Targets:         []string{"metrics", "pprof", "replication-status"},
				Redactions:      []*redact.Redact{},
			},
			productDuration: 2 * time.Minute,
			productInterval: 30 * time.Second,
			filterString:    " -target=metrics -target=pprof -target=replication-status",
			expected:        "vault debug -compress=false -duration=2m -interval=30s -log-format=standard -metrics-interval=10s -output=/tmp/hcdiag/VaultDebug123/VaultDebug* -target=metrics -target=pprof -target=replication-status",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			d, err := NewVaultDebug(tc.cfg, "/tmp", tc.productDuration, tc.productInterval)
			assert.NoError(t, err)

			cmdString := vaultCmdString(*d, tc.filterString, "/tmp/hcdiag/VaultDebug123")
			matched, _ := regexp.MatchString(tc.expected, cmdString)
			assert.True(t, matched, "got:", cmdString, "expected:", tc.expected)
		})
	}
}
