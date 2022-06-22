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
func NewIPTables() *op.Op {
	return &op.Op{
		Identifier: "iptables",
		Runner: IPTables{
			commands: []string{
				"iptables -L -n -v",
				"iptables -L -n -v -t nat",
				"iptables -L -n -v -t mangle",
			},
		},
	}
}

func (s IPTables) Run() (interface{}, op.Status, error) {
	if runtime.GOOS != "linux" {
		return nil, op.Success, fmt.Errorf("os not linux, skipping, os=%s", runtime.GOOS)
	}
	result := make(map[string]string)
	for _, c := range s.commands {
		res, _, err := op.NewCommander(c, "string").Runner.Run()
		result[c] = res.(string)
		if err != nil {
			return result, op.Fail, err
		}
	}
	return result, op.Success, nil
}
