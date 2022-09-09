package host

import (
	"fmt"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = FSTab{}

type FSTab struct {
	OS      string        `json:"os"`
	Sheller runner.Runner `json:"sheller"`
}

func NewFSTab(os string, redactions []*redact.Redact) *FSTab {
	return &FSTab{
		OS:      os,
		Sheller: runner.NewSheller("cat /etc/fstab", redactions),
	}
}

func (r FSTab) ID() string {
	return "/etc/fstab"
}

func (r FSTab) Run() []op.Op {
	opList := make([]op.Op, 0)

	// Only Linux is supported currently; Windows is unsupported, and Darwin doesn't use /etc/fstab by default.
	if r.OS != "linux" {
		return append(opList, op.New(r.ID(), nil, op.Skip, fmt.Errorf("FSTab.Run() not available on os, os=%s", r.OS), runner.Params(r)))
	}
	o := r.Sheller.Run()
	first := o[0]
	if first.Error != nil {
		return append(opList, op.New(r.ID(), first.Result, op.Fail, first.Error, runner.Params(r)))
	}

	return append(opList, op.New(r.ID(), first.Result, op.Success, nil, runner.Params(r)))
}
