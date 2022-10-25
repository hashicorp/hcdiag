package host

import (
	"fmt"
	"time"

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
	startTime := time.Now()
	result := make(map[string]any)
	if p.OS != "linux" {
		return op.New(p.ID(), nil, op.Skip, fmt.Errorf("os not linux, skipping, os=%s", p.OS), runner.Params(p), startTime, time.Now())
	}
	for _, c := range p.Commands {
		shell := runner.NewShell(c, p.Redactions).Run()
		if shell.Error != nil {
			return op.New(p.ID(), shell.Result, op.Fail, shell.Error, runner.Params(p), startTime, time.Now())
		}
		result[shell.Identifier] = shell.Result
	}
	return op.New(p.ID(), result, op.Success, nil, runner.Params(p), startTime, time.Now())
}
