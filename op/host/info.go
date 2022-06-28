package host

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
	"github.com/shirou/gopsutil/v3/host"
)

var _ op.Runner = Info{}

type Info struct{}

func (i Info) ID() string {
	return "info"
}

func (i Info) Run() op.Op {
	// third party
	hostInfo, err := host.Info()
	if err != nil {
		hclog.L().Trace("op/host.Info.Run()", "error", err)
		return op.New(i, hostInfo, op.Fail, err)
	}

	return op.New(i, hostInfo, op.Success, nil)
}
