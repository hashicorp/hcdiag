package host

import (
	"time"

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
	startTime := time.Now()
	var procs []Proc

	psProcs, err := ps.Processes()
	if err != nil {
		hclog.L().Trace("runner/host.Process.Run()", "error", err)
		results := map[string]any{"procs": psProcs}
		return op.New(p.ID(), results, op.Fail, err, runner.Params(p), startTime, time.Now())
	}

	procs, err = p.procs(psProcs)
	results := map[string]any{"procs": procs}
	if err != nil {
		hclog.L().Trace("runner/host.Process.Run()", "error", err)
		return op.New(p.ID(), results, op.Fail, err, runner.Params(p), startTime, time.Now())
	}

	return op.New(p.ID(), results, op.Success, nil, runner.Params(p), startTime, time.Now())
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
