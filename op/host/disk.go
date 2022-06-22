package host

import (
	"fmt"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/hcdiag/op"
	"github.com/shirou/gopsutil/v3/disk"
)

var _ op.Runner = Disk{}

type Disk struct{}

func NewDisk() *op.Op {
	return &op.Op{
		Identifier: "disks",
		Runner:     Disk{},
	}
}

func (dp Disk) Run() (interface{}, op.Status, error) {
	// third party
	diskInfo, err := disk.Partitions(true)
	if err != nil {
		hclog.L().Trace("op/host.Disk.Run()", "error", err)
		return diskInfo, op.Unknown, fmt.Errorf("error getting disk information err=%w", err)
	}

	return diskInfo, op.Success, nil
}
