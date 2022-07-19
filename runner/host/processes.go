package host

import (
	"fmt"

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

	// Maps parent PIDs to child processes
	var processInfo = make(map[int][]proc)

	for _, process := range processes {
		newProc := proc{
			Name: process.Executable(),
			PID:  process.Pid(),
			PPID: process.PPid(),
		}
		// Append to slice of Process under this proc's PPID
		processInfo[newProc.PPID] = append(processInfo[newProc.PPID], newProc)
	}

	// Reprocess our original processInfo map with the process list, to populate parent names in our new namedProcessMap
	// (only a single pass through the process list each time, to avoid using quadratic time)
	var namedProcessMap = make(map[string][]proc)

	for _, process := range processes {
		// If our PID is listed as a PPID in processInfo, we're a parent!
		children, ok := processInfo[process.Pid()]
		if ok {
			newProc := proc{
				Name: process.Executable(),
				PID:  process.Pid(),
				PPID: process.PPid(),
			}

			// Process names are not unique by default, but we need unique keys
			uniqueName := uniqueProcName(newProc)
			namedProcessMap[uniqueName] = children
		}
	}

	return op.New(p.ID(), namedProcessMap, op.Success, nil, nil)
}

// Creates a unique name from a proc's name and PID
func uniqueProcName(process proc) string {
	procName := process.Name
	pl := len(procName)
	// truncate the name to the first and last 5 letters: "processfoobarname" => "proce...rname"
	if pl > 10 {
		procName = fmt.Sprintf("%s...%s", procName[0:5], procName[pl-5:pl])
	}

	return fmt.Sprintf("%d-%s", process.PID, procName)
}
