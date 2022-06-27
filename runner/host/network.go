package host

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/runner"
	"github.com/shirou/gopsutil/v3/net"
)

var _ runner.Runner = &Network{}

type Network struct{}

func (n Network) ID() string {
	return "network"
}

func (n Network) Run() runner.Op {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		hclog.L().Trace("runner/host.Network.Run()", "error", err)
		return runner.New(n, nil, runner.Fail, err)
	}

	return runner.New(n, netInterfaces, runner.Success, nil)
}
