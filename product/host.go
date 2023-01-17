// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package product

import (
	"context"
	"runtime"
	"time"

	"github.com/hashicorp/hcdiag/runner/do"
	"github.com/hashicorp/hcdiag/runner/host"

	"github.com/hashicorp/hcdiag/hcl"
	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/hcdiag/runner"
)

const (
	TimeoutTenSeconds    = runner.Timeout(10 * time.Second)
	TimeoutThirtySeconds = runner.Timeout(30 * time.Second)
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

		hclRunners, err := hcl.BuildRunnersWithContext(ctx, hcl2, cfg.TmpDir, cfg.DebugDuration, cfg.DebugInterval, nil, cfg.Since, cfg.Until, cfg.Redactions)
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
		host.NewOSWithContext(ctx, host.OSConfig{OS: os, Redactions: redactions, Timeout: time.Duration(TimeoutTenSeconds)}),
		host.NewDiskWithContext(ctx, host.DiskConfig{Redactions: redactions}),
		host.NewInfoWithContext(ctx, host.InfoConfig{Redactions: redactions, Timeout: time.Duration(TimeoutTenSeconds)}),
		host.NewMemoryWithContext(ctx, TimeoutThirtySeconds),
		host.NewProcessWithContext(ctx, host.ProcessConfig{Redactions: redactions, Timeout: time.Duration(TimeoutTenSeconds)}),
		host.NewNetworkWithContext(ctx, host.NetworkConfig{Redactions: redactions, Timeout: time.Duration(TimeoutTenSeconds)}),
		host.NewIPTablesWithContext(ctx, host.IPTablesConfig{
			OS:         os,
			Redactions: redactions,
			Timeout:    TimeoutThirtySeconds,
		}),
		host.NewEtcHostsWithContext(ctx, host.EtcHostsConfig{
			OS:         os,
			Redactions: redactions,
			Timeout:    TimeoutThirtySeconds,
		}),
		host.NewProcFileWithContext(ctx, host.ProcFileConfig{
			OS:         os,
			Redactions: redactions,
			Timeout:    TimeoutThirtySeconds,
		}),
	}

	fsTab, err := host.NewFSTab(host.FSTabConfig{
		OS:         os,
		Timeout:    TimeoutTenSeconds,
		Redactions: redactions,
	})
	// TODO(mkcp): Errors should propagate back from hostRunners().
	if err != nil {
		l.Error("unable to create host.FSTab runner.", "err=", err)
	}
	r = append(r, fsTab)

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
