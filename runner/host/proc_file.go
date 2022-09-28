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
	result := make(map[string]any)
	if p.OS != "linux" {
		return op.New(p.ID(), nil, op.Skip, fmt.Errorf("os not linux, skipping, os=%s", p.OS), runner.Params(p))
	}
	for _, c := range p.Commands {
		sheller := runner.NewSheller(c, p.Redactions).Run()
		if sheller.Error != nil {
			return op.New(p.ID(), sheller.Result, op.Fail, sheller.Error, runner.Params(p))
		}
		result[sheller.Identifier] = sheller.Result
	}
	return op.New(p.ID(), result, op.Success, nil, runner.Params(p))
}
