// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package host

import (
	"fmt"
	"time"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = FSTab{}

type FSTab struct {
	OS    string        `json:"os"`
	Shell runner.Runner `json:"shell"`
}

func NewFSTab(os string, redactions []*redact.Redact) (*FSTab, error) {
	shell, err := runner.NewShell(runner.ShellConfig{
		Command:    "cat /etc/fstab",
		Redactions: redactions,
	})
	if err != nil {
		return nil, err
	}
	return &FSTab{
		OS:    os,
		Shell: shell,
	}, nil
}

func (r FSTab) ID() string {
	return "/etc/fstab"
}

func (r FSTab) Run() op.Op {
	startTime := time.Now()

	// Only Linux is supported currently; Windows is unsupported, and Darwin doesn't use /etc/fstab by default.
	if r.OS != "linux" {
		return op.New(r.ID(), nil, op.Skip, fmt.Errorf("FSTab.Run() not available on os, os=%s", r.OS), runner.Params(r), startTime, time.Now())
	}
	o := r.Shell.Run()
	if o.Error != nil {
		return op.New(r.ID(), o.Result, op.Fail, o.Error, runner.Params(r), startTime, time.Now())
	}

	return op.New(r.ID(), o.Result, op.Success, nil, runner.Params(r), startTime, time.Now())
}
