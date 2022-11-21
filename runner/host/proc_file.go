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
		shell, err := runner.NewShell(runner.ShellConfig{
			Command:    c,
			Redactions: p.Redactions,
		})
		if err != nil {
			return op.New(p.ID(), map[string]any{}, op.Fail, err, runner.Params(p), startTime, time.Now())
		}
		o := shell.Run()
		if o.Error != nil {
			return op.New(p.ID(), o.Result, op.Fail, o.Error, runner.Params(p), startTime, time.Now())
		}
		result[o.Identifier] = o.Result
	}
	return op.New(p.ID(), result, op.Success, nil, runner.Params(p), startTime, time.Now())
}
