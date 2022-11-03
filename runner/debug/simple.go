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
	Product product.Name `json:"product"`
	// "Filters" is a generic name for the target/topic/capture option (depending on the product)
	Filters []string       `json:"filters"`
	Command runner.Command `json:"command"`
	// TODO maybe simpleDebugs don't have redactions, since they're always going to be created inside of hcdiag default product runners?
	// Maybe only custom productDebugs have hcl/redactions?
	Redactions []*redact.Redact `json:"redactions"`
}

func (d SimpleDebug) ID() string {
	return fmt.Sprintf("SimpleDebug-%s", d.Product)
}

// NewSimpleDebug takes a product config, product debug filters, and redactions, returning a pointer to a new SimpleDebug
func NewSimpleDebug(cfg product.Config, filters []string, redactions []*redact.Redact) *SimpleDebug {
	var cmdStr string
	var product = cfg.Name

	filterString, err := productFilterString(product, filters)
	if err != nil {
		// TODO figure out error handling inside of a runner constructor -- no other runners need this
		panic(err)
	}

	switch product {
	case "nomad":
		cmdStr = fmt.Sprintf("nomad operator debug -log-level=TRACE -duration=%s -interval=%s -node-id=all -max-nodes=100 -output=%s/%s", cfg.DebugDuration, cfg.DebugInterval, cfg.TmpDir, filterString)
	case "vault":
		cmdStr = fmt.Sprintf("vault debug -compress=true -duration=%s -interval=%s -output=%s/VaultDebug.tar.gz%s", cfg.DebugDuration, cfg.DebugInterval, cfg.TmpDir, filterString)
	case "consul":
		cmdStr = fmt.Sprintf("consul debug -duration=%s -interval=%s -output=%s/ConsulDebug%s", cfg.DebugDuration, cfg.DebugInterval, cfg.TmpDir, filterString)
	}

	return &SimpleDebug{
		Product: product,
		Filters: filters,
		Command: runner.Command{
			Command:    cmdStr,
			Format:     "string",
			Redactions: redactions,
		},
		Redactions: redactions,
	}
}

func (d SimpleDebug) Run() op.Op {
	startTime := time.Now()

	o := d.Command.Run()
	if o.Error != nil {
		return op.New(d.ID(), o.Result, op.Fail, o.Error, runner.Params(d), startTime, time.Now())
	}

	return op.New(d.ID(), o.Result, op.Success, nil, runner.Params(d), startTime, time.Now())
}

// productFilterString takes a product.Name and a slice of filter strings, and produces valid, product-specific filter flags.
// The returned string is in the form " -target=metrics -target=pprof" (for Vault), " -capture=host" (for Consul), or " -event-topic=Allocation" (for Nomad)
func productFilterString(product product.Name, filters []string) (string, error) {
	var filterString string
	var legalFilters map[string]bool
	var optFlag string

	// Define valid filter flagnames and values for all products
	nomadOptFlag := "event-topic"
	nomadFilters := map[string]bool{
		"*":          true,
		"ACLToken":   true,
		"ACLPolicy":  true,
		"ACLRole":    true,
		"Job":        true,
		"Allocation": true,
		"Deployment": true,
		"Evaluation": true,
		"Node":       true,
		"Service":    true,
	}

	vaultOptFlag := "target"
	vaultFilters := map[string]bool{
		// TODO(dcohen) is "all" or "*" valid?
		"config":             true,
		"host":               true,
		"metrics":            true,
		"pprof":              true,
		"replication-status": true,
		"server-status":      true,
	}

	consulOptFlag := "capture"
	consulFilters := map[string]bool{
		// TODO(dcohen) is "all" or "*" valid?
		"agent":   true,
		"host":    true,
		"members": true,
		"metrics": true,
		"logs":    true,
		"pprof":   true,
	}

	switch product {
	case "nomad":
		legalFilters = nomadFilters
		optFlag = nomadOptFlag
	case "vault":
		legalFilters = vaultFilters
		optFlag = vaultOptFlag
	case "consul":
		legalFilters = consulFilters
		optFlag = consulOptFlag
	default:
		return "", fmt.Errorf("invalid product used in debug.productFilterString(): %s", product)
	}

	for _, f := range filters {
		if legalFilters[f] {
			// includes leading space
			filterString = fmt.Sprintf("%s -%s=%s", filterString, optFlag, f)
		} else {
			return "", fmt.Errorf("invalid filter string (%s) for %s used in debug.productFilterString()", f, product)
		}
	}

	return filterString, nil
}
