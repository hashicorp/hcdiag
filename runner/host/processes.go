package host

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/runner"
	"github.com/mitchellh/go-ps"
)

var _ runner.Runner = &Process{}

type Process struct{}

func (p Process) ID() string {
	return "process"
}

func (p Process) Run() runner.Op {
	processes, err := ps.Processes()
	if err != nil {
		hclog.L().Trace("runner/host.Process.Run()", "error", err)
		return runner.New(p, processes, runner.Fail, err)
	}

	processInfo := make([]string, 0)
	for eachProcess := range processes {
		process := processes[eachProcess]
		processInfo = append(processInfo, process.Executable())
	}

	return runner.New(p, processInfo, runner.Success, nil)
}
