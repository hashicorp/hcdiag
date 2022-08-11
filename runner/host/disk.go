package host

import (
	"fmt"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/hcdiag/runner"
	"github.com/shirou/gopsutil/v3/disk"
)

var _ runner.Runner = Disk{}

type Disk struct {
	Redactions []*redact.Redact `json:"redactions"`
}

func NewDisk(redactions []*redact.Redact) *Disk {
	return &Disk{
		Redactions: redactions,
	}
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

	if (d.Redactions != nil) && (len(d.Redactions) > 0) {
		var redactedDisk []RedactedPartitionStat
		for _, di := range diskInfo {
			rd := RedactedPartitionStat{
				Device:     redact.NewRedactedString(di.Device, d.Redactions),
				Mountpoint: redact.NewRedactedString(di.Mountpoint, d.Redactions),
				Fstype:     redact.NewRedactedString(di.Fstype, d.Redactions),
				Opts:       redact.NewRedactedStringSlice(di.Opts, d.Redactions),
			}
			redactedDisk = append(redactedDisk, rd)
		}

		return op.New(d.ID(), redactedDisk, op.Success, nil, nil)
	}

	return op.New(d.ID(), diskInfo, op.Success, nil, nil)
}

type RedactedPartitionStat struct {
	Device     redact.RedactedString   `json:"device"`
	Mountpoint redact.RedactedString   `json:"mountpoint"`
	Fstype     redact.RedactedString   `json:"fstype"`
	Opts       []redact.RedactedString `json:"opts"`
}
