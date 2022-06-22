package host

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
	"github.com/shirou/gopsutil/v3/mem"
)

var _ op.Runner = Memory{}

type Memory struct{}

func NewMemory() *op.Op {
	return &op.Op{
		Identifier: "memory",
		Runner:     Memory{},
	}
}

// Run calls out to mem.VirtualMemory and returns it for results
func (m Memory) Run() (interface{}, op.Status, error) {
	memoryInfo, err := mem.VirtualMemory()
	if err != nil {
		hclog.L().Trace("op/host.Memory.Run()", "error", err)
		return memoryInfo, op.Fail, err
	}

	return memoryInfo, op.Success, nil
}
