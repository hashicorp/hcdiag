package host

import (
	"fmt"
	"runtime"

	"github.com/hashicorp/hcdiag/op"

	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = IPTables{}

type IPTables struct {
	Commands []string `json:"commands"`
}

// NewIPTables returns a runner configured to run several iptables commands
func NewIPTables() *IPTables {
	return &IPTables{
		Commands: []string{
			"iptables -L -n -v",
			"iptables -L -n -v -t nat",
			"iptables -L -n -v -t mangle",
		},
	}
}

func (r IPTables) ID() string {
	return "iptables"
}

func (r IPTables) Run() op.Op {
	if runtime.GOOS != "linux" {
		return op.New(r.ID(), nil, op.Skip, fmt.Errorf("os not linux, skipping, os=%s", runtime.GOOS), runner.Params(r))
	}
	result := make(map[string]string)
	for _, c := range r.Commands {
		o := runner.NewCommander(c, "string", nil).Run()
		result[c] = o.Result.(string)
		if o.Error != nil {
			return op.New(r.ID(), result, op.Fail, o.Error, runner.Params(r))
		}
	}
	return op.New(r.ID(), result, op.Success, nil, runner.Params(r))
}
