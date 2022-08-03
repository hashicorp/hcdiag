package product

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcdiag/hcl"
	"github.com/hashicorp/hcdiag/redact"

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
	product := &Product{
		l:      logger.Named("product"),
		Name:   Consul,
		Config: cfg,
	}
	api, err := client.NewConsulAPI()
	if err != nil {
		return nil, err
	}

	// Prepend product-specific redactions to agent-level redactions from cfg
	cfg.Redactions = append(getDefaultConsulRedactions(), cfg.Redactions...)

	if cfg.HCL != nil {
		// Map product-specific redactions from our config
		// TODO(dcohen) error handling
		hclRedactions, _ := hcl.MapRedactions(cfg.HCL.Redactions)

		// Prepend product HCL redactions to our product defaults
		cfg.Redactions = append(hclRedactions, cfg.Redactions...)

		// TODO(dcohen) buildrunners needs to take redactions
		hclRunners, err := hcl.BuildRunners(cfg.HCL, cfg.TmpDir, api, cfg.Since, cfg.Until)
		if err != nil {
			return nil, err
		}
		product.Runners = append(product.Runners, hclRunners...)
		product.Excludes = cfg.HCL.Excludes
		product.Selects = cfg.HCL.Selects
	}

	fmt.Println("DCOHENDELETEME: finished setting redactions in NewConsul():")
	for _, r := range cfg.Redactions {
		fmt.Printf("%s : %s\n", r.ID, r.Replace)
	}

	// Add built-in runners
	builtInRunners, err := consulRunners(cfg, api)
	if err != nil {
		return nil, err
	}
	product.Runners = append(product.Runners, builtInRunners...)

	return product, nil
}

// consulRunners generates a slice of runners to inspect consul.
func consulRunners(cfg Config, api *client.APIClient) ([]runner.Runner, error) {
	runners := []runner.Runner{
		runner.NewCommander("consul version", "string", cfg.Redactions),
		runner.NewCommander(fmt.Sprintf("consul debug -output=%s/ConsulDebug -duration=%s -interval=%s", cfg.TmpDir, cfg.DebugDuration, cfg.DebugInterval), "string", cfg.Redactions),

		runner.NewHTTPer(api, "/v1/agent/self"),
		runner.NewHTTPer(api, "/v1/agent/metrics"),
		runner.NewHTTPer(api, "/v1/catalog/datacenters"),
		runner.NewHTTPer(api, "/v1/catalog/services"),
		runner.NewHTTPer(api, "/v1/namespace"),
		runner.NewHTTPer(api, "/v1/status/leader"),
		runner.NewHTTPer(api, "/v1/status/peers"),

		logs.NewDocker("consul", cfg.TmpDir, cfg.Since, cfg.Redactions),
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

// getDefaultConsulRedactions returns a slice of default redactions for this product
func getDefaultConsulRedactions() []*redact.Redact {
	redactions := []struct {
		name    string
		matcher string
		replace string
	}{
		{
			name:    "consul",
			matcher: "/consul/",
			replace: "consul-product-default-redactions",
		},
		{
			name:    "consul",
			matcher: "consul",
			replace: "consul-product-default-redactions",
		},
		{
			name:    "consul",
			matcher: "consultest",
		},
	}

	var defaultConsulRedactions = make([]*redact.Redact, len(redactions))
	for i, r := range redactions {
		redaction, err := redact.New(r.matcher, "", r.replace)
		if err != nil {
			// If there's an issue, return an empty slice so that we can just ignore agent redactions
			return make([]*redact.Redact, 0)
		}
		defaultConsulRedactions[i] = redaction
	}
	return defaultConsulRedactions
}
