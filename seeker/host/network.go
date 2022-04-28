package host

import (
	"net"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/seeker"
)

var _ seeker.Runner = &Network{}

func NewNetwork() *seeker.Seeker {
	return &seeker.Seeker{Identifier: "network", Runner: Network{}}
}

type Network struct{}

func (n Network) Run() (interface{}, seeker.Status, error) {
	networkInfo, err := net.Interfaces()
	if err != nil {
		hclog.L().Trace("seeker/host.Network.Run()", "error", err)
		return networkInfo, seeker.Fail, err
	}

	return networkInfo, seeker.Success, err
}
