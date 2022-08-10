package product

import (
	"runtime"

	"github.com/hashicorp/hcdiag/hcl"
	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/hcdiag/runner"
	"github.com/hashicorp/hcdiag/runner/host"
)

// NewHost takes a logger, config, and HCL, and it creates a Product with all the host's default runners.
func NewHost(logger hclog.Logger, cfg Config, hcl2 *hcl.Host) (*Product, error) {
	// Prepend product-specific redactions to agent-level redactions from cfg
	cfg.Redactions = redact.Flatten(getDefaultHostRedactions(), cfg.Redactions)

	product := &Product{
		l:      logger.Named("product"),
		Name:   Host,
		Config: cfg,
	}
	var os string
	if cfg.OS == "auto" {
		os = runtime.GOOS
	}

	if hcl2 != nil {
		// Map product-specific redactions from our config
		hclProductRedactions, err := hcl.MapRedacts(hcl2.Redactions)
		if err != nil {
			return nil, err
		}

		// Prepend product HCL redactions to our product defaults
		cfg.Redactions = redact.Flatten(hclProductRedactions, cfg.Redactions)

		hclRunners, err := hcl.BuildRunners(hcl2, cfg.TmpDir, nil, cfg.Since, cfg.Until, cfg.Redactions)
		if err != nil {
			return nil, err
		}
		product.Runners = append(product.Runners, hclRunners...)
		product.Excludes = hcl2.Excludes
		product.Selects = hcl2.Selects
	}

	// TODO(mkcp): Host can have an API client now and it would simplify quite a bit.
	// Add built-in runners
	builtInRunners := hostRunners(os, cfg.Redactions)
	product.Runners = append(product.Runners, builtInRunners...)

	return product, nil
}

// hostRunners generates a slice of runners to inspect the host.
func hostRunners(os string, redactions []*redact.Redact) []runner.Runner {
	return []runner.Runner{
		host.NewOS(os),
		host.NewDisk(),
		host.Info{},
		host.Memory{},
		host.Process{},
		host.Network{},
		host.NewEtcHosts(),
		host.NewIPTables(),
		host.NewProcFile(os),
		host.NewFSTab(os),
	}
}

// getDefaultHostRedactions returns a slice of default redactions for this product
func getDefaultHostRedactions() []*redact.Redact {
	configs := []redact.Config{}

	redactions, err := redact.MapNew(configs)
	if err != nil {
		panic("error getting default host redactions")
	}
	return redactions
}
