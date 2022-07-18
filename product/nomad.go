package product

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/hashicorp/hcdiag/hcl"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/hcdiag/client"
	"github.com/hashicorp/hcdiag/runner"
	logs "github.com/hashicorp/hcdiag/runner/log"
)

const (
	NomadClientCheck   = "nomad version"
	NomadAgentCheck    = "nomad server members"
	NomadDebugDuration = 2 * time.Minute
	NomadDebugInterval = 30 * time.Second
)

// NewNomad takes a product config and creates a Product with all of Nomad's default ops
func NewNomad(logger hclog.Logger, cfg Config) (*Product, error) {
	product := &Product{
		l:      logger.Named("product"),
		Name:   Nomad,
		Config: cfg,
	}
	api, err := client.NewNomadAPI()
	if err != nil {
		return nil, err
	}

	// Apply nomad duration and interval default if CLI is using global defaults.
	// NOTE(mkcp): The downside to this approach is that Nomad cannot be run with a 10s duration and 5s interval.
	//  passing in a zero value from the agent would allow us to do that, but the flags library requires a default value
	//  to be set in order to _show_ that default to the user, so we have to set the agent with that default.
	if DefaultDuration == cfg.DebugDuration {
		cfg.DebugDuration = NomadDebugDuration
	}
	if DefaultInterval == cfg.DebugInterval {
		cfg.DebugInterval = NomadDebugInterval
	}
	product.Runners, err = nomadRunners(cfg, api)
	if err != nil {
		return nil, err
	}

	if cfg.HCL != nil {
		hclRunners, err := hcl.BuildRunners(cfg.HCL, cfg.TmpDir, api)
		if err != nil {
			return nil, err
		}
		product.Runners = append(product.Runners, hclRunners...)
		product.Excludes = cfg.HCL.Excludes
		product.Selects = cfg.HCL.Selects
	}

	return product, nil
}

// nomadRunners generates a slice of runners to inspect nomad
func nomadRunners(cfg Config, api *client.APIClient) ([]runner.Runner, error) {
	runners := []runner.Runner{
		runner.NewCommander("nomad version", "string"),
		runner.NewCommander("nomad node status -self -json", "json"),
		runner.NewCommander("nomad agent-info -json", "json"),
		runner.NewCommander(fmt.Sprintf("nomad operator debug -log-level=TRACE -node-id=all -max-nodes=10 -output=%s -duration=%s -interval=%s", cfg.TmpDir, cfg.DebugDuration, cfg.DebugInterval), "string"),

		runner.NewHTTPer(api, "/v1/agent/members?stale=true"),
		runner.NewHTTPer(api, "/v1/operator/autopilot/configuration?stale=true"),
		runner.NewHTTPer(api, "/v1/operator/raft/configuration?stale=true"),

		logs.NewDocker("nomad", cfg.TmpDir, cfg.Since),
		logs.NewJournald("nomad", cfg.TmpDir, cfg.Since, cfg.Until),
	}

	// try to detect log location to copy
	if logPath, err := client.GetNomadLogPath(api); err == nil {
		dest := filepath.Join(cfg.TmpDir, "logs", "nomad")
		logCopier := runner.NewCopier(logPath, dest, cfg.Since, cfg.Until)
		runners = append([]runner.Runner{logCopier}, runners...)
	}

	return runners, nil
}
