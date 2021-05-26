package products

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/host-diagnostics/apiclients"
	s "github.com/hashicorp/host-diagnostics/seeker"
)

const (
	VaultClientCheck = "vault version"
	VaultAgentCheck  = "vault status"
)

// VaultSeekers seek information about Vault.
func VaultSeekers(tmpDir string) ([]*s.Seeker, error) {
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
		logCopier := s.NewCopier(logPath, dest)
		seekers = append([]*s.Seeker{logCopier}, seekers...)
	}

	return seekers, nil
}
