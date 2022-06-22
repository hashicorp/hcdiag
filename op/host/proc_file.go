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

func NewProcFile(os string) *op.Op {
	return &op.Op{
		Identifier: "/proc/ files",
		Runner: ProcFile{
			os: os,
			commands: []string{
				"cat /proc/cpuinfo",
				"cat /proc/loadavg",
				"cat /proc/version",
				"cat /proc/vmstat",
			},
		},
	}
}

func (s ProcFile) Run() (interface{}, op.Status, error) {
	if s.os != "linux" {
		// TODO(mkcp): Replace status with op.Skip when we implement it
		return nil, op.Success, fmt.Errorf("os not linux, skipping, os=%s", s.os)
	}
	m := make(map[string]interface{})
	for _, c := range s.commands {
		res, _, err := op.NewSheller(c).Runner.Run()
		m[c] = res
		if err != nil {
			return m, op.Fail, err
		}
	}
	return m, op.Success, nil
}
