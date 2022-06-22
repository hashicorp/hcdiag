package host

import (
	"fmt"
	"runtime"

	"github.com/hashicorp/hcdiag/op"
)

var _ op.Runner = EtcHosts{}

type EtcHosts struct {
	os string
}

func NewEtcHosts() *op.Op {
	return &op.Op{
		Identifier: "/etc/hosts",
		Runner:     EtcHosts{os: runtime.GOOS},
	}
}

func (s EtcHosts) Run() (interface{}, op.Status, error) {
	// Not compatible with windows
	if s.os == "windows" {
		// TODO(mkcp): This should be op.Status("skip") once we implement it
		return nil, op.Success, fmt.Errorf(" EtcHosts.Run() not available on os, os=%s", s.os)
	}
	res, _, err := op.NewSheller("cat /etc/hosts").Runner.Run()
	if err != nil {
		return res, op.Fail, err
	}
	return res, op.Success, nil
}
