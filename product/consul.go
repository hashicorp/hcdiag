package product

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/hcdiag/client"
	"github.com/hashicorp/hcdiag/runner"
	s "github.com/hashicorp/hcdiag/runner"
	logs "github.com/hashicorp/hcdiag/runner/log"
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
		l:       logger.Named("product"),
		Name:    Consul,
		Runners: ops,
		Config:  cfg,
	}, nil
}

// ConsulOps generates a slice of ops to inspect consul
func ConsulOps(cfg Config, api *client.APIClient) ([]s.Runner, error) {
	runners := []runner.Runner{
		runner.NewCommander("consul version", "string"),
		runner.NewCommander(fmt.Sprintf("consul debug -output=%s/ConsulDebug -duration=%s -interval=%s", cfg.TmpDir, cfg.DebugDuration, cfg.DebugInterval), "string"),

		runner.NewHTTPer(api, "/v1/agent/self"),
		runner.NewHTTPer(api, "/v1/agent/metrics"),
		runner.NewHTTPer(api, "/v1/catalog/datacenters"),
		runner.NewHTTPer(api, "/v1/catalog/services"),
		runner.NewHTTPer(api, "/v1/namespace"),
		runner.NewHTTPer(api, "/v1/status/leader"),
		runner.NewHTTPer(api, "/v1/status/peers"),

		logs.NewDocker("consul", cfg.TmpDir, cfg.Since),
		logs.NewJournald("consul", cfg.TmpDir, cfg.Since, cfg.Until),
	}

	// try to detect log location to copy
	if logPath, err := client.GetConsulLogPath(api); err == nil {
		dest := filepath.Join(cfg.TmpDir, "logs/consul")
		logCopier := runner.NewCopier(logPath, dest, cfg.Since, cfg.Until)
		runners = append([]runner.Runner{logCopier}, runners...)
	}

	return runners, nil
}
