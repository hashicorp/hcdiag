package host

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
	"github.com/mitchellh/go-ps"
)

var _ op.Runner = &Process{}

func NewProcess() *op.Op {
	return &op.Op{Identifier: "process", Runner: Process{}}
}

type Process struct{}

func (p Process) Run() (interface{}, op.Status, error) {
	processes, err := ps.Processes()
	if err != nil {
		hclog.L().Trace("op/host.Process.Run()", "error", err)
		return processes, op.Fail, err
	}

	processInfo := make([]string, 0)
	for eachProcess := range processes {
		process := processes[eachProcess]
		processInfo = append(processInfo, process.Executable())
	}

	return processInfo, op.Success, nil
}
