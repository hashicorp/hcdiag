package host

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/runner"
	"github.com/mitchellh/go-ps"
)

var _ runner.Runner = &Process{}

// Process represents a single OS Process
type Process struct {
	Redactions []*redact.Redact `json:"redactions"`
}

func NewProcess(redactions []*redact.Redact) *Process {
	return &Process{
		Redactions: redactions,
	}
}

// Proc represents the process data we're collecting and returning
type Proc struct {
	Name string `json:"name"`
	PID  int    `json:"pid"`
	PPID int    `json:"ppid"`
}

func (p Process) ID() string {
	return "process"
}

func (p Process) Run() op.Op {
	// A simple slice of processes
	var processList []Proc

	processes, err := ps.Processes()
	if err != nil {
		hclog.L().Trace("runner/host.Process.Run()", "error", err)
		return op.New(p.ID(), processList, op.Fail, err, nil)
	}

	processList, err = p.convertProcessInfo(processes)
	if err != nil {
		hclog.L().Trace("runner/host.Process.Run()", "error", err)
		return op.New(p.ID(), processList, op.Fail, err, nil)
	}

	return op.New(p.ID(), processList, op.Success, nil, nil)
}

func (p Process) convertProcessInfo(inputInfo []ps.Process) ([]Proc, error) {
	var processList []Proc

	for _, process := range inputInfo {
		executable, err := redact.String(process.Executable(), p.Redactions)
		if err != nil {
			return processList, err
		}
		newProc := Proc{
			Name: executable,
			PID:  process.Pid(),
			PPID: process.PPid(),
		}
		processList = append(processList, newProc)
	}

	return processList, nil
}
