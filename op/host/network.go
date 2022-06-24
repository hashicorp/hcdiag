package host

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/util"
	"github.com/shirou/gopsutil/v3/net"
)

var _ op.Runner = &Network{}

type Network struct{}

func (n Network) ID() string {
	return "network"
}

func (n Network) Run() op.Op {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		hclog.L().Trace("op/host.Network.Run()", "error", err)
		return n.op(nil, op.Fail, err)
	}

	return n.op(netInterfaces, op.Success, nil)
}

func (n Network) op(result interface{}, status op.Status, err error) op.Op {
	return op.Op{
		Identifier: n.ID(),
		Result:     result,
		Error:      err,
		ErrString:  err.Error(),
		Status:     status,
		Params:     util.RunnerParams(n),
	}
}
