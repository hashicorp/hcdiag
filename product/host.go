package product

import (
	"context"
	"runtime"

	"github.com/hashicorp/hcdiag/runner/do"
	"github.com/hashicorp/hcdiag/runner/host"

	"github.com/hashicorp/hcdiag/hcl"
	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/hcdiag/runner"
)

// NewHost takes a logger, config, and HCL, and it creates a Product with all the host's default runners.
func NewHost(logger hclog.Logger, cfg Config, hcl2 *hcl.Host) (*Product, error) {
	return NewHostWithContext(context.Background(), logger, cfg, hcl2)
}

// NewHostWithContext takes a context, a logger, config, and HCL, and it creates a Product with all the host's default runners.
func NewHostWithContext(ctx context.Context, logger hclog.Logger, cfg Config, hcl2 *hcl.Host) (*Product, error) {
	// Prepend product-specific redactions to agent-level redactions from cfg
	defaultRedactions, err := hostRedactions()
	if err != nil {
		return nil, err
	}
	cfg.Redactions = redact.Flatten(defaultRedactions, cfg.Redactions)

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

		hclRunners, err := hcl.BuildRunnersWithContext(ctx, hcl2, cfg.TmpDir, nil, cfg.Since, cfg.Until, cfg.Redactions)
		if err != nil {
			return nil, err
		}
		product.Runners = append(product.Runners, hclRunners...)
		product.Excludes = hcl2.Excludes
		product.Selects = hcl2.Selects
	}

	// TODO(mkcp): Host can have an API client now and it would simplify quite a bit.
	// Add built-in runners
	builtInRunners := hostRunners(ctx, os, cfg.Redactions, product.l)
	product.Runners = append(product.Runners, builtInRunners...)

	return product, nil
}

// hostRunners generates a slice of runners to inspect the host.
func hostRunners(ctx context.Context, os string, redactions []*redact.Redact, l hclog.Logger) []runner.Runner {
	r := []runner.Runner{
		host.NewOS(os, redactions),
		host.NewDisk(redactions),
		host.NewInfo(redactions),
		host.Memory{},
		host.NewProcess(redactions),
		host.NewNetwork(redactions),
		host.NewEtcHosts(redactions),
		host.NewIPTables(os, redactions),
		host.NewProcFile(os, redactions),
		host.NewFSTab(os, redactions),
	}

	runners := []runner.Runner{
		do.New(l, "host", "host runners", r),
	}
	return runners
}

// hostRedactions returns a slice of default redactions for this product
func hostRedactions() ([]*redact.Redact, error) {
	configs := []redact.Config{}
	redactions, err := redact.MapNew(configs)
	if err != nil {
		return nil, err
	}
	return redactions, nil
}
