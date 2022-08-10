package product

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcdiag/hcl"
	"github.com/hashicorp/hcdiag/redact"

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
	cfg.Redactions = redact.Flatten(getDefaultVaultRedactions(), cfg.Redactions)

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
	builtInRunners, err := vaultRunners(cfg, api)
	if err != nil {
		return nil, err
	}
	product.Runners = append(product.Runners, builtInRunners...)

	return product, nil
}

// vaultRunners provides a list of default runners to inspect vault.
func vaultRunners(cfg Config, api *client.APIClient) ([]runner.Runner, error) {
	runners := []runner.Runner{
		runner.NewCommander("vault version", "string", cfg.Redactions),
		runner.NewCommander("vault status -format=json", "json", cfg.Redactions),
		runner.NewCommander("vault read sys/health -format=json", "json", cfg.Redactions),
		runner.NewCommander("vault read sys/seal-status -format=json", "json", cfg.Redactions),
		runner.NewCommander("vault read sys/host-info -format=json", "json", cfg.Redactions),
		runner.NewCommander(fmt.Sprintf("vault debug -output=%s/VaultDebug.tar.gz -duration=%s -interval=%s", cfg.TmpDir, cfg.DebugDuration, cfg.DebugInterval), "string", cfg.Redactions),

		logs.NewDocker("vault", cfg.TmpDir, cfg.Since),
		logs.NewJournald("vault", cfg.TmpDir, cfg.Since, cfg.Until),
	}

	// try to detect log location to copy
	if logPath, err := client.GetVaultAuditLogPath(api); err == nil {
		dest := filepath.Join(cfg.TmpDir, "logs/vault")
		logCopier := runner.NewCopier(logPath, dest, cfg.Since, cfg.Until, cfg.Redactions)
		runners = append([]runner.Runner{logCopier}, runners...)
	}

	return runners, nil
}

// getDefaultVaultRedactions returns a slice of default redactions for this product
func getDefaultVaultRedactions() []*redact.Redact {
	redactions := []struct {
		name    string
		matcher string
		replace string
	}{}

	var defaultVaultRedactions = make([]*redact.Redact, len(redactions))
	for i, r := range redactions {
		redaction, err := redact.New(r.matcher, "", r.replace)
		if err != nil {
			// If there's an issue, return an empty slice so that we can just ignore these redactions
			return make([]*redact.Redact, 0)
		}
		defaultVaultRedactions[i] = redaction
	}
	return defaultVaultRedactions
}
