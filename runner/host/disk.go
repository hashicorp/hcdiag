package host

import (
	"fmt"

	"github.com/hashicorp/hcdiag/op"

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

func (d Disk) Run() op.Op {
	// third party
	diskInfo, err := disk.Partitions(true)
	if err != nil {
		hclog.L().Trace("runner/host.Disk.Run()", "error", err)
		err1 := fmt.Errorf("error getting disk information err=%w", err)
		return op.New(d.ID(), diskInfo, op.Unknown, err1, nil)
	}

	return op.New(d.ID(), diskInfo, op.Success, nil, nil)
}
