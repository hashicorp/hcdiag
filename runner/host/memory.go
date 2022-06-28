package host

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/runner"
	"github.com/shirou/gopsutil/v3/mem"
)

var _ runner.Runner = Memory{}

type Memory struct{}

func (m Memory) ID() string {
	return "memory"
}

// Run calls out to mem.VirtualMemory
func (m Memory) Run() op.Op {
	memoryInfo, err := mem.VirtualMemory()
	if err != nil {
		hclog.L().Trace("runner/host.Memory.Run()", "error", err)
		return op.New(m.ID(), memoryInfo, op.Fail, err, nil)
	}

	return op.New(m.ID(), memoryInfo, op.Success, nil, nil)
}
