package host

import (
	"fmt"

	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = FSTab{}

type FSTab struct {
	os      string
	sheller runner.Runner
}

func NewFSTab(os string) *FSTab {
	return &FSTab{
		os:      os,
		sheller: runner.NewSheller("cat /etc/fstab"),
	}
}

func (r FSTab) ID() string {
	return "/etc/fstab"
}

func (r FSTab) Run() runner.Op {
	// Only Linux is supported currently; Windows is unsupported, and Darwin doesn't use /etc/fstab by default.
	if r.os != "linux" {
		// TODO(nwchandler): This should be op.Status("skip") once we implement it
		return runner.New(r, nil, runner.Success, fmt.Errorf("FSTab.Run() not available on os, os=%s", r.os))
	}
	o := r.sheller.Run()
	if o.Error != nil {
		return runner.New(r, o.Result, runner.Fail, o.Error)
	}

	return runner.New(r, o.Result, runner.Success, nil)
}
