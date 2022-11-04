package debug

import (
	"testing"
	"time"

	"github.com/hashicorp/hcdiag/redact"
	"github.com/stretchr/testify/assert"
)

func TestProductFilterString(t *testing.T) {
	tcs := []struct {
		name      string
		product   string
		filters   []string
		expect    string
		expectErr bool
	}{
		{
			name:      "empty filters should produce an error 1",
			product:   "nomad",
			filters:   []string{""},
			expect:    "",
			expectErr: true,
		},
		{
			name:      "empty filters inside valid ones should produce an error",
			product:   "consul",
			filters:   []string{"members", "", "logs"},
			expect:    "",
			expectErr: true,
		},
		{
			name:      "test a valid nomad filter",
			product:   "nomad",
			filters:   []string{"ACLToken"},
			expect:    " -event-topic=ACLToken",
			expectErr: false,
		},
		{
			name:      "test valid nomad filters",
			product:   "nomad",
			filters:   []string{"ACLToken", "ACLPolicy", "ACLRole", "Job"},
			expect:    " -event-topic=ACLToken -event-topic=ACLPolicy -event-topic=ACLRole -event-topic=Job",
			expectErr: false,
		},
		{
			name:      "test an invalid consul filter",
			product:   "consul",
			filters:   []string{"floob"},
			expect:    "",
			expectErr: true,
		},
		{
			name:      "Incorrectly capitalized filters should produce an error",
			product:   "nomad",
			filters:   []string{"aclpolicy", "acltoken"},
			expect:    " -event-topic=ACLPolicy -event-topic=ACLToken",
			expectErr: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			result, err := productFilterString(tc.product, tc.filters)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expect, result)
			}
		})
	}
}

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
			filterString:    "",
			expected:        "vault debug -compress=false -duration=5m0s -interval=45s -log-format=standard -metrics-interval=10s -output=/tmp/hcdiag/VaultDebug",
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
			filterString:    "",
			expected:        "vault debug -compress=false -duration=3m -interval=30s -log-format=standard -metrics-interval=10s -output=/tmp/hcdiag/VaultDebug",
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
			filterString:    "",
			expected:        "vault debug -compress=true -duration=3m -interval=30s -log-format=standard -metrics-interval=10s -output=/tmp/hcdiag/VaultDebug.tar.gz",
		},
		{
			name:            "default config for a vaultDebug runner should make the resulting -output end with .tar.gz",
			cfg:             VaultDebugConfig{},
			productDuration: 2 * time.Minute,
			productInterval: 30 * time.Second,
			filterString:    "",
			expected:        "vault debug -compress=true -duration=2m0s -interval=30s -log-format=standard -metrics-interval=10s -output=/tmp/hcdiag/VaultDebug.tar.gz",
		},
		{
			name:            "an empty config should produce a valid VaultDebug command",
			cfg:             VaultDebugConfig{},
			productDuration: 2 * time.Minute,
			productInterval: 30 * time.Second,
			filterString:    "",
			expected:        "vault debug -compress=true -duration=2m0s -interval=30s -log-format=standard -metrics-interval=10s -output=/tmp/hcdiag/VaultDebug.tar.gz",
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
			expected:        "vault debug -compress=false -duration=2m -interval=30s -log-format=standard -metrics-interval=10s -output=/tmp/hcdiag/VaultDebug -target=metrics -target=pprof -target=replication-status",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			d := NewVaultDebug(tc.cfg, "/tmp/hcdiag", tc.productDuration, tc.productInterval)
			cmdString := vaultCmdString(*d, tc.filterString)

			if tc.expected != cmdString {
				t.Error(tc.name, cmdString)
			}
		})
	}
}
