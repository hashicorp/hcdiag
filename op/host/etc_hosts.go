package host

import (
	"fmt"
	"runtime"

	"github.com/hashicorp/hcdiag/util"

	"github.com/hashicorp/hcdiag/op"
)

var _ op.Runner = EtcHosts{}

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

func (r EtcHosts) Run() op.Op {
	// Not compatible with windows
	if r.os == "windows" {
		// TODO(mkcp): This should be op.Status("skip") once we implement it
		err := fmt.Errorf(" EtcHosts.Run() not available on os, os=%s", r.os)
		return r.op(nil, op.Success, err)
	}
	s := op.NewSheller("cat /etc/hosts").Run()
	if s.Error != nil {
		return r.op(s.Result, op.Fail, s.Error)
	}
	return r.op(s.Result, op.Success, nil)
}

func (r EtcHosts) op(result interface{}, status op.Status, err error) op.Op {
	return op.Op{
		Identifier: r.ID(),
		Result:     result,
		Error:      err,
		ErrString:  err.Error(),
		Status:     status,
		Params:     util.RunnerParams(r),
	}
}
