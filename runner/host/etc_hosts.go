package host

import (
	"fmt"
	"runtime"

	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = EtcHosts{}

type EtcHosts struct {
	os string
}

func NewEtcHosts() *EtcHosts {
	return &EtcHosts{
		os: runtime.GOOS,
	}
}

func (r EtcHosts) ID() string {
	return "/etc/hosts"
}

func (r EtcHosts) Run() runner.Op {
	// Not compatible with windows
	if r.os == "windows" {
		// TODO(mkcp): This should be op.Status("skip") once we implement it
		err := fmt.Errorf(" EtcHosts.Run() not available on os, os=%s", r.os)
		return runner.New(r, nil, runner.Success, err)
	}
	s := runner.NewSheller("cat /etc/hosts").Run()
	if s.Error != nil {
		return runner.New(r, s.Result, runner.Fail, s.Error)
	}
	return runner.New(r, s.Result, runner.Success, nil)
}
