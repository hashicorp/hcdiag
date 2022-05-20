package host

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/seeker"
	"github.com/shirou/gopsutil/v3/net"
)

var _ seeker.Runner = &Network{}

func NewNetwork() *seeker.Seeker {
	return &seeker.Seeker{Identifier: "network", Runner: Network{}}
}

type Network struct{}

func (n Network) Run() (interface{}, seeker.Status, error) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		hclog.L().Trace("seeker/host.Network.Run()", "error", err)
		return nil, seeker.Fail, err
	}

	return netInterfaces, seeker.Success, nil
}
