// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package host

import (
	"fmt"
	"runtime"
	"time"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = IPTables{}

type IPTables struct {
	Commands   []string         `json:"commands"`
	Redactions []*redact.Redact `json:"redactions"`
	OS         string           `json:"os"`
}

// NewIPTables returns a runner configured to run several iptables commands
func NewIPTables(os string, redactions []*redact.Redact) *IPTables {
	return &IPTables{
		OS: os,
		Commands: []string{
			"iptables -L -n -v",
			"iptables -L -n -v -t nat",
			"iptables -L -n -v -t mangle",
		},
		Redactions: redactions,
	}
}

func (r IPTables) ID() string {
	return "iptables"
}

func (r IPTables) Run() op.Op {
	startTime := time.Now()
	if r.OS != "linux" {
		return op.New(r.ID(), nil, op.Skip, fmt.Errorf("os not linux, skipping, os=%s", runtime.GOOS), runner.Params(r), startTime, time.Now())
	}
	result := make(map[string]any)
	for _, c := range r.Commands {
		cmdCfg := runner.CommandConfig{
			Command:    c,
			Format:     "string",
			Redactions: r.Redactions,
		}
		cmdRunner, err := runner.NewCommand(cmdCfg)
		if err != nil {
			return op.New(r.ID(), nil, op.Fail, err, runner.Params(r), startTime, time.Now())
		}
		o := cmdRunner.Run()
		result[c] = o.Result
	}
	return op.New(r.ID(), result, op.Success, nil, runner.Params(r), startTime, time.Now())
}
