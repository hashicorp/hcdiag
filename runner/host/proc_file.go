package host

import (
	"fmt"

	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = ProcFile{}

type ProcFile struct {
	os       string
	commands []string
}

func NewProcFile(os string) *ProcFile {
	return &ProcFile{
		os: os,
		commands: []string{
			"cat /proc/cpuinfo",
			"cat /proc/loadavg",
			"cat /proc/version",
			"cat /proc/vmstat",
		},
	}
}

func (p ProcFile) ID() string {
	return "/proc/ files"
}

func (p ProcFile) Run() runner.Op {
	if p.os != "linux" {
		// TODO(mkcp): Replace status with op.Skip when we implement it
		return runner.New(p, nil, runner.Success, fmt.Errorf("os not linux, skipping, os=%s", p.os))
	}
	m := make(map[string]interface{})
	for _, c := range p.commands {
		sheller := runner.NewSheller(c).Run()
		m[c] = sheller.Result
		if sheller.Error != nil {
			return runner.New(p, m, runner.Fail, sheller.Error)
		}
	}
	return runner.New(p, m, runner.Success, nil)
}
