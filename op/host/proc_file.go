package host

import (
	"fmt"

	"github.com/hashicorp/hcdiag/op"
)

var _ op.Runner = ProcFile{}

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

func (p ProcFile) Run() op.Op {
	if p.os != "linux" {
		// TODO(mkcp): Replace status with op.Skip when we implement it
		return op.New(p, nil, op.Success, fmt.Errorf("os not linux, skipping, os=%s", p.os))
	}
	m := make(map[string]interface{})
	for _, c := range p.commands {
		sheller := op.NewSheller(c).Run()
		m[c] = sheller.Result
		if sheller.Error != nil {
			return op.New(p, m, op.Fail, sheller.Error)
		}
	}
	return op.New(p, m, op.Success, nil)
}
