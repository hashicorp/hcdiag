package host

import (
	"fmt"
	"time"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/hcdiag/runner"
	"github.com/shirou/gopsutil/v3/disk"
)

// Partition includes details about a disk partition. This serves as the basis for the results produced by
// the Disk runner.
type Partition struct {
	Device     string   `json:"device"`
	Mountpoint string   `json:"mountpoint"`
	Fstype     string   `json:"fstype"`
	Opts       []string `json:"opts"`
}

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
	var partitions []Partition
	startTime := time.Now()

	dp, err := disk.Partitions(true)
	if err != nil {
		hclog.L().Trace("runner/host.Disk.Run()", "error", err)
		err1 := fmt.Errorf("error getting disk information err=%w", err)
		return op.New(d.ID(), partitions, op.Unknown, err1, nil, startTime, time.Now())
	}

	partitions, err = d.partitions(dp)
	if err != nil {
		hclog.L().Trace("runner/host.Disk.Run() failed to convert partition info", "error", err)
		err1 := fmt.Errorf("error converting partition information err=%w", err)
		return op.New(d.ID(), partitions, op.Fail, err1, nil, startTime, time.Now())
	}

	return op.New(d.ID(), partitions, op.Success, nil, nil, startTime, time.Now())
}

func (d Disk) partitions(dps []disk.PartitionStat) ([]Partition, error) {
	var partitions []Partition

	for _, dp := range dps {
		var partition Partition
		dev, err := redact.String(dp.Device, d.Redactions)
		if err != nil {
			return partitions, err
		}
		partition.Device = dev

		mp, err := redact.String(dp.Mountpoint, d.Redactions)
		if err != nil {
			return partitions, err
		}
		partition.Mountpoint = mp

		fst, err := redact.String(dp.Fstype, d.Redactions)
		if err != nil {
			return partitions, err
		}
		partition.Fstype = fst

		// Make a slice rather than an empty var declaration to avoid later marshalling null instead of empty array
		opts := make([]string, 0)
		for _, opt := range dp.Opts {
			redactedOpt, err := redact.String(opt, d.Redactions)
			if err != nil {
				return partitions, err
			}
			opts = append(opts, redactedOpt)
		}
		partition.Opts = opts

		partitions = append(partitions, partition)
	}

	return partitions, nil
}
