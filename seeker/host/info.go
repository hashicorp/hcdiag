package host

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/seeker"
	"github.com/shirou/gopsutil/v3/host"
)

var _ seeker.Runner = Info{}

func NewInfo() *seeker.Seeker {
	return &seeker.Seeker{
		Identifier: "info",
		Runner:     Info{},
	}
}

type Info struct{}

func (i Info) Run() (interface{}, seeker.Status, error) {
	// third party
	hostInfo, err := host.Info()
	if err != nil {
		hclog.L().Error("GetHost", "Error getting host information", err)
		return hostInfo, seeker.Fail, err
	}

	return hostInfo, seeker.Success, nil
}
