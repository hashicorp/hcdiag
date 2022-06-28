package host

import (
	"fmt"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/hcdiag/op"
	"github.com/shirou/gopsutil/v3/disk"
)

var _ op.Runner = Disk{}

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
		hclog.L().Trace("op/host.Disk.Run()", "error", err)
		err1 := fmt.Errorf("error getting disk information err=%w", err)
		return op.New(d, diskInfo, op.Unknown, err1)
	}

	return op.New(d, diskInfo, op.Success, nil)
}
