package host

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/util"
	"github.com/shirou/gopsutil/v3/mem"
)

var _ op.Runner = Memory{}

type Memory struct{}

func (m Memory) ID() string {
	return "memory"
}

// Run calls out to mem.VirtualMemory
func (m Memory) Run() op.Op {
	memoryInfo, err := mem.VirtualMemory()
	if err != nil {
		hclog.L().Trace("op/host.Memory.Run()", "error", err)
		return m.op(memoryInfo, op.Fail, err)
	}

	return m.op(memoryInfo, op.Success, nil)
}

func (m Memory) op(result interface{}, status op.Status, err error) op.Op {
	return op.Op{
		Identifier: m.ID(),
		Result:     result,
		Error:      err,
		Status:     status,
		Params:     util.RunnerParams(m),
	}
}
