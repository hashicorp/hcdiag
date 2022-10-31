package hcl

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/runner"

	"github.com/hashicorp/hcdiag/client"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
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
					pRunners, err := BuildRunners(product, tc.tmpDir, tc.client, tc.since, tc.until, nil)
					assert.NoError(t, err)
					runners = append(runners, pRunners...)
				}
				assert.Len(t, runners, tc.expect)
			}
			if host != nil {
				hostRunners, err := BuildRunners(host, tc.tmpDir, tc.client, tc.since, tc.until, nil)
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
		runners, err := mapDockerLogs(context.Background(), tc.config, defaultDest, defaultSince, nil)
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
		name       string
		config     []JournaldLog
		expected   int
		redactions []*redact.Redact
	}{
		{
			name:       "none",
			config:     []JournaldLog{},
			expected:   0,
			redactions: make([]*redact.Redact, 0),
		},
		{
			name: "only service name",
			config: []JournaldLog{
				{Service: "testService"},
			},
			expected:   1,
			redactions: []*redact.Redact{},
		},
		{
			name: "all attrs",
			config: []JournaldLog{
				{
					Service: "testService",
					Since:   testDuration,
				},
			},
			expected:   1,
			redactions: nil,
		},
		{
			name: "multi-runners with multi-attrs",
			config: []JournaldLog{
				{
					Service: "testService",
					Since:   testDuration,
				},
				{
					Service:    "testService",
					Redactions: []Redact{},
				},
				{
					Service: "testService2",
				},
			},
			expected:   3,
			redactions: nil,
		},
	}

	for _, tc := range cases {
		runners, err := mapJournaldLogs(context.Background(), tc.config, defaultDest, defaultSince, defaultUntil, tc.redactions)
		assert.NoError(t, err)
		assert.Len(t, runners, tc.expected)
	}
}

func TestValidateRedactions(t *testing.T) {
	type testCase struct {
		name       string
		redactions []Redact
	}
	shouldPass := []testCase{
		{
			name:       "empty redactions",
			redactions: []Redact{},
		},
		// TODO(mkcp): Uncomment when we support literal validation
		// {
		// 	name: "one literal",
		// 	redactions: []Redact{
		// 		{
		// 			Label: "literal",
		// 			Match: "something",
		// 		},
		// 	},
		// },
		// {
		//	name: "many literals",
		//	redactions: []Redact{
		//		{
		//			Label: "literal",
		//			ID:    "one",
		//			Match: "something",
		//		},
		//		{
		//			Label: "literal",
		//			ID:    "two",
		//			Match: "something else",
		//		},
		//	},
		//	},
		{
			name: "one regex",
			redactions: []Redact{
				{
					Label: "regex",
					ID:    "reg1",
					Match: "just a regex",
				},
			},
		},
		{
			name: "many regexes",
			redactions: []Redact{
				{
					Label: "regex",
					ID:    "reg1",
					Match: "just a regex",
				},
				{
					Label: "regex",
					ID:    "reg2",
					Match: "/just a fancy regex/",
				},
				{
					Label: "regex",
					ID:    "reg3",
					Match: "^a very fancy (.) regex?",
				},
			},
		},
		// TODO(mkcp): Uncomment when we support literal validation
		// {
		// 	name: "both regexes and literals",
		// 	redactions: []Redact{
		// 		{
		// 			Label: "regex",
		// 			ID:    "reg",
		// 			Match: "just a regex",
		// 		},
		// 		{
		// 			Label: "literal",
		// 			ID:    "lit",
		// 			Match: "something",
		// 		},
		// 	},
		// },
	}
	shouldErr := []testCase{
		{
			name: "bad label",
			redactions: []Redact{
				{
					Label: "shouldNotMatchAnyRegexLabel",
				},
			},
		},
		{
			name: "one bad regex",
			redactions: []Redact{
				{
					Label: "regex",
					ID:    "bad-reg-perl-stuff",
					Match: "\"^/(?!/)(.*?)\"",
				},
			},
		},
		{
			name: "good and bad regexes",
			redactions: []Redact{
				{
					Label: "regex",
					ID:    "the good stuff",
					Match: "/hello/",
				},
				{
					Label: "regex",
					ID:    "bad-reg-perl-stuff",
					Match: "\"^/(?!/)(.*?)\"",
				},
			},
		},
	}
	for _, tc := range shouldPass {
		assert.NoError(t, ValidateRedactions(tc.redactions), tc)
	}
	for _, tc := range shouldErr {
		assert.Error(t, ValidateRedactions(tc.redactions), tc)
	}
}
