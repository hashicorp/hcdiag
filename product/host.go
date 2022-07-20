package product

import (
	"runtime"

	"github.com/hashicorp/hcdiag/hcl"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/hcdiag/runner"
	"github.com/hashicorp/hcdiag/runner/host"
)

// NewHost takes a logger, config, and HCL, and it creates a Product with all the host's default runners.
func NewHost(logger hclog.Logger, cfg Config, hcl2 *hcl.Host) (*Product, error) {
	product := &Product{
		l:      logger.Named("product"),
		Name:   Host,
		Config: cfg,
	}
	var os string
	if cfg.OS == "auto" {
		os = runtime.GOOS
	}
	// TODO(mkcp): Host can have an API client now and it would simplify quite a bit.
	product.Runners = hostRunners(os)
	if hcl2 != nil {
		hclRunners, err := hcl.BuildRunners(hcl2, cfg.TmpDir, nil)
		if err != nil {
			return nil, err
		}
		product.Runners = append(product.Runners, hclRunners...)
		product.Excludes = hcl2.Excludes
		product.Selects = hcl2.Selects
	}
	return product, nil
}

// hostRunners generates a slice of runners to inspect the host.
func hostRunners(os string) []runner.Runner {
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
