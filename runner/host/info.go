package host

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/runner"
	"github.com/shirou/gopsutil/v3/host"
)

var _ runner.Runner = Info{}

type Info struct{}

func (i Info) ID() string {
	return "info"
}

func (i Info) Run() runner.Op {
	// third party
	hostInfo, err := host.Info()
	if err != nil {
		hclog.L().Trace("runner/host.Info.Run()", "error", err)
		return runner.New(i, hostInfo, runner.Fail, err)
	}

	return runner.New(i, hostInfo, runner.Success, nil)
}
