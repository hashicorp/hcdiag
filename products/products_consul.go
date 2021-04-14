package products

import (
	"fmt"

	s "github.com/hashicorp/host-diagnostics/seeker"
)

// ConsulSeekers seek information about Consul.
func ConsulSeekers(tmpDir string) []*s.Seeker {
	return []*s.Seeker{
		s.NewCommander(fmt.Sprintf("consul debug -output=%s/ConsulDebug.tar.gz -duration=%ds -interval=%ds", tmpDir, DebugSeconds, IntervalSeconds), "string", false),
		s.NewCommander("consul members >> %s/consul-members.txt", "string", false),
		s.NewCommander("consul operator raft list-peers >> %s/consul-peers.txt", "string", false),
		s.NewCommander("consul info >> %s/consul-info.txt", "string", false),
	}
}
