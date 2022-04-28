package host

import (
	"fmt"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/hcdiag/seeker"
	"github.com/shirou/gopsutil/v3/disk"
)

var _ seeker.Runner = Disk{}

type Disk struct{}

func NewDisk() *seeker.Seeker {
	return &seeker.Seeker{
		Identifier: "disks",
		Runner:     Disk{},
	}
}

func (dp Disk) Run() (interface{}, seeker.Status, error) {
	// third party
	diskInfo, err := disk.Partitions(true)
	if err != nil {
		hclog.L().Trace("seeker/host.Disk.Run()", "error", err)
		return diskInfo, seeker.Unknown, fmt.Errorf("error getting disk information err=%w", err)
	}

	return diskInfo, seeker.Success, nil
}
