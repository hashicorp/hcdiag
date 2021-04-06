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
			"nomad plugin status",
			"string",
		),
		// TODO: this is a good example of a "table" format that could be parsed out.
		util.NewCommand(
			"nomad server members -detailed",
			"string",
		),
		util.NewCommand(
			"nomad operator raft list-peers",
			"string",
		),
		// util.NewCommand( // TODO: this one is rather noisy...
		// 	"nomad operator metrics",
		// 	"json",
		// ),
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
