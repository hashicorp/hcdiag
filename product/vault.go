package product

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-hclog"

	logs "github.com/hashicorp/hcdiag/op/log"

	"github.com/hashicorp/hcdiag/client"
	s "github.com/hashicorp/hcdiag/op"
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

	ops, err := VaultOps(cfg, api)
	if err != nil {
		return nil, err
	}
	return &Product{
		l:      logger.Named("product"),
		Name:   Vault,
		Ops:    ops,
		Config: cfg,
	}, nil
}

// VaultOps generates a list of ops to inspect Vault.
func VaultOps(cfg Config, api *client.APIClient) ([]*s.Op, error) {
	ops := []*s.Op{
		s.NewCommander("vault version", "string"),
		s.NewCommander("vault status -format=json", "json"),
		s.NewCommander("vault read sys/health -format=json", "json"),
		s.NewCommander("vault read sys/seal-status -format=json", "json"),
		s.NewCommander("vault read sys/host-info -format=json", "json"),
		s.NewCommander(fmt.Sprintf("vault debug -output=%s/VaultDebug.tar.gz -duration=%s -interval=%s", cfg.TmpDir, cfg.DebugDuration, cfg.DebugInterval), "string"),

		logs.NewDocker("vault", cfg.TmpDir, cfg.Since),
		logs.NewJournald("vault", cfg.TmpDir, cfg.Since, cfg.Until),
	}

	// try to detect log location to copy
	if logPath, err := client.GetVaultAuditLogPath(api); err == nil {
		dest := filepath.Join(cfg.TmpDir, "logs/vault")
		logCopier := s.NewCopier(logPath, dest, cfg.Since, cfg.Until)
		ops = append([]*s.Op{logCopier}, ops...)
	}

	return ops, nil
}
