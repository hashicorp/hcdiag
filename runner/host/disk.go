package host

import (
	"fmt"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/hcdiag/runner"
	"github.com/shirou/gopsutil/v3/disk"
)

type PartitionInfo struct {
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
	// third party
	var diskInfo []PartitionInfo

	dp, err := disk.Partitions(true)
	if err != nil {
		hclog.L().Trace("runner/host.Disk.Run()", "error", err)
		err1 := fmt.Errorf("error getting disk information err=%w", err)
		return op.New(d.ID(), diskInfo, op.Unknown, err1, nil)
	}

	diskInfo, err = d.convertPartitions(dp)
	if err != nil {
		hclog.L().Trace("runner/host.Disk.Run() failed to convert partition info", "error", err)
		err1 := fmt.Errorf("error converting partition information err=%w", err)
		return op.New(d.ID(), diskInfo, op.Fail, err1, nil)
	}

	return op.New(d.ID(), diskInfo, op.Success, nil, nil)
}

func (d Disk) convertPartitions(inputPartitions []disk.PartitionStat) ([]PartitionInfo, error) {
	var outputPartitions []PartitionInfo

	for _, inPartition := range inputPartitions {
		var outPartition PartitionInfo
		dev, err := redact.String(inPartition.Device, d.Redactions)
		if err != nil {
			return outputPartitions, err
		}
		outPartition.Device = dev

		mp, err := redact.String(inPartition.Mountpoint, d.Redactions)
		if err != nil {
			return outputPartitions, err
		}
		outPartition.Mountpoint = mp

		fst, err := redact.String(inPartition.Fstype, d.Redactions)
		if err != nil {
			return outputPartitions, err
		}
		outPartition.Fstype = fst

		var opts []string
		for _, opt := range inPartition.Opts {
			redactedOpt, err := redact.String(opt, d.Redactions)
			if err != nil {
				return outputPartitions, err
			}
			opts = append(opts, redactedOpt)
		}
		outPartition.Opts = opts

		outputPartitions = append(outputPartitions, outPartition)
	}

	return outputPartitions, nil
}
