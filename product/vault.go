package product

import (
	"fmt"
	"path/filepath"
	"time"

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
	seekers, err := VaultSeekers(cfg.TmpDir, cfg.Since, cfg.Until)
	if err != nil {
		return nil, err
	}
	return &Product{
		Seekers: seekers,
	}, nil
}

// VaultSeekers seek information about Vault.
func VaultSeekers(tmpDir string, since, until time.Time) ([]*s.Seeker, error) {
	api, err := client.NewVaultAPI()
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

		logs.NewDocker("vault", tmpDir, since),
	}

	// try to detect log location to copy
	if logPath, err := client.GetVaultAuditLogPath(api); err == nil {
		dest := filepath.Join(tmpDir, "logs/vault")
		logCopier := s.NewCopier(logPath, dest, since, until)
		seekers = append([]*s.Seeker{logCopier}, seekers...)
	}
	// get logs from journald if available
	if journald := s.JournaldGetter("vault", tmpDir, since, until); journald != nil {
		seekers = append(seekers, journald)
	}

	return seekers, nil
}
