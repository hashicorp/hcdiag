package debug

import (
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/hcdiag/redact"
	"github.com/stretchr/testify/assert"
)

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
			expected:        "consul debug -archive=false -duration=5m0s -interval=45s -output=/tmp/hcdiag/ConsulDebug123/ConsulDebug[0-9]*",
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
			expected:        "consul debug -archive=false -duration=3m -interval=30s -output=/tmp/hcdiag/ConsulDebug123/ConsulDebug[0-9]*",
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
			expected:        "consul debug -archive=true -duration=3m -interval=30s -output=/tmp/hcdiag/ConsulDebug123/ConsulDebug[0-9]*",
		},
		{
			name:            "an empty config should produce a valid ConsulDebug command",
			cfg:             ConsulDebugConfig{},
			productDuration: 2 * time.Minute,
			productInterval: 30 * time.Second,
			filterString:    "",
			expected:        "consul debug -archive=true -duration=2m0s -interval=30s -output=/tmp/hcdiag/ConsulDebug123/ConsulDebug[0-9]*",
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
			expected:        "consul debug -archive=false -duration=2m -interval=30s -output=/tmp/hcdiag/ConsulDebug123/ConsulDebug[0-9]* -target=metrics -target=pprof -target=replication-status",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			d, err := NewConsulDebug(tc.cfg, "/tmp", tc.productDuration, tc.productInterval)
			assert.NoError(t, err)

			if d.Archive == "true" {
				d.output = d.output + ".tar.gz"
			}
			cmdString := consulCmdString(*d, tc.filterString, "/tmp/hcdiag/ConsulDebug123")

			matched, _ := regexp.MatchString(tc.expected, cmdString)
			assert.True(t, matched, "got:", cmdString, "expected:", tc.expected)
		})
	}
}
