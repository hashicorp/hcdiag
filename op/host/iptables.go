package host

import (
	"fmt"
	"runtime"

	"github.com/hashicorp/hcdiag/op"
)

var _ op.Runner = IPTables{}

type IPTables struct {
	commands []string
}

// NewIPTables returns a op configured to run several iptables commands
func NewIPTables() *IPTables {
	return &IPTables{
		commands: []string{
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
		// TODO(mkcp): use skip status once available
		return op.New(r, nil, op.Success, fmt.Errorf("os not linux, skipping, os=%s", runtime.GOOS))
	}
	result := make(map[string]string)
	for _, c := range r.commands {
		o := op.NewCommander(c, "string").Run()
		result[c] = o.Result.(string)
		if o.Error != nil {
			return op.New(r, result, op.Fail, o.Error)
		}
	}
	return op.New(r, result, op.Success, nil)
}
