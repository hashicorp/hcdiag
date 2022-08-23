package host

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/runner"
	"github.com/shirou/gopsutil/v3/net"
)

var _ runner.Runner = &Network{}

// NetworkInterface represents details about a network interface. This serves as the basis for the results produced
// by the Network runner.
type NetworkInterface struct {
	Index int      `json:"index"`
	MTU   int      `json:"mtu"`
	Name  string   `json:"name"`
	Flags []string `json:"flags"`
	Addrs []string `json:"addrs"`
}

type Network struct {
	Redactions []*redact.Redact
}

func NewNetwork(redactions []*redact.Redact) *Network {
	return &Network{
		Redactions: redactions,
	}
}

func (n Network) ID() string {
	return "network"
}

func (n Network) Run() op.Op {
	var interfaces []NetworkInterface
	netIfs, err := net.Interfaces()
	if err != nil {
		hclog.L().Trace("runner/host.Network.Run()", "error", err)
		return op.New(n.ID(), interfaces, op.Fail, err, nil)
	}

	for _, netIf := range netIfs {
		ifce, err := n.networkInterface(netIf)
		if err != nil {
			hclog.L().Trace("runner/host.Network.Run()", "error", err)
			err1 := fmt.Errorf("error converting network information err=%w", err)
			return op.New(n.ID(), interfaces, op.Fail, err1, nil)
		}
		interfaces = append(interfaces, ifce)
	}

	return op.New(n.ID(), interfaces, op.Success, nil, nil)
}

func (n Network) networkInterface(nis net.InterfaceStat) (NetworkInterface, error) {
	netIf := NetworkInterface{
		Index: nis.Index,
		MTU:   nis.MTU,
	}

	name, err := redact.String(nis.Name, n.Redactions)
	if err != nil {
		return NetworkInterface{}, err
	}
	netIf.Name = name

	// Make a slice rather than an empty var declaration to avoid later marshalling null instead of empty array
	flags := make([]string, 0)
	for _, f := range nis.Flags {
		flag, err := redact.String(f, n.Redactions)
		if err != nil {
			return NetworkInterface{}, err
		}
		flags = append(flags, flag)
	}
	netIf.Flags = flags

	// Make a slice rather than an empty var declaration to avoid later marshalling null instead of empty array
	addrs := make([]string, 0)
	for _, a := range nis.Addrs {
		addr, err := redact.String(a.Addr, n.Redactions)
		if err != nil {
			return NetworkInterface{}, err
		}
		addrs = append(addrs, addr)
	}
	netIf.Addrs = addrs

	return netIf, nil
}
