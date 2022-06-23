package host

import (
	"fmt"
	"runtime"

	"github.com/hashicorp/hcdiag/op"
)

var _ op.Runner = EtcHosts{}

type EtcHosts struct {
	id string
	os string
}

func NewEtcHosts() *EtcHosts {
	return &EtcHosts{
		id: "/etc/hosts",
		os: runtime.GOOS,
	}
}

func (r EtcHosts) ID() string {
	return r.id
}

func (r EtcHosts) Run() op.Op {
	// Not compatible with windows
	if r.os == "windows" {
		// TODO(mkcp): This should be op.Status("skip") once we implement it
		err := fmt.Errorf(" EtcHosts.Run() not available on os, os=%s", r.os)
		return op.Op{
			Identifier: r.id,
			Result:     nil,
			Error:      err,
			ErrString:  err.Error(),
			Status:     op.Success,
			Params:     map[string]string{"host": r.os},
		}
	}
	res := op.NewSheller("cat /etc/hosts").Run()
	if res.Error != nil {
		return op.Op{
			Identifier: r.id,
			Result:     nil,
			ErrString:  res.Error.Error(),
			Error:      res.Error,
			Status:     op.Fail,
			Params:     map[string]string{"host": r.os},
		}
	}
	return op.Op{
		Identifier: r.id,
		Result:     nil,
		ErrString:  "",
		Error:      nil,
		Status:     "",
		Params:     nil,
	}
}
