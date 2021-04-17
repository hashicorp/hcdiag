package products

import (
	"fmt"

	"github.com/hashicorp/host-diagnostics/apiclients"
	s "github.com/hashicorp/host-diagnostics/seeker"
)

// NomadSeekers seek information about Nomad.
func NomadSeekers(tmpDir string) []*s.Seeker {
	api := apiclients.NewNomadAPI()
	return []*s.Seeker{
		s.NewCommander("nomad version", "string", true),
		s.NewCommander("nomad node status -json", "json", false),
		s.NewCommander("nomad agent-info -json", "json", false),
		s.NewCommander(fmt.Sprintf("nomad operator debug -output=%s -duration=%ds", tmpDir, DebugSeconds), "string", false),
		// s.NewCommander("nomad operator metrics", "json", false), // TODO: uncomment (it's very verbose, so not noisy during testing)

		s.NewHTTPer(api, "/v1/agent/members", false),
		s.NewHTTPer(api, "/v1/plugins?type=csi", false),
		s.NewHTTPer(api, "/v1/operator/autopilot/configuration", false),
		s.NewHTTPer(api, "/v1/operator/raft/configuration", false),
	}
}
