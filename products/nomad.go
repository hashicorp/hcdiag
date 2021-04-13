package products

import (
	"fmt"

	s "github.com/hashicorp/host-diagnostics/seeker"
	"github.com/hashicorp/host-diagnostics/util"
)

// NomadSeekers seek information about Nomad.
func NomadSeekers(tmpDir string) []*s.Seeker {
	return []*s.Seeker{
		s.NewCommander("nomad version", "string", true),
		s.NewCommander("nomad plugin status", "string", false),
		s.NewCommander("nomad server members -detailed", "string", false),
		s.NewCommander("nomad node status -json", "json", false),
		s.NewCommander("nomad operator raft list-peers", "string", false),
		s.NewCommander("nomad operator autopilot get-config", "string", false),
		s.NewCommander("nomad agent-info -json", "json", false),
		s.NewCommander(fmt.Sprintf("nomad operator debug -output=%s -duration=%ds", tmpDir, DebugSeconds), "string", false),
		// s.NewCommander("nomad operator metrics", "json", false), // TODO: uncomment (it's very verbose, so not noisy during testing)
	}
}

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
