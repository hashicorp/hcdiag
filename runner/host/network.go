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

// InterfaceAddr represents an interface address. This serves as an input into the results produced by
// the Network runner.
type InterfaceAddr struct {
	Addr string `json:"addr"`
}

// InterfaceInfo represents details about a network interface. This serves as the basis for the results produced
// by the Network runner.
type InterfaceInfo struct {
	Index        int             `json:"index"`
	MTU          int             `json:"mtu"`
	Name         string          `json:"name"`
	HardwareAddr string          `json:"hardwareAddr"`
	Flags        []string        `json:"flags"`
	Addrs        []InterfaceAddr `json:"addrs"`
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
	var interfaceInfoList []InterfaceInfo
	netInterfaces, err := net.Interfaces()
	if err != nil {
		hclog.L().Trace("runner/host.Network.Run()", "error", err)
		return op.New(n.ID(), nil, op.Fail, err, nil)
	}

	for _, netInterface := range netInterfaces {
		interfaceInfo, err := n.convertInterfaceInfo(netInterface)
		if err != nil {
			hclog.L().Trace("runner/host.Network.Run()", "error", err)
			err1 := fmt.Errorf("error converting network information err=%w", err)
			return op.New(n.ID(), interfaceInfoList, op.Fail, err1, nil)
		}
		interfaceInfoList = append(interfaceInfoList, interfaceInfo)
	}

	return op.New(n.ID(), interfaceInfoList, op.Success, nil, nil)
}

func (n Network) convertInterfaceInfo(interfaceStat net.InterfaceStat) (InterfaceInfo, error) {
	interfaceInfo := InterfaceInfo{
		Index: interfaceStat.Index,
		MTU:   interfaceStat.MTU,
	}

	name, err := redact.String(interfaceStat.Name, n.Redactions)
	if err != nil {
		return InterfaceInfo{}, err
	}
	interfaceInfo.Name = name

	hwAddr, err := redact.String(interfaceStat.HardwareAddr, n.Redactions)
	if err != nil {
		return InterfaceInfo{}, err
	}
	interfaceInfo.HardwareAddr = hwAddr

	// Make a slice rather than an empty var declaration to avoid later marshalling null instead of empty array
	flags := make([]string, 0)
	for _, f := range interfaceStat.Flags {
		flag, err := redact.String(f, n.Redactions)
		if err != nil {
			return InterfaceInfo{}, err
		}
		flags = append(flags, flag)
	}
	interfaceInfo.Flags = flags

	// Make a slice rather than an empty var declaration to avoid later marshalling null instead of empty array
	addrs := make([]InterfaceAddr, 0)
	for _, a := range interfaceStat.Addrs {
		addr, err := redact.String(a.Addr, n.Redactions)
		if err != nil {
			return InterfaceInfo{}, err
		}
		addrs = append(addrs, InterfaceAddr{Addr: addr})
	}
	interfaceInfo.Addrs = addrs

	return interfaceInfo, nil
}
