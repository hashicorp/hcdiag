package product

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcdiag/client"
	s "github.com/hashicorp/hcdiag/seeker"
	logs "github.com/hashicorp/hcdiag/seeker/log"
)

const (
	ConsulClientCheck = "consul version"
	ConsulAgentCheck  = "consul info"
)

// NewConsul takes a product config and creates a Product with all of Consul's default seekers
func NewConsul(cfg Config) (*Product, error) {
	api, err := client.NewConsulAPI()
	if err != nil {
		return nil, err
	}

	seekers, err := ConsulSeekers(cfg, api)
	if err != nil {
		return nil, err
	}
	return &Product{
		Seekers: seekers,
	}, nil
}

// ConsulSeekers seek information about Consul.
func ConsulSeekers(cfg Config, api *client.APIClient) ([]*s.Seeker, error) {
	seekers := []*s.Seeker{
		s.NewCommander("consul version", "string"),
		s.NewCommander(fmt.Sprintf("consul debug -output=%s/ConsulDebug -duration=%s -interval=%s", cfg.TmpDir, cfg.DebugDuration, cfg.DebugInterval), "string"),

		s.NewHTTPer(api, "/v1/agent/self"),
		s.NewHTTPer(api, "/v1/agent/metrics"),
		s.NewHTTPer(api, "/v1/catalog/datacenters"),
		s.NewHTTPer(api, "/v1/catalog/services"),
		s.NewHTTPer(api, "/v1/namespace"),
		s.NewHTTPer(api, "/v1/status/leader"),
		s.NewHTTPer(api, "/v1/status/peers"),

		logs.NewDocker("consul", cfg.TmpDir, cfg.Since),
		logs.NewJournald("consul", cfg.TmpDir, cfg.Since, cfg.Until),
	}

	// try to detect log location to copy
	if logPath, err := client.GetConsulLogPath(api); err == nil {
		dest := filepath.Join(cfg.TmpDir, "logs/consul")
		logCopier := s.NewCopier(logPath, dest, cfg.Since, cfg.Until)
		seekers = append([]*s.Seeker{logCopier}, seekers...)
	}

	return seekers, nil
}
