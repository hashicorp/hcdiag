package host

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
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
		return op.New(n, nil, op.Fail, err)
	}

	return op.New(n, netInterfaces, op.Success, nil)
}
