package host

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
	"github.com/shirou/gopsutil/v3/host"
)

var _ op.Runner = Info{}

func NewInfo() *op.Op {
	return &op.Op{
		Identifier: "info",
		Runner:     Info{},
	}
}

type Info struct{}

func (i Info) Run() (interface{}, op.Status, error) {
	// third party
	hostInfo, err := host.Info()
	if err != nil {
		hclog.L().Trace("op/host.Info.Run()", "error", err)
		return hostInfo, op.Fail, err
	}

	return hostInfo, op.Success, nil
}
