package product

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/hcdiag/client"
	s "github.com/hashicorp/hcdiag/op"
	logs "github.com/hashicorp/hcdiag/op/log"
)

const (
	ConsulClientCheck = "consul version"
	ConsulAgentCheck  = "consul info"
)

// NewConsul takes a product config and creates a Product with all of Consul's default ops
func NewConsul(logger hclog.Logger, cfg Config) (*Product, error) {
	api, err := client.NewConsulAPI()
	if err != nil {
		return nil, err
	}

	ops, err := ConsulOps(cfg, api)
	if err != nil {
		return nil, err
	}
	return &Product{
		l:      logger.Named("product"),
		Name:   Consul,
		Ops:    ops,
		Config: cfg,
	}, nil
}

// ConsulOps generates a slice of ops to inspect consul
func ConsulOps(cfg Config, api *client.APIClient) ([]*s.Op, error) {
	ops := []*s.Op{
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
		ops = append([]*s.Op{logCopier}, ops...)
	}

	return ops, nil
}
