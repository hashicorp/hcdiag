package host

import (
	"fmt"

	"github.com/hashicorp/hcdiag/op"

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

func (r FSTab) Run() op.Op {
	// Only Linux is supported currently; Windows is unsupported, and Darwin doesn't use /etc/fstab by default.
	if r.os != "linux" {
		// TODO(nwchandler): This should be op.Status("skip") once we implement it
		return op.New(r, nil, op.Success, fmt.Errorf("FSTab.Run() not available on os, os=%s", r.os))
	}
	o := r.sheller.Run()
	if o.Error != nil {
		return op.New(r, o.Result, op.Fail, o.Error)
	}

	return op.New(r, o.Result, op.Success, nil)
}
