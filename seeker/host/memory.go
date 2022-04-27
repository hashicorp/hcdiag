package host

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/seeker"
	"github.com/shirou/gopsutil/v3/mem"
)

var _ seeker.Runner = Memory{}

type Memory struct{}

func NewMemory() *seeker.Seeker {
	return &seeker.Seeker{
		Identifier: "memory",
		Runner:     Memory{},
	}
}

// Run calls out to mem.VirtualMemory and returns it for results
func (m Memory) Run() (interface{}, seeker.Status, error) {
	memoryInfo, err := mem.VirtualMemory()
	if err != nil {
		hclog.L().Error("GetMemory", "Error getting memory information", err)
		return memoryInfo, seeker.Fail, err
	}

	return memoryInfo, seeker.Success, nil
}
