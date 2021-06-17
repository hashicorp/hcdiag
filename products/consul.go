package products

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/hashicorp/host-diagnostics/apiclients"
	s "github.com/hashicorp/host-diagnostics/seeker"
)

const (
	ConsulClientCheck = "consul version"
	ConsulAgentCheck  = "consul info"
)

// NewConsul takes a product config and creates a Product with all of Consul's default seekers
func NewConsul(cfg Config) *Product {
	return &Product{
		Seekers: ConsulSeekers(cfg.TmpDir, cfg.From, cfg.To),
	}
}

// ConsulSeekers seek information about Consul.
func ConsulSeekers(tmpDir string, from, to time.Time) []*s.Seeker {
	api := apiclients.NewConsulAPI()

	seekers := []*s.Seeker{
		s.NewCommander("consul version", "string"),
		s.NewCommander(fmt.Sprintf("consul debug -output=%s/ConsulDebug -duration=%ds -interval=%ds", tmpDir, DefaultDebugSeconds, DefaultIntervalSeconds), "string"),

		s.NewHTTPer(api, "/v1/agent/self"),
		s.NewHTTPer(api, "/v1/agent/metrics"),
		s.NewHTTPer(api, "/v1/catalog/datacenters"),
		s.NewHTTPer(api, "/v1/catalog/services"),
		s.NewHTTPer(api, "/v1/namespace"),
		s.NewHTTPer(api, "/v1/status/leader"),
		s.NewHTTPer(api, "/v1/status/peers"),
	}

	// try to detect log location to copy
	if logPath, err := apiclients.GetConsulLogPath(api); err == nil {
		dest := filepath.Join(tmpDir, "logs/consul")
		logCopier := s.NewCopier(logPath, dest, from, to)
		seekers = append([]*s.Seeker{logCopier}, seekers...)
	}

	return seekers
}
