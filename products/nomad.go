package products

import "github.com/hashicorp/host-diagnostics/util"

func NomadCommands() []util.CommandStruct {
	return []util.CommandStruct{
		util.NewCommand(
			"nomad version",
			"string",
		),
		util.NewCommand(
			"nomad node status -json",
			"json",
		),
	}
}
