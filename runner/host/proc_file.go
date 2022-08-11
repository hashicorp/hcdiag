package host

import (
	"fmt"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = ProcFile{}

type ProcFile struct {
	OS         string           `json:"os"`
	Commands   []string         `json:"commands"`
	Redactions []*redact.Redact `json:"redactions"`
}

func NewProcFile(os string, redactions []*redact.Redact) *ProcFile {
	return &ProcFile{
		OS: os,
		Commands: []string{
			"cat /proc/cpuinfo",
			"cat /proc/loadavg",
			"cat /proc/version",
			"cat /proc/vmstat",
		},
		Redactions: redactions,
	}
}

func (p ProcFile) ID() string {
	return "/proc/ files"
}

func (p ProcFile) Run() op.Op {
	if p.OS != "linux" {
		return op.New(p.ID(), nil, op.Skip, fmt.Errorf("os not linux, skipping, os=%s", p.OS), runner.Params(p))
	}
	m := make(map[string]interface{})
	for _, c := range p.Commands {
		sheller := runner.NewSheller(c, p.Redactions).Run()
		m[c] = sheller.Result
		if sheller.Error != nil {
			return op.New(p.ID(), m, op.Fail, sheller.Error, runner.Params(p))
		}
	}
	return op.New(p.ID(), m, op.Success, nil, runner.Params(p))
}
