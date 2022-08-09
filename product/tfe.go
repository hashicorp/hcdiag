package product

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/client"
	"github.com/hashicorp/hcdiag/hcl"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/runner"
)

// NewTFE takes a logger and product config, and it creates a Product with all of TFE's default runners.
func NewTFE(logger hclog.Logger, cfg Config) (*Product, error) {
	// Prepend product-specific redactions to agent-level redactions from cfg
	cfg.Redactions = redact.Flatten(getDefaultTFERedactions(), cfg.Redactions)

	product := &Product{
		l:      logger.Named("product"),
		Name:   TFE,
		Config: cfg,
	}
	api, err := client.NewTFEAPI()
	if err != nil {
		return nil, err
	}

	if cfg.HCL != nil {
		// Map product-specific redactions from our config
		hclProductRedactions, err := hcl.MapRedacts(cfg.HCL.Redactions)
		if err != nil {
			return nil, err
		}
		// Prepend product HCL redactions to our product defaults
		cfg.Redactions = redact.Flatten(hclProductRedactions, cfg.Redactions)

		hclRunners, err := hcl.BuildRunners(cfg.HCL, cfg.TmpDir, api, cfg.Since, cfg.Until, nil)
		if err != nil {
			return nil, err
		}
		product.Runners = append(product.Runners, hclRunners...)
		product.Excludes = cfg.HCL.Excludes
		product.Selects = cfg.HCL.Selects
	}

	// Add built-in runners
	builtInRunners, err := tfeRunners(cfg, api)
	if err != nil {
		return nil, err
	}
	product.Runners = append(product.Runners, builtInRunners...)

	return product, nil
}

// tfeRunners configures a set of default runners for TFE.
func tfeRunners(cfg Config, api *client.APIClient) ([]runner.Runner, error) {
	return []runner.Runner{
		runner.NewCommander("replicatedctl support-bundle", "string", cfg.Redactions),

		runner.NewCopier("/var/lib/replicated/support-bundles/replicated-support*.tar.gz", cfg.TmpDir, cfg.Since, cfg.Until, cfg.Redactions),

		runner.NewHTTPer(api, "/api/v2/admin/customization-settings", cfg.Redactions),
		runner.NewHTTPer(api, "/api/v2/admin/general-settings", cfg.Redactions),
		runner.NewHTTPer(api, "/api/v2/admin/organizations", cfg.Redactions),
		runner.NewHTTPer(api, "/api/v2/admin/terraform-versions", cfg.Redactions),
		runner.NewHTTPer(api, "/api/v2/admin/twilio-settings", cfg.Redactions),
		// page size 1 because we only actually care about total workspace count in the `meta` field
		runner.NewHTTPer(api, "/api/v2/admin/workspaces?page[size]=1", cfg.Redactions),

		runner.NewCommander("docker -v", "string", cfg.Redactions),
		runner.NewCommander("replicatedctl app status --output json", "json", cfg.Redactions),
		runner.NewCommander("lsblk --json", "json", cfg.Redactions),

		runner.NewSheller("getenforce", cfg.Redactions),
		runner.NewSheller("env | grep -i proxy", cfg.Redactions),
	}, nil
}

// getDefaultTFERedactions returns a slice of default redactions for this product
func getDefaultTFERedactions() []*redact.Redact {
	redactions := []struct {
		name    string
		matcher string
		replace string
	}{
		// {
		// 	name:    "TFE-product-default",
		// 	matcher: "/tfe/",
		// 	replace: "TFE-product-default-redaction",
		// },
	}

	var defaultTFERedactions = make([]*redact.Redact, len(redactions))
	for i, r := range redactions {
		redaction, err := redact.New(r.matcher, "", r.replace)
		if err != nil {
			// If there's an issue, return an empty slice so that we can just ignore these redactions
			return make([]*redact.Redact, 0)
		}
		defaultTFERedactions[i] = redaction
	}
	return defaultTFERedactions
}
