package product

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/hashicorp/hcdiag/hcl"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/runner/do"

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

// NewNomad takes a logger and product config, and it creates a Product with all of Nomad's default runners.
func NewNomad(logger hclog.Logger, cfg Config) (*Product, error) {
	// Prepend product-specific redactions to agent-level redactions from cfg
	defaultRedactions, err := nomadRedactions()
	if err != nil {
		return nil, err
	}
	cfg.Redactions = redact.Flatten(defaultRedactions, cfg.Redactions)

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
	// passing in a zero value from the agent would allow us to do that, but the flags library requires a default value
	// to be set in order to _show_ that default to the user, so we have to set the agent with that default.
	if DefaultDuration == cfg.DebugDuration {
		cfg.DebugDuration = NomadDebugDuration
	}
	if DefaultInterval == cfg.DebugInterval {
		cfg.DebugInterval = NomadDebugInterval
	}

	if cfg.HCL != nil {
		// Map product-specific redactions from our config
		hclProductRedactions, err := hcl.MapRedacts(cfg.HCL.Redactions)
		if err != nil {
			return nil, err
		}
		// Prepend product HCL redactions to our product defaults
		cfg.Redactions = redact.Flatten(hclProductRedactions, cfg.Redactions)

		hclRunners, err := hcl.BuildRunners(cfg.HCL, cfg.TmpDir, api, cfg.Since, cfg.Until, cfg.Redactions)
		if err != nil {
			return nil, err
		}
		product.Runners = append(product.Runners, hclRunners...)
		product.Excludes = cfg.HCL.Excludes
		product.Selects = cfg.HCL.Selects
	}

	// Add built-in runners
	builtInRunners, err := nomadRunners(cfg, api, product.l)
	if err != nil {
		return nil, err
	}
	product.Runners = append(product.Runners, builtInRunners...)

	return product, nil
}

// nomadRunners generates a slice of runners to inspect nomad
func nomadRunners(cfg Config, api *client.APIClient, l hclog.Logger) ([]runner.Runner, error) {
	r := []runner.Runner{
		runner.NewCommand("nomad version", "string", cfg.Redactions),
		runner.NewCommand("nomad node status -self -json", "json", cfg.Redactions),
		runner.NewCommand("nomad agent-info -json", "json", cfg.Redactions),
		runner.NewCommand(fmt.Sprintf("nomad operator debug -log-level=TRACE -node-id=all -max-nodes=10 -output=%s -duration=%s -interval=%s", cfg.TmpDir, cfg.DebugDuration, cfg.DebugInterval), "string", cfg.Redactions),

		runner.NewHTTP(api, "/v1/agent/members?stale=true", cfg.Redactions),
		runner.NewHTTP(api, "/v1/operator/autopilot/configuration?stale=true", cfg.Redactions),
		runner.NewHTTP(api, "/v1/operator/raft/configuration?stale=true", cfg.Redactions),

		logs.NewDocker("nomad", cfg.TmpDir, cfg.Since, cfg.Redactions),
		logs.NewJournald("nomad", cfg.TmpDir, cfg.Since, cfg.Until, cfg.Redactions),
	}

	// try to detect log location to copy
	if logPath, err := client.GetNomadLogPath(api); err == nil {
		dest := filepath.Join(cfg.TmpDir, "logs", "nomad")
		logCopy := runner.NewCopy(logPath, dest, cfg.Since, cfg.Until, cfg.Redactions)
		r = append([]runner.Runner{logCopy}, r...)
	}

	runners := []runner.Runner{
		do.New(l, "nomad", "nomad runners", r),
	}
	return runners, nil
}

// nomadRedactions returns a slice of default redactions for this product
func nomadRedactions() ([]*redact.Redact, error) {
	configs := []redact.Config{}
	redactions, err := redact.MapNew(configs)
	if err != nil {
		return nil, err
	}
	return redactions, nil
}
