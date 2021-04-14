package products

import (
	"fmt"

	s "github.com/hashicorp/host-diagnostics/seeker"
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
