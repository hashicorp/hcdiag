package debug

import (
	"testing"
	"time"

	"github.com/hashicorp/hcdiag/product"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/stretchr/testify/assert"
)

func TestSimpleDebug(t *testing.T) {
	tcs := []struct {
		name    string
		cfg     product.Config
		filters []string
		expect  string
	}{
		{
			name: "nomad config should produce valid command",
			cfg: product.Config{
				Name:          "nomad",
				TmpDir:        "/tmp/hcdiag",
				DebugDuration: 2 * time.Minute,
				DebugInterval: 30 * time.Second,
			},
			filters: []string{"Allocation", "Job"},
			expect:  "nomad operator debug -log-level=TRACE -duration=2m0s -interval=30s -node-id=all -max-nodes=100 -output=/tmp/hcdiag/ -event-topic=Allocation -event-topic=Job",
		},
		{
			name: "vault config should produce valid command",
			cfg: product.Config{
				Name:          "vault",
				TmpDir:        "/tmp/hcdiag",
				DebugDuration: 2 * time.Minute,
				DebugInterval: 30 * time.Second,
			},
			filters: []string{"metrics", "pprof", "replication-status"},
			expect:  "vault debug -compress=true -duration=2m0s -interval=30s -output=/tmp/hcdiag/VaultDebug.tar.gz -target=metrics -target=pprof -target=replication-status",
		},
		{
			name: "consul config should produce valid command",
			cfg: product.Config{
				Name:          "consul",
				TmpDir:        "/tmp/hcdiag",
				DebugDuration: 2 * time.Minute,
				DebugInterval: 30 * time.Second,
			},
			filters: []string{"members", "metrics"},
			expect:  "consul debug -duration=2m0s -interval=30s -output=/tmp/hcdiag/ConsulDebug -capture=members -capture=metrics",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			d := NewSimpleDebug(tc.cfg, tc.filters, []*redact.Redact{})
			cmdString := d.Command.Command

			if tc.expect != cmdString {
				t.Error(tc.name, cmdString)
			}
		})
	}
}

func TestProductFilterString(t *testing.T) {
	tcs := []struct {
		name      string
		product   product.Name
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

func TestVaultDebug(t *testing.T) {
	tcs := []struct {
		name         string
		cfg          VaultDebugConfig
		filterString string
		expected     string
	}{
		{
			name: "a new VaultDebug (using defaults) should have correct vault debug command",
			cfg: VaultDebugConfig{
				ProductConfig: product.Config{
					Name:          "vault",
					TmpDir:        "/tmp/hcdiag",
					DebugDuration: 2 * time.Minute,
					DebugInterval: 30 * time.Second,
				},
				Compress:        "false",
				Duration:        "3m",
				Interval:        "30s",
				LogFormat:       "standard",
				MetricsInterval: "10s",
				Targets:         []string{},
				Redactions:      []*redact.Redact{},
			},
			filterString: "",
			expected:     "vault debug -compress=false -duration=3m -interval=30s -logformat=standard -metricsinterval=10s -output=/tmp/hcdiag/VaultDebug",
		},
		{
			name: "turning on compression should make the resulting -output end with .tar.gz",
			cfg: VaultDebugConfig{
				ProductConfig: product.Config{
					Name:          "vault",
					TmpDir:        "/tmp/hcdiag",
					DebugDuration: 2 * time.Minute,
					DebugInterval: 30 * time.Second,
				},
				Compress: "true",
			},
			filterString: "",
			expected:     "vault debug -compress=true -duration=2m -interval=30s -logformat=standard -metricsinterval=10s -output=/tmp/hcdiag/VaultDebug.tar.gz",
		},
		{
			name: "a new VaultDebug (with options) should have correct vault debug command",
			cfg: VaultDebugConfig{
				ProductConfig: product.Config{
					Name:          "vault",
					TmpDir:        "/tmp/hcdiag",
					DebugDuration: 2 * time.Minute,
					DebugInterval: 30 * time.Second,
				},
				Compress:        "false",
				Duration:        "2m",
				Interval:        "30s",
				LogFormat:       "standard",
				MetricsInterval: "10s",
				Targets:         []string{"metrics", "pprof", "replication-status"},
				Redactions:      []*redact.Redact{},
			},
			filterString: " -target=metrics -target=pprof -target=replication-status",
			expected:     "vault debug -compress=false -duration=2m -interval=30s -logformat=standard -metricsinterval=10s -output=/tmp/hcdiag/VaultDebug -target=metrics -target=pprof -target=replication-status",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			d := NewVaultDebug(tc.cfg)
			cmdString := vaultCmdString(*d, tc.filterString)

			if tc.expected != cmdString {
				t.Error(tc.name, cmdString)
			}
		})
	}
}
