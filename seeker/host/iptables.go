package host

import (
	"fmt"
	"runtime"

	"github.com/hashicorp/hcdiag/seeker"
)

var _ seeker.Runner = IPTables{}

type IPTables struct {
	commands []string
}

// NewIPTables returns a seeker configured to run several iptables commands
func NewIPTables() *seeker.Seeker {
	return &seeker.Seeker{
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

func (s IPTables) Run() (interface{}, seeker.Status, error) {
	if runtime.GOOS != "linux" {
		return nil, seeker.Success, fmt.Errorf("os not linux, skipping, os=%s", runtime.GOOS)
	}
	result := make(map[string]string)
	for _, c := range s.commands {
		res, _, err := seeker.NewCommander(c, "string").Runner.Run()
		result[c] = res.(string)
		if err != nil {
			return result, seeker.Fail, err
		}
	}
	return result, seeker.Success, nil
}
