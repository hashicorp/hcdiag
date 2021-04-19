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
		s.NewCommander(fmt.Sprintf("consul debug -output=%s/ConsulDebug -duration=%ds -interval=%ds", tmpDir, DebugSeconds, IntervalSeconds), "string", false),

		// s.NewCommander("consul info", "string", false),                     // use api instead
		// s.NewCommander("consul members", "string", false),                  // use api instead
		// s.NewCommander("consul operator raft list-peers", "string", false), // use api instead

		s.NewHTTPer(api, "/v1/agent/self", false),          // config and member info of local agent
		s.NewHTTPer(api, "/v1/agent/metrics", false),       // metrics for most recent finished interval
		s.NewHTTPer(api, "/v1/catalog/datacenters", false), // list of known datacenters
		s.NewHTTPer(api, "/v1/catalog/services", false),    // list of registered services, consider accepting 'dc' and 'ns' params
		s.NewHTTPer(api, "/v1/namespace", false),           // list all Namespaces (enterprise)
		s.NewHTTPer(api, "/v1/status/leader", false),       // get Raft leader for dc
		s.NewHTTPer(api, "/v1/status/peers", false),        // get Raft peers for dc

		// consider allowing service param to enable /catalog/service/:service, /catalog/connect/:service
		// consider allowing node param to enable /catalog/node-services/:node
		// consider allowing gateway param to enable /catalog/gateway-services/:gateway
		//	https://www.consul.io/api-docs/catalog

		// consider allowing config kind to enable /config/:kind, or add each possible separately e.g. /config/service-defaults
		//	https://www.consul.io/api-docs/config

		// params (or assume all) also enables /health endpoints

		// consul config files (discover)
		// consul logs (after discover config)
		// consul info (basic info cli, find api equiv?)
		// consul list peers
		// consul members
	}
}
