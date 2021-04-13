package products

import (
	"fmt"

	s "github.com/hashicorp/host-diagnostics/seeker"
)

// VaultSeekers seek information about Vault.
func VaultSeekers(tmpDir string) []*s.Seeker {
	return []*s.Seeker{
		s.NewCommander("vault version", "string", true),
		s.NewCommander("vault status -format=json", "json", false),
		s.NewCommander("vault read sys/health -format=json", "json", false),
		s.NewCommander("vault read sys/seal-status -format=json", "json", false),
		s.NewCommander("vault read sys/host-info -format=json", "json", false),
		s.NewCommander(fmt.Sprintf("vault debug -output=%s/VaultDebug.tar.gz -duration=%ds", tmpDir, DebugSeconds), "string", false),
	}
}
