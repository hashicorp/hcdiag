package hcl

import (
	"testing"
	"time"

	"github.com/hashicorp/hcdiag/runner"

	"github.com/hashicorp/hcdiag/client"

	"github.com/stretchr/testify/assert"
)

func Test_ParseHCL(t *testing.T) {
	testCases := []struct {
		name   string
		path   string
		expect HCL
	}{
		{
			name:   "Empty config is valid",
			path:   "../tests/resources/config/empty.hcl",
			expect: HCL{},
		},
		{
			name: "Host with no attributes is valid",
			path: "../tests/resources/config/host_no_ops.hcl",
			expect: HCL{
				Host: &Host{},
			},
		},
		{
			name: "Host with one of each op is valid",
			path: "../tests/resources/config/host_each_op.hcl",
			expect: HCL{
				Host: &Host{
					Commands: []Command{
						{Run: "testing", Format: "string"},
					},
					Shells: []Shell{
						{Run: "testing"},
					},
					GETs: []GET{
						{Path: "/v1/api/lol"},
					},
					Copies: []Copy{
						{Path: "./*", Since: "10h"},
					},
				},
			},
		},
		{
			name: "Host with multiple of a op type is valid",
			path: "../tests/resources/config/multi_ops.hcl",
			expect: HCL{
				Host: &Host{
					Commands: []Command{
						{
							Run:    "testing",
							Format: "string",
						},
						{
							Run:    "another one",
							Format: "string",
						},
						{
							Run:    "do a thing",
							Format: "json",
						},
					},
				},
			},
		},
		{
			name: "Config with a host and one product with everything is valid",
			path: "../tests/resources/config/config.hcl",
			expect: HCL{
				Host: &Host{
					Commands: []Command{
						{Run: "ps aux", Format: "string"},
					},
				},
				Products: []*Product{
					{
						Name: "consul",
						Commands: []Command{
							{Run: "consul version", Format: "json"},
							{Run: "consul operator raft list-peers", Format: "json"},
						},
						Shells: []Shell{
							{Run: "consul members | grep ."},
						},
						GETs: []GET{
							{Path: "/v1/api/metrics?format=prometheus"},
						},
						Copies: []Copy{
							{Path: "/another/test/log", Since: "240h"},
						},
						Excludes: []string{"consul some-awfully-long-command"},
						Selects: []string{
							"consul just this",
							"consul and this",
						},
					},
				},
			},
		},
		{
			name: "Config with multiple products is valid",
			path: "../tests/resources/config/multi_product.hcl",
			expect: HCL{
				Products: []*Product{
					{
						Name:     "consul",
						Commands: []Command{{Run: "consul version", Format: "string"}},
					},
					{
						Name:     "nomad",
						Commands: []Command{{Run: "nomad version", Format: "string"}},
					},
					{
						Name:     "vault",
						Commands: []Command{{Run: "vault version", Format: "string"}},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := Parse(tc.path)
			assert.NoError(t, err)
			assert.Equal(t, tc.expect, res, tc.name)
		})
	}
}

func TestBuildRunners(t *testing.T) {
	testCases := []struct {
		name   string
		hcl    HCL
		client *client.APIClient
		tmpDir string
		since  time.Time
		until  time.Time
		expect int
	}{
		{
			name: "contains host",
			hcl: HCL{
				Host: &Host{
					Commands: []Command{{
						Run:    "testCommand",
						Format: "string",
					}},
				},
			},
			expect: 1,
		},
		{
			name: "contains one product",
			hcl: HCL{
				Products: []*Product{
					{
						Name: "hcdiag",
						Commands: []Command{{
							Run:    "testCommand",
							Format: "string",
						}},
					},
				},
			},
			client: &client.APIClient{},
			expect: 1,
		},
		{
			name: "contains many products",
			hcl: HCL{
				Products: []*Product{
					{
						Name: "hcdiag",
						Commands: []Command{{
							Run:    "testCommand",
							Format: "string",
						}},
					},
					{
						Name: "hcdiag 2 the sequel to hcdiag",
						Commands: []Command{{
							Run:    "testCommand",
							Format: "string",
						}},
					},
				},
			},
			client: &client.APIClient{},
			expect: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			products := tc.hcl.Products
			host := tc.hcl.Host
			if 0 < len(products) {
				runners := make([]runner.Runner, 0)
				for _, product := range products {
					pRunners, err := BuildRunners(product, tc.tmpDir, tc.client, tc.since, tc.until)
					assert.NoError(t, err)
					runners = append(runners, pRunners...)
				}
				assert.Len(t, runners, tc.expect)
			}
			if host != nil {
				hostRunners, err := BuildRunners(host, tc.tmpDir, tc.client, tc.since, tc.until)
				assert.NoError(t, err)
				assert.Len(t, hostRunners, tc.expect)
			}
		})
	}
}

func TestMapDockerLogs(t *testing.T) {
	defaultDest := "/some/path"
	defaultSince := time.Now()
	testDuration := "48h"

	cases := []struct {
		name     string
		config   []DockerLog
		expected int
	}{
		{
			name:     "none",
			config:   []DockerLog{},
			expected: 0,
		},
		{
			name: "only service name",
			config: []DockerLog{
				{Container: "testService"},
			},
			expected: 1,
		},
		{
			name: "all attrs",
			config: []DockerLog{
				{
					Container: "testService",
					Since:     testDuration,
				},
			},
			expected: 1,
		},
		{
			name: "multi-runners with multi-attrs",
			config: []DockerLog{
				{
					Container: "testService",
					Since:     testDuration,
				},
				{
					Container: "testService",
				},
				{
					Container: "testService2",
				},
			},
			expected: 3,
		},
	}

	for _, tc := range cases {
		runners, err := mapDockerLogs(tc.config, defaultDest, defaultSince)
		assert.NoError(t, err)
		assert.Len(t, runners, tc.expected)
	}
}

func TestMapJournaldLogs(t *testing.T) {
	defaultDest := "/some/path"
	defaultSince := time.Now()
	defaultUntil := time.Time{}
	testDuration := "48h"

	cases := []struct {
		name     string
		config   []JournaldLog
		expected int
	}{
		{
			name:     "none",
			config:   []JournaldLog{},
			expected: 0,
		},
		{
			name: "only service name",
			config: []JournaldLog{
				{Service: "testService"},
			},
			expected: 1,
		},
		{
			name: "all attrs",
			config: []JournaldLog{
				{
					Service: "testService",
					Since:   testDuration,
				},
			},
			expected: 1,
		},
		{
			name: "multi-runners with multi-attrs",
			config: []JournaldLog{
				{
					Service: "testService",
					Since:   testDuration,
				},
				{
					Service: "testService",
				},
				{
					Service: "testService2",
				},
			},
			expected: 3,
		},
	}

	for _, tc := range cases {
		runners, err := mapJournaldLogs(tc.config, defaultDest, defaultSince, defaultUntil)
		assert.NoError(t, err)
		assert.Len(t, runners, tc.expected)
	}
}
