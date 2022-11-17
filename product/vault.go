package product

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcdiag/hcl"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/runner/debug"
	"github.com/hashicorp/hcdiag/runner/do"

	"github.com/hashicorp/go-hclog"

	logs "github.com/hashicorp/hcdiag/runner/log"

	"github.com/hashicorp/hcdiag/client"
	"github.com/hashicorp/hcdiag/runner"
)

const (
	VaultClientCheck = "vault version"
	VaultAgentCheck  = "vault status"
)

// NewVault takes a product config and creates a Product containing all of Vault's runners.
func NewVault(logger hclog.Logger, cfg Config) (*Product, error) {
	return NewVaultWithContext(context.Background(), logger, cfg)
}

// NewVaultWithContext takes a context, a logger, and a config and creates a Product containing all of Vault's runners.
func NewVaultWithContext(ctx context.Context, logger hclog.Logger, cfg Config) (*Product, error) {
	// Prepend product-specific redactions to agent-level redactions from cfg
	defaultRedactions, err := vaultRedactions()
	if err != nil {
		return nil, err
	}
	cfg.Redactions = redact.Flatten(defaultRedactions, cfg.Redactions)

	product := &Product{
		l:      logger.Named("product"),
		Name:   Vault,
		Config: cfg,
	}
	api, err := client.NewVaultAPI()
	if err != nil {
		return nil, err
	}

	if cfg.HCL != nil {
		// Map product-specific redactions from our config
		hclProductRedactions, err := hcl.MapRedacts(cfg.HCL.Redactions)
		if err != nil {
			return nil, err
		}
		// Prepend product HCL redactions to our product defaults
		cfg.Redactions = redact.Flatten(hclProductRedactions, cfg.Redactions)

		hclRunners, err := hcl.BuildRunnersWithContext(ctx, cfg.HCL, cfg.TmpDir, cfg.DebugDuration, cfg.DebugInterval, api, cfg.Since, cfg.Until, nil)
		if err != nil {
			return nil, err
		}
		product.Runners = append(product.Runners, hclRunners...)
		product.Excludes = cfg.HCL.Excludes
		product.Selects = cfg.HCL.Selects
	}

	// Add built-in runners
	builtInRunners, err := vaultRunners(ctx, cfg, api, logger)
	if err != nil {
		return nil, err
	}
	product.Runners = append(product.Runners, builtInRunners...)

	return product, nil
}

// vaultRunners provides a list of default runners to inspect vault.
func vaultRunners(ctx context.Context, cfg Config, api *client.APIClient, l hclog.Logger) ([]runner.Runner, error) {
	var r []runner.Runner

	// Set up Command runners
	for _, cc := range []runner.CommandConfig{
		{Command: "vault version", Redactions: cfg.Redactions},
		{Command: "vault status -format=json", Format: "json", Redactions: cfg.Redactions},
		{Command: "vault read sys/health -format=json", Format: "json", Redactions: cfg.Redactions},
		{Command: "vault read sys/seal-status -format=json", Format: "json", Redactions: cfg.Redactions},
		{Command: "vault read sys/host-info -format=json", Format: "json", Redactions: cfg.Redactions},
		{Command: fmt.Sprintf("vault debug -output=%s/VaultDebug.tar.gz -duration=%s -interval=%s", cfg.TmpDir, cfg.DebugDuration, cfg.DebugInterval), Redactions: cfg.Redactions},
	} {
		c, err := runner.NewCommandWithContext(ctx, cc)
		if err != nil {
			return nil, err
		}
		r = append(r, c)
	}

	dbg, err := debug.NewVaultDebug(
		debug.VaultDebugConfig{
			Redactions: cfg.Redactions,
		},
		cfg.TmpDir,
		cfg.DebugDuration,
		cfg.DebugInterval,
	)
	if err != nil {
		return nil, err
	}
	r = append(r, dbg)

	r = append(r,
		logs.NewDocker("vault", cfg.TmpDir, cfg.Since, cfg.Redactions),
		logs.NewJournald("vault", cfg.TmpDir, cfg.Since, cfg.Until, cfg.Redactions),
	)

	// try to detect log location to copy
	if logPath, err := client.GetVaultAuditLogPath(api); err == nil {
		dest := filepath.Join(cfg.TmpDir, "logs/vault")
		logCopy := runner.NewCopy(logPath, dest, cfg.Since, cfg.Until, cfg.Redactions)
		r = append([]runner.Runner{logCopy}, r...)
	}

	runners := []runner.Runner{
		do.New(l, "vault", "vault runners", r),
	}
	return runners, nil
}

// vaultRedactions returns a slice of default redactions for this product
func vaultRedactions() ([]*redact.Redact, error) {
	configs := []redact.Config{}
	redactions, err := redact.MapNew(configs)
	if err != nil {
		return nil, err
	}
	return redactions, nil
}
