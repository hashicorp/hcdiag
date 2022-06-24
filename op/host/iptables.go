package host

import (
	"fmt"
	"runtime"

	"github.com/hashicorp/hcdiag/util"

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
		return r.op(nil, op.Success, fmt.Errorf("os not linux, skipping, os=%s", runtime.GOOS))
	}
	result := make(map[string]string)
	for _, c := range r.commands {
		o := op.NewCommander(c, "string").Run()
		result[c] = o.Result.(string)
		if o.Error != nil {
			return r.op(result, op.Fail, o.Error)
		}
	}
	return r.op(result, op.Success, nil)
}

// TODO(mkcp): This pattern can be migrated to in op.NewOp(op.Runner, result, status, err) op.Op
func (r IPTables) op(result interface{}, status op.Status, err error) op.Op {
	return op.Op{
		Identifier: r.ID(),
		Result:     result,
		Error:      err,
		Status:     status,
		Params:     util.RunnerParams(r),
	}
}
