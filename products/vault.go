package products

import (
	"fmt"

	s "github.com/hashicorp/host-diagnostics/seeker"
	"github.com/hashicorp/host-diagnostics/util"
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

func VaultCommands(tempDir string) []util.CommandStruct {
	return []util.CommandStruct{
		util.NewCommand(
			"vault version",
			"string",
		),
		util.NewCommand(
			"vault status -format=json",
			"json",
		),
		util.NewCommand(
			"vault read sys/health -format=json",
			"json",
		),
		util.NewCommand(
			"vault read sys/host-info -format=json",
			"json",
		),
		util.NewCommand(
			"vault read sys/seal-status -format=json",
			"json",
		),
		util.NewCommand(
			fmt.Sprintf("vault debug -duration=5s -output=%s/VaultDebug.tar.gz", tempDir),
			"string",
		),
	}
}
