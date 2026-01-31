// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package host

import (
	"context"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = Memory{}

type Memory struct {
	ctx     context.Context
	Timeout runner.Timeout `json:"timeout"`
}

func (m Memory) ID() string {
	return "memory"
}

func NewMemory(timeout runner.Timeout) *Memory {
	return NewMemoryWithContext(context.Background(), timeout)
}

func NewMemoryWithContext(ctx context.Context, timeout runner.Timeout) *Memory {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Memory{
		ctx:     ctx,
		Timeout: timeout,
	}
}

// Run calls out to mem.VirtualMemory
func (m Memory) Run() op.Op {
	startTime := time.Now()
	resChan := make(chan op.Op, 1)

	var runCtx context.Context
	var cancel context.CancelFunc
	if 0 < m.Timeout {
		runCtx, cancel = context.WithTimeout(m.ctx, time.Duration(m.Timeout))
		defer cancel()
	}

	go func(resChan chan op.Op, start time.Time) {
		o := m.run()
		o.Start = start
		resChan <- o
	}(resChan, startTime)

	select {
	case <-runCtx.Done():
		switch runCtx.Err() {
		case context.Canceled:
			return op.New(m.ID(), nil, op.Canceled, runCtx.Err(), runner.Params(m), startTime, time.Now())
		case context.DeadlineExceeded:
			return op.New(m.ID(), nil, op.Timeout, runCtx.Err(), runner.Params(m), startTime, time.Now())
		default:
			return op.New(m.ID(), nil, op.Unknown, runCtx.Err(), runner.Params(m), startTime, time.Now())
		}
	case o := <-resChan:
		return o
	}
}

func (m Memory) run() op.Op {
	memoryInfo, err := mem.VirtualMemory()
	res := map[string]any{"memoryInfo": memoryInfo}
	if err != nil {
		hclog.L().Trace("runner/host.Memory.Run()", "error", err)
		return op.New(m.ID(), res, op.Fail, err, runner.Params(m), time.Time{}, time.Now())
	}
	return op.New(m.ID(), res, op.Success, nil, runner.Params(m), time.Time{}, time.Now())
}
