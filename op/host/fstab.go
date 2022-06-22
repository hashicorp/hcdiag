package host

import (
	"fmt"

	"github.com/hashicorp/hcdiag/op"
)

var _ op.Runner = FSTab{}

type FSTab struct {
	os      string
	sheller *op.Op
}

func NewFSTab(os string) *op.Op {
	return &op.Op{
		Identifier: "/etc/fstab",
		Runner: FSTab{
			os:      os,
			sheller: op.NewSheller("cat /etc/fstab"),
		},
	}
}

func (s FSTab) Run() (interface{}, op.Status, error) {
	// Only Linux is supported currently; Windows is unsupported, and Darwin doesn't use /etc/fstab by default.
	if s.os != "linux" {
		// TODO(nwchandler): This should be op.Status("skip") once we implement it
		return nil, op.Success, fmt.Errorf("FSTab.Run() not available on os, os=%s", s.os)
	}
	res, _, err := s.sheller.Runner.Run()
	if err != nil {
		return res, op.Fail, err
	}

	return res, op.Success, nil
}
