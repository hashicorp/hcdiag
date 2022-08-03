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
	product := &Product{
		l:      logger.Named("product"),
		Name:   TFE,
		Config: cfg,
	}
	api, err := client.NewTFEAPI()
	if err != nil {
		return nil, err
	}

	// Read in product-specific redactions here
	defaultRedactions := getDefaultTFERedactions()
	// Prepend our default reactions, so we run redactions from most-specific (runner) to least-specific (agent)
	cfg.Redactions = append(defaultRedactions, cfg.Redactions...)

	product.Runners, err = tfeRunners(cfg, api)
	if err != nil {
		return nil, err
	}

	if cfg.HCL != nil {
		hclRunners, err := hcl.BuildRunners(cfg.HCL, cfg.TmpDir, api, cfg.Since, cfg.Until)
		if err != nil {
			return nil, err
		}
		product.Runners = append(product.Runners, hclRunners...)
		product.Excludes = cfg.HCL.Excludes
		product.Selects = cfg.HCL.Selects
	}

	return product, nil
}

// tfeRunners configures a set of default runners for TFE.
func tfeRunners(cfg Config, api *client.APIClient) ([]runner.Runner, error) {
	return []runner.Runner{
		runner.NewCommander("replicatedctl support-bundle", "string", cfg.Redactions),

		runner.NewCopier("/var/lib/replicated/support-bundles/replicated-support*.tar.gz", cfg.TmpDir, cfg.Since, cfg.Until),

		runner.NewHTTPer(api, "/api/v2/admin/customization-settings"),
		runner.NewHTTPer(api, "/api/v2/admin/general-settings"),
		runner.NewHTTPer(api, "/api/v2/admin/organizations"),
		runner.NewHTTPer(api, "/api/v2/admin/terraform-versions"),
		runner.NewHTTPer(api, "/api/v2/admin/twilio-settings"),
		// page size 1 because we only actually care about total workspace count in the `meta` field
		runner.NewHTTPer(api, "/api/v2/admin/workspaces?page[size]=1"),
	}, nil
}

// getDefaultTFERedactions returns a set of default redactions for this product
func getDefaultTFERedactions() []*redact.Redact {
	redactions := []struct {
		name    string
		matcher string
	}{
		{
			name:    "tfe",
			matcher: "/tfe/",
		},
		{
			name:    "tfe",
			matcher: "tfe",
		},
		{
			name:    "tfe",
			matcher: "tfetest",
		},
	}

	var defaultTFERedactions = make([]*redact.Redact, len(redactions))
	for i, r := range redactions {
		redaction, err := redact.New(r.matcher, "", "")
		if err != nil {
			// If there's an issue, return an empty slice so that we can just ignore agent redactions
			return make([]*redact.Redact, 0)
		}
		defaultTFERedactions[i] = redaction
	}
	return defaultTFERedactions
}
