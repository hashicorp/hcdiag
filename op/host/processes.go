package host

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
	"github.com/mitchellh/go-ps"
)

var _ op.Runner = &Process{}

type Process struct{}

func (p Process) ID() string {
	return "process"
}

func (p Process) Run() op.Op {
	processes, err := ps.Processes()
	if err != nil {
		hclog.L().Trace("op/host.Process.Run()", "error", err)
		return p.op(processes, op.Fail, err)
	}

	processInfo := make([]string, 0)
	for eachProcess := range processes {
		process := processes[eachProcess]
		processInfo = append(processInfo, process.Executable())
	}

	return p.op(processInfo, op.Success, nil)
}

func (p Process) op(result interface{}, status op.Status, err error) op.Op {
	return op.Op{
		Identifier: p.ID(),
		Result:     result,
		Error:      err,
		ErrString:  err.Error(),
		Status:     status,
	}
}
