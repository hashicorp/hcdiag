package host

import (
	"fmt"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/hcdiag/runner"
	"github.com/shirou/gopsutil/v3/disk"
)

var _ runner.Runner = Disk{}

type Disk struct{}

func NewDisk() *Disk {
	return &Disk{}
}

func (d Disk) ID() string {
	return "disks"
}

func (d Disk) Run() runner.Op {
	// third party
	diskInfo, err := disk.Partitions(true)
	if err != nil {
		hclog.L().Trace("runner/host.Disk.Run()", "error", err)
		err1 := fmt.Errorf("error getting disk information err=%w", err)
		return runner.New(d, diskInfo, runner.Unknown, err1)
	}

	return runner.New(d, diskInfo, runner.Success, nil)
}
