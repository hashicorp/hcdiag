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
		return op.Op{
			Identifier: d.ID(),
			Result:     diskInfo,
			ErrString:  err1.Error(),
			Error:      err1,
			Status:     op.Unknown,
		}
	}

	return op.Op{
		Identifier: d.ID(),
		Result:     diskInfo,
		Status:     op.Success,
	}
}
