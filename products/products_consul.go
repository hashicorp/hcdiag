package products

import (
	"fmt"

	s "github.com/hashicorp/host-diagnostics/seeker"
)

// ConsulSeekers seek information about Consul.
func ConsulSeekers(tmpDir string) []*s.Seeker {
	return []*s.Seeker{
		s.NewCommander("consul info", "string", false),
		s.NewCommander("consul members", "string", false),
		s.NewCommander("consul operator raft list-peers", "string", false),
		s.NewCommander(fmt.Sprintf("consul debug -output=%s/ConsulDebug -duration=%ds -interval=%ds", tmpDir, DebugSeconds, IntervalSeconds), "string", false),
	}
}
