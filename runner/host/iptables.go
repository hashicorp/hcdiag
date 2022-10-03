package host

import (
	"fmt"
	"runtime"
	"time"

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

func (r IPTables) Run() op.Op {
	startTime := time.Now()

	if runtime.GOOS != "linux" {
		return op.New(r.ID(), nil, op.Skip, fmt.Errorf("os not linux, skipping, os=%s", runtime.GOOS), runner.Params(r), startTime, time.Now())
	}
	result := make(map[string]string)
	for _, c := range r.Commands {
		o := runner.NewCommander(c, "string", r.Redactions).Run()

		if o.Result != nil {
			result[c] = o.Result.(string)
		}

		if o.Error != nil {
			// If there's an error, pass through the Op's status and Error
			return op.New(r.ID(), result, o.Status, o.Error, runner.Params(r), startTime, time.Now())
		}
	}
	return op.New(r.ID(), result, op.Success, nil, runner.Params(r), startTime, time.Now())
}
