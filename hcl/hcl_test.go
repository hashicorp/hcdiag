package hcl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHCL(t *testing.T) {
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
