package debug

import (
	"fmt"
	"time"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = SimpleDebug{}

// The "Simple" debug wrapper can be used for Nomad, Vault, and Consul.
// It only deals with flags common to all three 'debug' commands
type SimpleDebug struct {
	ProductName string
	Duration    string
	Interval    string
	Output      string
	// "Filters" is a generic name for the target/topic/capture option (depending on the product)
	Filters    []string
	Redactions []*redact.Redact
}

func (d SimpleDebug) ID() string {
	return fmt.Sprintf("SimpleDebug-%s", d.ProductName)
}

// NewSimpleDebug takes a product config, product debug filters, and redactions, returning a pointer to a new SimpleDebug
func NewSimpleDebug(productName string, output string, debugDuration time.Duration, debugInterval time.Duration, filters []string, redactions []*redact.Redact) *SimpleDebug {
	return &SimpleDebug{
		ProductName: productName,
		Duration:    debugDuration.String(),
		Interval:    debugInterval.String(),
		// TODO should we worry about possible name collisions here? It would be strange to run multiple vault debugs.
		Output:     output,
		Filters:    filters,
		Redactions: redactions,
	}
}

func (d SimpleDebug) Run() op.Op {
	startTime := time.Now()

	filterString, err := productFilterString(d.ProductName, d.Filters)
	if err != nil {
		return op.New(d.ID(), map[string]any{}, op.Fail, err, runner.Params(d), startTime, time.Now())
	}

	// Assemble the debug command to execute
	cmdStr := simpleCmdString(d, filterString)

	cmd := runner.Command{
		Command:    cmdStr,
		Format:     "string",
		Redactions: d.Redactions,
	}

	o := cmd.Run()
	if o.Error != nil {
		return op.New(d.ID(), o.Result, op.Fail, o.Error, runner.Params(d), startTime, time.Now())
	}

	return op.New(d.ID(), o.Result, op.Success, nil, runner.Params(d), startTime, time.Now())
}

func simpleCmdString(d SimpleDebug, filterString string) string {
	var cmdStr string

	switch d.ProductName {
	case "nomad":
		cmdStr = fmt.Sprintf(
			"nomad operator debug -log-level=TRACE -duration=%s -interval=%s -node-id=all -max-nodes=100 -output=%s%s",
			d.Duration,
			d.Interval,
			d.Output,
			filterString,
		)

	case "vault":
		cmdStr = fmt.Sprintf(
			"vault debug -compress=true -duration=%s -interval=%s -output=%s/VaultDebug.tar.gz%s",
			d.Duration,
			d.Interval,
			d.Output,
			filterString,
		)

	case "consul":
		cmdStr = fmt.Sprintf(
			"consul debug -duration=%s -interval=%s -output=%s%s",
			d.Duration,
			d.Interval,
			d.Output,
			filterString,
		)
	}

	return cmdStr
}
