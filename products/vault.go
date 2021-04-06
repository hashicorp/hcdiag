package products

import (
	"fmt"

	"github.com/hashicorp/host-diagnostics/util"
)

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
