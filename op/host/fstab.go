package host

import (
	"fmt"

	"github.com/hashicorp/hcdiag/util"

	"github.com/hashicorp/hcdiag/op"
)

var _ op.Runner = FSTab{}

type FSTab struct {
	os      string
	sheller op.Runner
}

func NewFSTab(os string) *FSTab {
	return &FSTab{
		os:      os,
		sheller: op.NewSheller("cat /etc/fstab"),
	}
}

func (r FSTab) ID() string {
	return "/etc/fstab"
}

func (r FSTab) Run() op.Op {
	// Only Linux is supported currently; Windows is unsupported, and Darwin doesn't use /etc/fstab by default.
	if r.os != "linux" {
		// TODO(nwchandler): This should be op.Status("skip") once we implement it
		return r.op(nil, op.Success, fmt.Errorf("FSTab.Run() not available on os, os=%s", r.os))
	}
	o := r.sheller.Run()
	if o.Error != nil {
		return r.op(o.Result, op.Fail, o.Error)
	}

	return r.op(o.Result, op.Success, nil)
}

func (r FSTab) op(result interface{}, status op.Status, err error) op.Op {
	return op.Op{
		Identifier: r.ID(),
		Result:     result,
		Error:      err,
		Status:     status,
		Params:     util.RunnerParams(r),
	}
}
