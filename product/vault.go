package product

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-hclog"

	logs "github.com/hashicorp/hcdiag/seeker/log"

	"github.com/hashicorp/hcdiag/client"
	s "github.com/hashicorp/hcdiag/seeker"
)

const (
	VaultClientCheck = "vault version"
	VaultAgentCheck  = "vault status"
)

// NewVault takes a product config and creates a Product containing all of Vault's seekers.
func NewVault(logger hclog.Logger, cfg Config) (*Product, error) {
	api, err := client.NewVaultAPI()
	if err != nil {
		return nil, err
	}

	seekers, err := VaultSeekers(cfg, api)
	if err != nil {
		return nil, err
	}
	return &Product{
		l:       logger.Named("product"),
		Name:    Vault,
		Seekers: seekers,
		Config:  cfg,
	}, nil
}

// VaultSeekers seek information about Vault.
func VaultSeekers(cfg Config, api *client.APIClient) ([]*s.Seeker, error) {
	seekers := []*s.Seeker{
		s.NewCommander("vault version", "string"),
		s.NewCommander("vault status -format=json", "json"),
		s.NewCommander("vault read sys/health -format=json", "json"),
		s.NewCommander("vault read sys/seal-status -format=json", "json"),
		s.NewCommander("vault read sys/host-info -format=json", "json"),
		s.NewCommander(fmt.Sprintf("vault debug -output=%s/VaultDebug.tar.gz -duration=%s -interval=%s", cfg.TmpDir, cfg.DebugDuration, cfg.DebugInterval), "string"),
		// https://www.vaultproject.io/api-docs/system/internal-counters#client-count
		s.NewHTTPer(api, "/v1/sys/internal/counters/activity"),
		logs.NewDocker("vault", cfg.TmpDir, cfg.Since),
		logs.NewJournald("vault", cfg.TmpDir, cfg.Since, cfg.Until),
	}

	// try to detect log location to copy
	if logPath, err := client.GetVaultAuditLogPath(api); err == nil {
		dest := filepath.Join(cfg.TmpDir, "logs/vault")
		logCopier := s.NewCopier(logPath, dest, cfg.Since, cfg.Until)
		seekers = append([]*s.Seeker{logCopier}, seekers...)
	}

	return seekers, nil
}
