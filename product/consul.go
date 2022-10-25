package product

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcdiag/hcl"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/runner/do"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/hcdiag/client"
	"github.com/hashicorp/hcdiag/runner"
	logs "github.com/hashicorp/hcdiag/runner/log"
)

const (
	ConsulClientCheck = "consul version"
	ConsulAgentCheck  = "consul info"
)

// NewConsul takes a logger and product config, and it creates a Product with all of Consul's default runners.
func NewConsul(logger hclog.Logger, cfg Config) (*Product, error) {
	// Prepend product-specific redactions to agent-level redactions from cfg
	defaultRedactions, err := consulRedactions()
	if err != nil {
		return nil, err
	}
	cfg.Redactions = redact.Flatten(defaultRedactions, cfg.Redactions)

	product := &Product{
		l:      logger.Named("product"),
		Name:   Consul,
		Config: cfg,
	}
	api, err := client.NewConsulAPI()
	if err != nil {
		return nil, err
	}

	// HCL handling goes first, because it could add redactions to our built-in runners
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
	builtInRunners, err := consulRunners(cfg, api, logger)
	if err != nil {
		return nil, err
	}
	product.Runners = append(product.Runners, builtInRunners...)

	return product, nil
}

// consulRunners generates a slice of runners to inspect consul.
func consulRunners(cfg Config, api *client.APIClient, l hclog.Logger) ([]runner.Runner, error) {
	r := []runner.Runner{
		runner.NewCommand("consul version", "string", cfg.Redactions),
		runner.NewCommand(fmt.Sprintf("consul debug -output=%s/ConsulDebug -duration=%s -interval=%s", cfg.TmpDir, cfg.DebugDuration, cfg.DebugInterval), "string", cfg.Redactions),
		runner.NewCommand("consul operator raft list-peers -stale=true", "string", cfg.Redactions),

		runner.NewHTTPer(api, "/v1/agent/self", cfg.Redactions),
		runner.NewHTTPer(api, "/v1/agent/metrics", cfg.Redactions),
		runner.NewHTTPer(api, "/v1/catalog/datacenters", cfg.Redactions),
		runner.NewHTTPer(api, "/v1/catalog/services", cfg.Redactions),
		runner.NewHTTPer(api, "/v1/namespace", cfg.Redactions),
		runner.NewHTTPer(api, "/v1/status/leader", cfg.Redactions),
		runner.NewHTTPer(api, "/v1/status/peers", cfg.Redactions),
		runner.NewHTTPer(api, "/v1/agent/members?cached", cfg.Redactions),

		logs.NewDocker("consul", cfg.TmpDir, cfg.Since, cfg.Redactions),
		logs.NewJournald("consul", cfg.TmpDir, cfg.Since, cfg.Until, cfg.Redactions),
	}

	// try to detect log location to copy
	if logPath, err := client.GetConsulLogPath(api); err == nil {
		dest := filepath.Join(cfg.TmpDir, "logs/consul")
		logCopier := runner.NewCopy(logPath, dest, cfg.Since, cfg.Until, cfg.Redactions)
		r = append([]runner.Runner{logCopier}, r...)
	}

	runners := []runner.Runner{
		do.New(l, "consul", "consul runners", r),
	}
	return runners, nil
}

// consulRedactions returns a slice of default redactions for this product
func consulRedactions() ([]*redact.Redact, error) {
	configs := []redact.Config{}
	redactions, err := redact.MapNew(configs)
	if err != nil {
		return nil, err
	}
	return redactions, nil
}
