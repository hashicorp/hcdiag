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
	var procs []Proc

	psProcs, err := ps.Processes()
	if err != nil {
		hclog.L().Trace("runner/host.Process.Run()", "error", err)
		return op.New(p.ID(), procs, op.Fail, err, nil)
	}

	procs, err = p.procs(psProcs)
	if err != nil {
		hclog.L().Trace("runner/host.Process.Run()", "error", err)
		return op.New(p.ID(), procs, op.Fail, err, nil)
	}

	return op.New(p.ID(), procs, op.Success, nil, nil)
}

func (p Process) procs(psProcs []ps.Process) ([]Proc, error) {
	var result []Proc

	for _, psProc := range psProcs {
		executable, err := redact.String(psProc.Executable(), p.Redactions)
		if err != nil {
			return result, err
		}
		proc := Proc{
			Name: executable,
			PID:  psProc.Pid(),
			PPID: psProc.PPid(),
		}
		result = append(result, proc)
	}

	return result, nil
}
