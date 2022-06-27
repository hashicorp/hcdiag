package host

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/runner"
	"github.com/mitchellh/go-ps"
)

var _ runner.Runner = &Process{}

type Process struct{}

func (p Process) ID() string {
	return "process"
}

func (p Process) Run() op.Op {
	processes, err := ps.Processes()
	if err != nil {
		hclog.L().Trace("runner/host.Process.Run()", "error", err)
		return op.New(p, processes, op.Fail, err)
	}

	processInfo := make([]string, 0)
	for eachProcess := range processes {
		process := processes[eachProcess]
		processInfo = append(processInfo, process.Executable())
	}

	return op.New(p, processInfo, op.Success, nil)
}
