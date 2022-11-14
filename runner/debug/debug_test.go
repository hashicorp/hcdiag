package debug

import (
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/hcdiag/redact"
	"github.com/stretchr/testify/assert"
)

func TestFilterArgs(t *testing.T) {
	tcs := []struct {
		name     string
		flagname string
		filters  []string
		expect   string
	}{
		{
			name:     "to avoid duplicating product logic, invalid filters should work",
			flagname: "floob",
			filters:  []string{"one", "two"},
			expect:   " -floob=one -floob=two",
		},
		{
			name:     "test some vault targets",
			flagname: "target",
			filters:  []string{"pprof", "metrics"},
			expect:   " -target=pprof -target=metrics",
		},
		{
			name:     "test empty filters",
			flagname: "capture",
			filters:  []string{},
			expect:   "",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			result := filterArgs(tc.flagname, tc.filters)
			assert.Equal(t, tc.expect, result)
		})
	}
}

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
			expected:        "vault debug -compress=false -duration=5m0s -interval=45s -log-format=standard -metrics-interval=10s -output=/tmp/hcdiag/VaultDebug-1ab2",
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
			expected:        "vault debug -compress=false -duration=3m -interval=30s -log-format=standard -metrics-interval=10s -output=/tmp/hcdiag/VaultDebug-1ab2",
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
			expected:        "vault debug -compress=true -duration=3m -interval=30s -log-format=standard -metrics-interval=10s -output=/tmp/hcdiag/VaultDebug-1ab2.tar.gz",
		},
		{
			name:            "default config for a vaultDebug runner should make the resulting -output end with .tar.gz",
			cfg:             VaultDebugConfig{},
			productDuration: 2 * time.Minute,
			productInterval: 30 * time.Second,
			filterString:    "",
			expected:        "vault debug -compress=true -duration=2m0s -interval=30s -log-format=standard -metrics-interval=10s -output=/tmp/hcdiag/VaultDebug-1ab2.tar.gz",
		},
		{
			name:            "an empty config should produce a valid VaultDebug command",
			cfg:             VaultDebugConfig{},
			productDuration: 2 * time.Minute,
			productInterval: 30 * time.Second,
			filterString:    "",
			expected:        "vault debug -compress=true -duration=2m0s -interval=30s -log-format=standard -metrics-interval=10s -output=/tmp/hcdiag/VaultDebug-1ab2.tar.gz",
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
			expected:        "vault debug -compress=false -duration=2m -interval=30s -log-format=standard -metrics-interval=10s -output=/tmp/hcdiag/VaultDebug-1ab2 -target=metrics -target=pprof -target=replication-status",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := "/tmp/hcdiag"
			d := NewVaultDebug(tc.cfg, tmpDir, tc.productDuration, tc.productInterval)

			// String munging to deal with random string that is added to filename
			d.output = debugOutputPath(tmpDir, "VaultDebug", "1ab2")
			if d.Compress == "true" {
				d.output = d.output + ".tar.gz"
			}
			cmdString := vaultCmdString(*d, tc.filterString)

			if tc.expected != cmdString {
				t.Error(tc.name, cmdString)
			}
		})
	}
}

func TestConsul(t *testing.T) {
	tcs := []struct {
		name            string
		cfg             ConsulDebugConfig
		productDuration time.Duration
		productInterval time.Duration
		filterString    string
		expected        string
	}{
		{
			name: "product config defaults should be used when no configuration is passed in (duration and interval)",
			cfg: ConsulDebugConfig{
				Archive:    "false",
				Captures:   []string{},
				Redactions: []*redact.Redact{},
			},
			productDuration: 5 * time.Minute,
			productInterval: 45 * time.Second,
			filterString:    "",
			expected:        "consul debug -archive=false -duration=5m0s -interval=45s -output=/tmp/ConsulDebug[0-9]*",
		},
		{
			name: "config values should override product config defaults (compression, duration, and interval)",
			cfg: ConsulDebugConfig{
				Archive:    "false",
				Duration:   "3m",
				Interval:   "30s",
				Captures:   []string{},
				Redactions: []*redact.Redact{},
			},
			productDuration: 2 * time.Minute,
			productInterval: 20 * time.Second,
			filterString:    "",
			expected:        "consul debug -archive=false -duration=3m -interval=30s -output=/tmp/ConsulDebug[0-9]*",
		},
		{
			name: "Internal defaults should be used when not present in configuration (compress, logformat)",
			cfg: ConsulDebugConfig{
				Duration:   "3m",
				Interval:   "30s",
				Captures:   []string{},
				Redactions: []*redact.Redact{},
			},
			productDuration: 2 * time.Minute,
			productInterval: 20 * time.Second,
			filterString:    "",
			expected:        "consul debug -archive=true -duration=3m -interval=30s -output=/tmp/ConsulDebug[0-9]*.tar.gz",
		},
		{
			name:            "default config for a ConsulDebug runner should make the resulting -output end with .tar.gz",
			cfg:             ConsulDebugConfig{},
			productDuration: 2 * time.Minute,
			productInterval: 30 * time.Second,
			filterString:    "",
			expected:        "consul debug -archive=true -duration=2m0s -interval=30s -output=/tmp/ConsulDebug[0-9]*.tar.gz",
		},
		{
			name:            "an empty config should produce a valid ConsulDebug command",
			cfg:             ConsulDebugConfig{},
			productDuration: 2 * time.Minute,
			productInterval: 30 * time.Second,
			filterString:    "",
			expected:        "consul debug -archive=true -duration=2m0s -interval=30s -output=/tmp/ConsulDebug[0-9]*",
		},
		{
			name: "a new ConsulDebug (with options) should have correct consul debug command",
			cfg: ConsulDebugConfig{
				Archive:    "false",
				Duration:   "2m",
				Interval:   "30s",
				Captures:   []string{"metrics", "pprof", "replication-status"},
				Redactions: []*redact.Redact{},
			},
			productDuration: 2 * time.Minute,
			productInterval: 30 * time.Second,
			filterString:    " -target=metrics -target=pprof -target=replication-status",
			expected:        "consul debug -archive=false -duration=2m -interval=30s -output=/tmp/ConsulDebug[0-9]* -target=metrics -target=pprof -target=replication-status",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			d, err := NewConsulDebug(tc.cfg, "/tmp", tc.productDuration, tc.productInterval)
			assert.NoError(t, err)

			if d.Archive == "true" {
				d.output = d.output + ".tar.gz"
			}
			cmdString := consulCmdString(*d, tc.filterString)

			matched, _ := regexp.MatchString(tc.expected, cmdString)
			assert.True(t, matched, "got:", cmdString, "expected:", tc.expected)
		})
	}
}
