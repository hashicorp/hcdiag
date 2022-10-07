package host

import (
	"fmt"
	"runtime"
	"time"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = EtcHosts{}

type EtcHosts struct {
	OS         string           `json:"os"`
	Redactions []*redact.Redact `json:"redactions"`
}

func NewEtcHosts(redactions []*redact.Redact) *EtcHosts {
	return &EtcHosts{
		OS:         runtime.GOOS,
		Redactions: redactions,
	}
}

func (r EtcHosts) ID() string {
	return "/etc/hosts"
}

func (r EtcHosts) Run() op.Op {
	startTime := time.Now()

	// Not compatible with windows
	if r.OS == "windows" {
		err := fmt.Errorf(" EtcHosts.Run() not available on os, os=%s", r.OS)
		return op.New(r.ID(), nil, op.Skip, err, runner.Params(r), startTime, time.Now())
	}
	s := runner.NewSheller("cat /etc/hosts", r.Redactions).Run()
	if s.Error != nil {
		return op.New(r.ID(), s.Result, op.Fail, s.Error, runner.Params(r), startTime, time.Now())
	}
	return op.New(r.ID(), s.Result, op.Success, nil, runner.Params(r), startTime, time.Now())
}
