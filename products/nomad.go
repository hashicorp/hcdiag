package products

import (
	"fmt"

	"github.com/hashicorp/host-diagnostics/util"
)

func NomadCommands(tempDir string) []util.CommandStruct {
	return []util.CommandStruct{
		util.NewCommand(
			"nomad version",
			"string",
		),
		util.NewCommand(
			"nomad node status -json",
			"json",
		),
		util.NewCommand(
			"nomad agent-info -json",
			"json",
		),
		util.NewCommand(
			"nomad operator raft list-peers",
			"string",
		),
		util.NewCommand(
			"nomad operator metrics",
			"json",
		),
		util.NewCommand(
			"nomad operator autopilot get-config",
			"string",
		),
		util.NewCommand(
			fmt.Sprintf("nomad operator debug -output=%s -duration=5s", tempDir),
			"string",
		),
	}
}
