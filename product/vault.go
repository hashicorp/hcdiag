package product

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcdiag/hcl"
	"github.com/hashicorp/hcdiag/redact"
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

		hclRunners, err := hcl.BuildRunners(cfg.HCL, cfg.TmpDir, api, cfg.Since, cfg.Until, nil)
		if err != nil {
			return nil, err
		}
		product.Runners = append(product.Runners, hclRunners...)
		product.Excludes = cfg.HCL.Excludes
		product.Selects = cfg.HCL.Selects
	}

	// Add built-in runners
	builtInRunners, err := vaultRunners(cfg, api, logger)
	if err != nil {
		return nil, err
	}
	product.Runners = append(product.Runners, builtInRunners...)

	return product, nil
}

// vaultRunners provides a list of default runners to inspect vault.
func vaultRunners(cfg Config, api *client.APIClient, l hclog.Logger) ([]runner.Runner, error) {
	r := []runner.Runner{
		runner.NewCommander("vault version", "string", cfg.Redactions),
		runner.NewCommander("vault status -format=json", "json", cfg.Redactions),
		runner.NewCommander("vault read sys/health -format=json", "json", cfg.Redactions),
		runner.NewCommander("vault read sys/seal-status -format=json", "json", cfg.Redactions),
		runner.NewCommander("vault read sys/host-info -format=json", "json", cfg.Redactions),
		runner.NewCommander(fmt.Sprintf("vault debug -output=%s/VaultDebug.tar.gz -duration=%s -interval=%s", cfg.TmpDir, cfg.DebugDuration, cfg.DebugInterval), "string", cfg.Redactions),

		logs.NewDocker("vault", cfg.TmpDir, cfg.Since, cfg.Redactions),
		logs.NewJournald("vault", cfg.TmpDir, cfg.Since, cfg.Until, cfg.Redactions),
	}

	// try to detect log location to copy
	if logPath, err := client.GetVaultAuditLogPath(api); err == nil {
		dest := filepath.Join(cfg.TmpDir, "logs/vault")
		logCopier := runner.NewCopier(logPath, dest, cfg.Since, cfg.Until, cfg.Redactions)
		r = append([]runner.Runner{logCopier}, r...)
	}

	runners := []runner.Runner{do.New(l, "vault", "vault runners", r)}

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
