package product

import (
	"fmt"
	"path/filepath"
	"time"

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
	seekers, err := ConsulSeekers(cfg.TmpDir, cfg.Since, cfg.Until)
	if err != nil {
		return nil, err
	}
	return &Product{
		Seekers: seekers,
	}, nil
}

// ConsulSeekers seek information about Consul.
func ConsulSeekers(tmpDir string, since, until time.Time) ([]*s.Seeker, error) {
	api, err := client.NewConsulAPI()
	if err != nil {
		return nil, err
	}

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

		logs.NewDocker("consul", tmpDir, since),
		logs.NewJournald("consul", tmpDir, since, until),
	}

	// try to detect log location to copy
	if logPath, err := client.GetConsulLogPath(api); err == nil {
		dest := filepath.Join(tmpDir, "logs/consul")
		logCopier := s.NewCopier(logPath, dest, since, until)
		seekers = append([]*s.Seeker{logCopier}, seekers...)
	}

	return seekers, nil
}
