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
		return i.op(hostInfo, op.Fail, err)
	}

	return i.op(hostInfo, op.Success, nil)
}
func (i Info) op(result interface{}, status op.Status, err error) op.Op {
	return op.Op{
		Identifier: i.ID(),
		Result:     result,
		Error:      err,
		Status:     status,
	}
}
