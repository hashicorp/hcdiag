package host

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/seeker"
	"github.com/mitchellh/go-ps"
)

var _ seeker.Runner = &Process{}

func NewProcess() *seeker.Seeker {
	return &seeker.Seeker{Identifier: "process", Runner: Process{}}
}

type Process struct{}

func (p Process) Run() (interface{}, seeker.Status, error) {
	processes, err := ps.Processes()
	if err != nil {
		hclog.L().Trace("seeker/host.Process.Run()", "error", err)
		return processes, seeker.Fail, err
	}

	processInfo := make([]string, 0)
	for eachProcess := range processes {
		process := processes[eachProcess]
		processInfo = append(processInfo, process.Executable())
	}

	return processInfo, seeker.Success, nil
}
