package debug

import (
	"fmt"
	"time"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/product"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = SimpleDebug{}

// The "Simple" debug wrapper can be used for Nomad, Vault, and Consul.
// It only deals with flags common to all three 'debug' commands
type SimpleDebug struct {
	ProductConfig product.Config `json:"productconfig"`
	// "Filters" is a generic name for the target/topic/capture option (depending on the product)
	Filters []string `json:"filters"`
	// TODO maybe simpleDebugs don't have redactions, since they're always going to be created inside of hcdiag default product runners?
	// Maybe only custom productDebugs have hcl/redactions?
	Redactions []*redact.Redact `json:"redactions"`
}

func (d SimpleDebug) ID() string {
	return fmt.Sprintf("SimpleDebug-%s", d.ProductConfig.Name)
}

// NewSimpleDebug takes a product config, product debug filters, and redactions, returning a pointer to a new SimpleDebug
func NewSimpleDebug(cfg product.Config, filters []string, redactions []*redact.Redact) *SimpleDebug {
	return &SimpleDebug{
		ProductConfig: cfg,
		Filters:       filters,
		Redactions:    redactions,
	}
}

func (d SimpleDebug) Run() op.Op {
	startTime := time.Now()

	filterString, err := productFilterString(d.ProductConfig.Name, d.Filters)
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

	switch d.ProductConfig.Name {
	case "nomad":
		cmdStr = fmt.Sprintf(
			"nomad operator debug -log-level=TRACE -duration=%s -interval=%s -node-id=all -max-nodes=100 -output=%s/%s",
			d.ProductConfig.DebugDuration,
			d.ProductConfig.DebugInterval,
			d.ProductConfig.TmpDir,
			filterString,
		)

	case "vault":
		cmdStr = fmt.Sprintf(
			"vault debug -compress=true -duration=%s -interval=%s -output=%s/VaultDebug.tar.gz%s",
			d.ProductConfig.DebugDuration,
			d.ProductConfig.DebugInterval,
			d.ProductConfig.TmpDir,
			filterString,
		)

	case "consul":
		cmdStr = fmt.Sprintf(
			"consul debug -duration=%s -interval=%s -output=%s/ConsulDebug%s",
			d.ProductConfig.DebugDuration,
			d.ProductConfig.DebugInterval,
			d.ProductConfig.TmpDir,
			filterString,
		)
	}

	return cmdStr
}
