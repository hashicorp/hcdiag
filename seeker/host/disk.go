package host

import (
	"fmt"

	"github.com/hashicorp/hcdiag/seeker"
	"github.com/shirou/gopsutil/v3/disk"
)

var _ seeker.Runner = Disks{}

type Disks struct{}

func NewDisks() *seeker.Seeker {
	return &seeker.Seeker{
		Identifier: "disks",
		Runner:     Disks{},
	}
}

func (dp Disks) Run() (interface{}, seeker.Status, error) {
	// third party
	diskInfo, err := disk.Partitions(true)
	if err != nil {
		return diskInfo, seeker.Unknown, fmt.Errorf("error getting disk information err=%w", err)
	}

	return diskInfo, seeker.Success, nil
}
