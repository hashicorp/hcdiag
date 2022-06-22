package host

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
	"github.com/shirou/gopsutil/v3/net"
)

var _ op.Runner = &Network{}

func NewNetwork() *op.Op {
	return &op.Op{Identifier: "network", Runner: Network{}}
}

type Network struct{}

func (n Network) Run() (interface{}, op.Status, error) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		hclog.L().Trace("op/host.Network.Run()", "error", err)
		return nil, op.Fail, err
	}

	return netInterfaces, op.Success, nil
}
