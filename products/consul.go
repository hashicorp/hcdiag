package products

import (
	"fmt"

	"github.com/hashicorp/host-diagnostics/apiclients"
	s "github.com/hashicorp/host-diagnostics/seeker"
)

// ConsulSeekers seek information about Consul.
func ConsulSeekers(tmpDir string) []*s.Seeker {
	api := apiclients.NewConsulAPI()
	return []*s.Seeker{
		s.NewCommander("consul version", "string", true),
		s.NewCommander(fmt.Sprintf("consul debug -output=%s/ConsulDebug -duration=%ds -interval=%ds", tmpDir, DefaultDebugSeconds, DefaultIntervalSeconds), "string", false),

		s.NewHTTPer(api, "/v1/agent/self", false),
		s.NewHTTPer(api, "/v1/agent/metrics", false),
		s.NewHTTPer(api, "/v1/catalog/datacenters", false),
		s.NewHTTPer(api, "/v1/catalog/services", false),
		s.NewHTTPer(api, "/v1/namespace", false),
		s.NewHTTPer(api, "/v1/status/leader", false),
		s.NewHTTPer(api, "/v1/status/peers", false),
	}
}
