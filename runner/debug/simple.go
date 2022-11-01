package debug

import (
	"fmt"
	"strings"
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
	return fmt.Sprintf("simpledebug-%s", d.Product)
}

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
		// TODO(dcohen): compress is currently true to maintain backwards compatibility, but I'd like to set this to false since everything gets compressed anyway
		// TODO loglevel TRACE?
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
	var legalFilters []string
	var optFlag string

	nomadFilters := []string{"*", "ACLToken", "ACLPolicy", "ACLRole", "Job", "Allocation", "Deployment", "Evaluation", "Node", "Service"}
	nomadOptFlag := "event-topic"
	vaultFilters := []string{"config", "host", "metrics", "pprof", "replication-status", "server-status"}
	vaultOptFlag := "target"
	consulFilters := []string{"agent", "host", "members", "metrics", "logs", "pprof"}
	consulOptFlag := "capture"

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
		// Skip empty entries. TODO maybe this is something we shouldn't allow?
		if f == "" {
			continue
		}

		found, idx := indexOf(legalFilters, f)
		if found {
			// includes leading space
			// legalFilters[idx] (as opposed to s) ensures the correct filter option capitalization
			filterString = fmt.Sprintf("%s -%s=%s", filterString, optFlag, legalFilters[idx])
		} else {
			return "", fmt.Errorf("invalid filter string (%s) for %s used in debug.productFilterString()", f, product)
		}
	}

	return filterString, nil
}

// This can't really be necessary in Go, right?
// Searches s for presence of val: returns true and index of val, or false and 0
func indexOf(s []string, val string) (found bool, idx int) {
	// Early exit
	if len(s) == 0 || len(val) == 0 {
		return false, 0
	}

	for i, v := range s {
		// compare case-insensitive strings to avoid erroring on capitalization
		if strings.EqualFold(v, val) {
			return true, i
		}
	}
	return false, 0
}
