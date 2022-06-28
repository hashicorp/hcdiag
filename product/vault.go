package product

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-hclog"

	logs "github.com/hashicorp/hcdiag/op/log"

	"github.com/hashicorp/hcdiag/client"
	"github.com/hashicorp/hcdiag/op"
)

const (
	VaultClientCheck = "vault version"
	VaultAgentCheck  = "vault status"
)

// NewVault takes a product config and creates a Product containing all of Vault's ops.
func NewVault(logger hclog.Logger, cfg Config) (*Product, error) {
	api, err := client.NewVaultAPI()
	if err != nil {
		return nil, err
	}

	runners, err := VaultRunners(cfg, api)
	if err != nil {
		return nil, err
	}
	return &Product{
		l:       logger.Named("product"),
		Name:    Vault,
		Runners: runners,
		Config:  cfg,
	}, nil
}

// TODO(mkcp): doccomment
// VaultRunners ...
func VaultRunners(cfg Config, api *client.APIClient) ([]op.Runner, error) {
	runners := []op.Runner{
		op.NewCommander("vault version", "string"),
		op.NewCommander("vault status -format=json", "json"),
		op.NewCommander("vault read sys/health -format=json", "json"),
		op.NewCommander("vault read sys/seal-status -format=json", "json"),
		op.NewCommander("vault read sys/host-info -format=json", "json"),
		op.NewCommander(fmt.Sprintf("vault debug -output=%s/VaultDebug.tar.gz -duration=%s -interval=%s", cfg.TmpDir, cfg.DebugDuration, cfg.DebugInterval), "string"),

		logs.NewDocker("vault", cfg.TmpDir, cfg.Since),
		logs.NewJournald("vault", cfg.TmpDir, cfg.Since, cfg.Until),
	}

	// try to detect log location to copy
	if logPath, err := client.GetVaultAuditLogPath(api); err == nil {
		dest := filepath.Join(cfg.TmpDir, "logs/vault")
		logCopier := op.NewCopier(logPath, dest, cfg.Since, cfg.Until)
		runners = append([]op.Runner{logCopier}, runners...)
	}

	return runners, nil
}
