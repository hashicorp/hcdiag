package product

import (
	"fmt"
	"path/filepath"

	logs "github.com/hashicorp/hcdiag/seeker/log"

	"github.com/hashicorp/hcdiag/client"
	s "github.com/hashicorp/hcdiag/seeker"
)

const (
	VaultClientCheck = "vault version"
	VaultAgentCheck  = "vault status"
)

// NewVault takes a product config and creates a Product containing all of Vault's seekers.
func NewVault(cfg Config) (*Product, error) {
	api, err := client.NewVaultAPI()
	if err != nil {
		return nil, err
	}

	seekers, err := VaultSeekers(cfg, api)
	if err != nil {
		return nil, err
	}
	return &Product{
		Seekers: seekers,
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
		s.NewCommander(fmt.Sprintf("vault debug -output=%s/VaultDebug.tar.gz -duration=%ds", cfg.TmpDir, DefaultDebugSeconds), "string"),

		logs.NewDocker("vault", cfg.TmpDir, cfg.Since),
	}

	// try to detect log location to copy
	if logPath, err := client.GetVaultAuditLogPath(api); err == nil {
		dest := filepath.Join(cfg.TmpDir, "logs/vault")
		logCopier := s.NewCopier(logPath, dest, cfg.Since, cfg.Until)
		seekers = append([]*s.Seeker{logCopier}, seekers...)
	}
	// get logs from journald if available
	if journald := s.JournaldGetter("vault", cfg.TmpDir, cfg.Since, cfg.Until); journald != nil {
		seekers = append(seekers, journald)
	}

	return seekers, nil
}
