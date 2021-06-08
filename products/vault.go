package products

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/hashicorp/host-diagnostics/apiclients"
	s "github.com/hashicorp/host-diagnostics/seeker"
)

const (
	VaultClientCheck = "vault version"
	VaultAgentCheck  = "vault status"
)

// NewVault takes a product config and creates a Product containing all of Vault's seekers.
func NewVault(cfg Config) (*Product, error) {
	seekers, err := VaultSeekers(cfg.TmpDir, cfg.From, cfg.To)
	if err != nil {
		return nil, err
	}
	return &Product{
		Seekers: seekers,
	}, nil
}

// VaultSeekers seek information about Vault.
func VaultSeekers(tmpDir string, from, to time.Time) ([]*s.Seeker, error) {
	api, err := apiclients.NewVaultAPI()
	if err != nil {
		return nil, err
	}

	seekers := []*s.Seeker{
		s.NewCommander("vault version", "string"),
		s.NewCommander("vault status -format=json", "json"),
		s.NewCommander("vault read sys/health -format=json", "json"),
		s.NewCommander("vault read sys/seal-status -format=json", "json"),
		s.NewCommander("vault read sys/host-info -format=json", "json"),
		s.NewCommander(fmt.Sprintf("vault debug -output=%s/VaultDebug.tar.gz -duration=%ds", tmpDir, DefaultDebugSeconds), "string"),
	}

	// try to detect log location to copy
	if logPath, err := apiclients.GetVaultAuditLogPath(api); err == nil {
		dest := filepath.Join(tmpDir, "logs/vault")
		logCopier := s.NewCopier(logPath, dest, from, to)
		seekers = append([]*s.Seeker{logCopier}, seekers...)
	}

	return seekers, nil
}
