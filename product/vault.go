package product

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcdiag/hcl"

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
	product := &Product{
		l:      logger.Named("product"),
		Name:   Vault,
		Config: cfg,
	}
	api, err := client.NewVaultAPI()
	if err != nil {
		return nil, err
	}

	product.Runners, err = vaultRunners(cfg, api)
	if err != nil {
		return nil, err
	}
	if cfg.HCL != nil {
		hclRunners, err := hcl.BuildRunners(cfg.HCL, cfg.TmpDir, api, cfg.Since, cfg.Until, nil)
		if err != nil {
			return nil, err
		}
		product.Runners = append(product.Runners, hclRunners...)
		product.Excludes = cfg.HCL.Excludes
		product.Selects = cfg.HCL.Selects
	}

	return product, nil
}

// vaultRunners provides a list of default runners to inspect vault.
func vaultRunners(cfg Config, api *client.APIClient) ([]runner.Runner, error) {
	runners := []runner.Runner{
		runner.NewCommander("vault version", "string", nil),
		runner.NewCommander("vault status -format=json", "json", nil),
		runner.NewCommander("vault read sys/health -format=json", "json", nil),
		runner.NewCommander("vault read sys/seal-status -format=json", "json", nil),
		runner.NewCommander("vault read sys/host-info -format=json", "json", nil),
		runner.NewCommander(fmt.Sprintf("vault debug -output=%s/VaultDebug.tar.gz -duration=%s -interval=%s", cfg.TmpDir, cfg.DebugDuration, cfg.DebugInterval), "string", nil),

		logs.NewDocker("vault", cfg.TmpDir, cfg.Since),
		logs.NewJournald("vault", cfg.TmpDir, cfg.Since, cfg.Until),
	}

	// try to detect log location to copy
	if logPath, err := client.GetVaultAuditLogPath(api); err == nil {
		dest := filepath.Join(cfg.TmpDir, "logs/vault")
		logCopier := runner.NewCopier(logPath, dest, cfg.Since, cfg.Until, nil)
		runners = append([]runner.Runner{logCopier}, runners...)
	}

	return runners, nil
}
