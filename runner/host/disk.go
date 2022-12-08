// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package host

import (
	"context"
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

type DiskConfig struct {
	// Redactions includes any redactions to apply to the output of the runner.
	Redactions []*redact.Redact
	// Timeout specifies the amount of time that the runner should be allowed to execute before cancellation.
	Timeout time.Duration
}

type Disk struct {
	ctx context.Context

	// Redactions includes any redactions to apply to the output of the runner.
	Redactions []*redact.Redact `json:"redactions"`
	// Timeout specifies the amount of time that the runner should be allowed to execute before cancellation.
	Timeout runner.Timeout `json:"timeout"`
}

func NewDisk(cfg DiskConfig) *Disk {
	return NewDiskWithContext(context.Background(), cfg)
}

func NewDiskWithContext(ctx context.Context, cfg DiskConfig) *Disk {
	return &Disk{
		ctx:        ctx,
		Redactions: cfg.Redactions,
		Timeout:    runner.Timeout(cfg.Timeout),
	}
}

func (d Disk) ID() string {
	return "disks"
}

func (d Disk) Run() op.Op {
	startTime := time.Now()

	if d.ctx == nil {
		d.ctx = context.Background()
	}

	resChan := make(chan op.Op, 1)

	runCtx := d.ctx
	var cancel context.CancelFunc
	if 0 < d.Timeout {
		runCtx, cancel = context.WithTimeout(d.ctx, time.Duration(d.Timeout))
		defer cancel()
	}

	go func(resChan chan op.Op) {
		o := d.run()
		o.Start = startTime
		resChan <- o
	}(resChan)

	select {
	case <-runCtx.Done():
		switch runCtx.Err() {
		case context.Canceled:
			return runner.CancelOp(d, runCtx.Err(), startTime)
		case context.DeadlineExceeded:
			return runner.TimeoutOp(d, runCtx.Err(), startTime)
		default:
			return op.New(d.ID(), nil, op.Unknown, runCtx.Err(), runner.Params(d), startTime, time.Now())
		}
	case o := <-resChan:
		return o
	}
}

func (d Disk) run() op.Op {
	var partitions []Partition
	dp, err := disk.Partitions(true)
	if err != nil {
		hclog.L().Trace("runner/host.Disk.Run()", "error", err)
		err1 := fmt.Errorf("error getting disk information err=%w", err)
		result := map[string]any{"partitions": partitions}
		return op.New(d.ID(), result, op.Unknown, err1, nil, time.Time{}, time.Now())
	}
	partitions, err = d.partitions(dp)
	if err != nil {
		hclog.L().Trace("runner/host.Disk.Run() failed to convert partition info", "error", err)
		err1 := fmt.Errorf("error converting partition information err=%w", err)
		result := map[string]any{"partitions": partitions}
		return op.New(d.ID(), result, op.Fail, err1, runner.Params(d), time.Time{}, time.Now())
	}

	result := map[string]any{"partitions": partitions}
	return op.New(d.ID(), result, op.Success, nil, runner.Params(d), time.Time{}, time.Now())
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
