package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseHCL(t *testing.T) {
	testTable := []struct {
		desc   string
		path   string
		expect Config
	}{
		{
			desc:   "Empty config is valid",
			path:   "tests/resources/config/empty.hcl",
			expect: Config{},
		},
		{
			desc: "Host with no attributes is valid",
			path: "tests/resources/config/host_no_seekers.hcl",
			expect: Config{
				Host: &HostConfig{},
			},
		},
		{
			desc: "Host with one of each seeker is valid",
			path: "tests/resources/config/host_each_seeker.hcl",
			expect: Config{
				Host: &HostConfig{
					Commands: []CommandConfig{
						{Run: "testing", Format: "string"},
					},
					GETs: []GETConfig{
						{Path: "/v1/api/lol"},
					},
					Copies: []CopyConfig{
						{Path: "./*", Since: "10h"},
					},
				},
			},
		},
		{
			desc: "Host with multiple of a seeker type is valid",
			path: "tests/resources/config/multi_seekers.hcl",
			expect: Config{
				Host: &HostConfig{
					Commands: []CommandConfig{
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
			desc: "Config with a host and one product with everything is valid",
			path: "tests/resources/config/config.hcl",
			expect: Config{
				Host: &HostConfig{
					Commands: []CommandConfig{
						{Run: "ps aux", Format: "string"},
					},
					Copies: []CopyConfig{
						{Path: "/var/log/syslog", Since: ""},
					},
				},
				Product: &ProductConfig{
					Name: "consul",
					Commands: []CommandConfig{
						{Run: "consul version", Format: "json"},
						{Run: "consul operator raft list-peers", Format: "json"},
					},
					GETs: []GETConfig{
						{Path: "/v1/api/metrics?format=prometheus"},
					},
					Copies: []CopyConfig{
						{Path: "/some/test/log"},
						{Path: "/another/test/log", Since: "10d"},
					},
					Excludes: []string{"consul some-awfully-long-command"},
					Selects: []string{
						"consul just this",
						"consul and this",
					},
				},
			},
		},
	}

	for _, c := range testTable {
		res, err := ParseHCL(c.path)
		assert.NoError(t, err)
		assert.Equal(t, c.expect, res, c.desc)
	}
}
