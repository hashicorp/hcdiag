// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package host

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/runner"
	"github.com/shirou/gopsutil/v3/host"
)

// InfoStat includes general information about the Host. It serves as the basis for the results produced
// by the Info runner.
type InfoStat struct {
	Hostname             string `json:"hostname"`
	OS                   string `json:"os"`
	Platform             string `json:"platform"`
	PlatformFamily       string `json:"platformFamily"`
	PlatformVersion      string `json:"platformVersion"`
	KernelVersion        string `json:"kernelVersion"`
	KernelArch           string `json:"kernelArch"`
	VirtualizationSystem string `json:"virtualizationSystem"`
	VirtualizationRole   string `json:"virtualizationRole"`
	HostID               string `json:"hostId"`

	Uptime   uint64 `json:"uptime"`
	BootTime uint64 `json:"bootTime"`
	Procs    uint64 `json:"procs"`
}

var _ runner.Runner = Info{}

type InfoConfig struct {
	// Redactions includes any redactions to apply to the output of the runner.
	Redactions []*redact.Redact
	// Timeout specifies the amount of time that the runner should be allowed to execute before cancellation.
	Timeout time.Duration
}

type Info struct {
	ctx context.Context

	// Redactions includes any redactions to apply to the output of the runner.
	Redactions []*redact.Redact `json:"redactions"`
	// Timeout specifies the amount of time that the runner should be allowed to execute before cancellation.
	Timeout runner.Timeout `json:"timeout"`
}

func NewInfo(cfg InfoConfig) *Info {
	return NewInfoWithContext(context.Background(), cfg)
}

func NewInfoWithContext(ctx context.Context, cfg InfoConfig) *Info {
	return &Info{
		ctx:        ctx,
		Redactions: cfg.Redactions,
		Timeout:    runner.Timeout(cfg.Timeout),
	}
}

func (i Info) ID() string {
	return "info"
}

func (i Info) Run() op.Op {
	startTime := time.Now()

	if i.ctx == nil {
		i.ctx = context.Background()
	}

	resChan := make(chan op.Op, 1)
	runCtx := i.ctx
	var cancel context.CancelFunc
	if 0 < i.Timeout {
		runCtx, cancel = context.WithTimeout(i.ctx, time.Duration(i.Timeout))
		defer cancel()
	}

	go func(resChan chan op.Op) {
		o := i.run()
		o.Start = startTime
		resChan <- o
	}(resChan)

	select {
	case <-runCtx.Done():
		switch runCtx.Err() {
		case context.Canceled:
			return runner.CancelOp(i, runCtx.Err(), startTime)
		case context.DeadlineExceeded:
			return runner.TimeoutOp(i, runCtx.Err(), startTime)
		default:
			return op.New(i.ID(), nil, op.Unknown, runCtx.Err(), runner.Params(i), startTime, time.Now())
		}
	case o := <-resChan:
		return o
	}
}

func (i Info) run() op.Op {
	// third party
	var hostInfo InfoStat

	hi, err := host.Info()
	if err != nil {
		hclog.L().Trace("runner/host.Info.Run()", "error", err)
		return op.New(i.ID(), nil, op.Fail, err, runner.Params(i), time.Time{}, time.Now())
	}

	hostInfo, err = i.infoStat(hi)
	result := map[string]any{"hostInfo": hostInfo}
	if err != nil {
		hclog.L().Trace("runner/host.Info.Run() failed to convert host info", "error", err)
		err1 := fmt.Errorf("error converting host information err=%w", err)
		return op.New(i.ID(), result, op.Fail, err1, runner.Params(i), time.Time{}, time.Now())
	}

	return op.New(i.ID(), result, op.Success, nil, runner.Params(i), time.Time{}, time.Now())
}

func (i Info) infoStat(hi *host.InfoStat) (InfoStat, error) {
	// start from the non-string values, which won't need redaction
	is := InfoStat{
		Uptime:   hi.Uptime,
		BootTime: hi.BootTime,
		Procs:    hi.Procs,
	}

	hostname, err := redact.String(hi.Hostname, i.Redactions)
	if err != nil {
		return InfoStat{}, err
	}
	is.Hostname = hostname

	os, err := redact.String(hi.OS, i.Redactions)
	if err != nil {
		return InfoStat{}, err
	}
	is.OS = os

	platform, err := redact.String(hi.Platform, i.Redactions)
	if err != nil {
		return InfoStat{}, err
	}
	is.Platform = platform

	platformFamily, err := redact.String(hi.PlatformFamily, i.Redactions)
	if err != nil {
		return InfoStat{}, err
	}
	is.PlatformFamily = platformFamily

	platformVersion, err := redact.String(hi.PlatformVersion, i.Redactions)
	if err != nil {
		return InfoStat{}, err
	}
	is.PlatformVersion = platformVersion

	kernelVersion, err := redact.String(hi.KernelVersion, i.Redactions)
	if err != nil {
		return InfoStat{}, err
	}
	is.KernelVersion = kernelVersion

	kernelArch, err := redact.String(hi.KernelArch, i.Redactions)
	if err != nil {
		return InfoStat{}, err
	}
	is.KernelArch = kernelArch

	virtualizationSystem, err := redact.String(hi.VirtualizationSystem, i.Redactions)
	if err != nil {
		return InfoStat{}, err
	}
	is.VirtualizationSystem = virtualizationSystem

	virtualizationRole, err := redact.String(hi.VirtualizationRole, i.Redactions)
	if err != nil {
		return InfoStat{}, err
	}
	is.VirtualizationRole = virtualizationRole

	hostID, err := redact.String(hi.HostID, i.Redactions)
	if err != nil {
		return InfoStat{}, err
	}
	is.HostID = hostID

	return is, nil
}
