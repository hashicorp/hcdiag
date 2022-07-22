package host

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/runner"
	"github.com/mitchellh/go-ps"
)

var _ runner.Runner = &Process{}

// Process represents a single OS Process
type Process struct{}

// proc represents the process data we're collecting and returning
type proc struct {
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

	// A simple slice of processes
	var processList []proc

	for _, process := range processes {
		newProc := proc{
			Name: process.Executable(),
			PID:  process.Pid(),
			PPID: process.PPid(),
		}

		processList = append(processList, newProc)
	}

	return op.New(p.ID(), processList, op.Success, nil, nil)
}
