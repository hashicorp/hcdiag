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
			name: "lower-cased filter values should produce correctly-cased options",
			cfg: product.Config{
				Name:          "nomad",
				TmpDir:        "/tmp/hcdiag",
				DebugDuration: 2 * time.Minute,
				DebugInterval: 30 * time.Second,
			},
			filters: []string{"allocation", "job"},
			expect:  "nomad operator debug -log-level=TRACE -duration=2m0s -interval=30s -node-id=all -max-nodes=100 -output=/tmp/hcdiag/ -event-topic=Allocation -event-topic=Job",
		},
		{
			name: "a small vault config should produce the correct vault debug command",
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
			name: "a small consul config should produce the correct consul debug command",
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
			name:      "empty filters should return empty string 1",
			product:   "nomad",
			filters:   []string{""},
			expect:    "",
			expectErr: false,
		},
		{
			name:      "empty filters should return empty string 2",
			product:   "vault",
			filters:   []string{},
			expect:    "",
			expectErr: false,
		},
		{
			name:      "empty filters inside valid ones should be ignored",
			product:   "consul",
			filters:   []string{"members", "", "logs"},
			expect:    " -capture=members -capture=logs",
			expectErr: false,
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
			filters:   []string{"ACLToken"},
			expect:    "",
			expectErr: true,
		},
		{
			name:      "Incorrectly capitalized filters should return correctly-cased versions",
			product:   "nomad",
			filters:   []string{"aclpolicy", "acltoken"},
			expect:    " -event-topic=ACLPolicy -event-topic=ACLToken",
			expectErr: false,
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

func TestIndexOf(t *testing.T) {
	tcs := []struct {
		name        string
		s           []string
		search      string
		expectFound bool
		expectIdx   int
	}{
		{
			name:        "empty slice should return false",
			s:           []string{},
			search:      "empty",
			expectFound: false,
			expectIdx:   0,
		},
		{
			name:        "empty search string should return false",
			s:           []string{"foo", "bar"},
			search:      "",
			expectFound: false,
			expectIdx:   0,
		},
		{
			name:        "non-matching search string should return false",
			s:           []string{"foo", "bar"},
			search:      "doesn'texist",
			expectFound: false,
			expectIdx:   0,
		},
		{
			name:        "matching search string should return (true, correct index) 1",
			s:           []string{"foo", "bar", "baz"},
			search:      "foo",
			expectFound: true,
			expectIdx:   0,
		},
		{
			name:        "matching search string should return (true, correct index) 2",
			s:           []string{"foo", "bar", "baz"},
			search:      "baz",
			expectFound: true,
			expectIdx:   2,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			found, idx := indexOf(tc.s, tc.search)
			assert.Equal(t, tc.expectFound, found)
			assert.Equal(t, tc.expectIdx, idx)
		})
	}
}

func TestVaultDebug(t *testing.T) {
	tcs := []struct {
		name            string
		vdb             *VaultDebug
		expectedCommand string
	}{
		{
			name: "a small vault config should produce the correct vault debug command",
			vdb: NewVaultDebug(
				product.Config{
					Name:          "vault",
					TmpDir:        "/tmp/hcdiag",
					DebugDuration: 2 * time.Minute,
					DebugInterval: 30 * time.Second,
				},
				"false",
				"2m",
				"30s",
				"standard",
				"10s",
				[]string{"metrics", "pprof", "replication-status"},
				[]*redact.Redact{},
			),
			expectedCommand: "vault debug -compress=false -duration=2m -interval=30s -logformat=standard -metricsinterval=10s -output=/tmp/hcdiag/VaultDebug -target=metrics -target=pprof -target=replication-status",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			d := tc.vdb
			cmdString := d.Command.Command

			if tc.expectedCommand != cmdString {
				t.Error(tc.name, cmdString)
			}
		})
	}
}
