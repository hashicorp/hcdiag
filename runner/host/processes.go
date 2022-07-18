package host

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/runner"
	"github.com/mitchellh/go-ps"
)

var _ runner.Runner = &Process{}

// Process represents a single OS Process
type Process struct {
	Name string `json:"name"`
	PID  int    `json:"pid"`
	PPID int    `json:"ppid"`
}

func (p Process) ID() string {
	return "process"
}

func (p Process) Run() op.Op {
	processes, err := ps.Processes()
	if err != nil {
		hclog.L().Trace("runner/host.Process.Run()", "error", err)
		return op.New(p.ID(), processes, op.Fail, err, nil)
	}

	// Maps parent PIDs to child processes
	var processInfo = make(map[int][]Process)

	for _, process := range processes {
		proc := Process{
			Name: process.Executable(),
			PID:  process.Pid(),
			PPID: process.PPid(),
		}
		// Append to slice of Process under this PPID
		processInfo[proc.PPID] = append(processInfo[proc.PPID], proc)
	}

	return op.New(p.ID(), processInfo, op.Success, nil, nil)
}
