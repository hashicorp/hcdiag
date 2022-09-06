package host

import (
	"fmt"
	"runtime"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = IPTables{}

type IPTables struct {
	Commands   []string         `json:"commands"`
	Redactions []*redact.Redact `json:"redactions"`
}

// NewIPTables returns a runner configured to run several iptables commands
func NewIPTables(redactions []*redact.Redact) *IPTables {
	return &IPTables{
		Commands: []string{
			"iptables -L -n -v",
			"iptables -L -n -v -t nat",
			"iptables -L -n -v -t mangle",
		},
		Redactions: redactions,
	}
}

func (r IPTables) ID() string {
	return "iptables"
}

func (r IPTables) Run() []op.Op {
	opList := make([]op.Op, 0)

	if runtime.GOOS != "linux" {
		return append(opList, op.New(r.ID(), nil, op.Skip, fmt.Errorf("os not linux, skipping, os=%s", runtime.GOOS), runner.Params(r)))
	}
	result := make(map[string]string)
	for _, c := range r.Commands {
		o := runner.NewCommander(c, "string", r.Redactions).Run()
		result[c] = o[0].Result.(string)
		if o[0].Error != nil {
			return append(opList, op.New(r.ID(), result, op.Fail, o[0].Error, runner.Params(r)))
		}
	}
	return append(opList, op.New(r.ID(), result, op.Success, nil, runner.Params(r)))
}
