package product

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/client"
	"github.com/hashicorp/hcdiag/hcl"
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
		runner.NewCommander("replicatedctl support-bundle", "string", nil),

		runner.NewCopier("/var/lib/replicated/support-bundles/replicated-support*.tar.gz", cfg.TmpDir, cfg.Since, cfg.Until, nil),

		runner.NewHTTPer(api, "/api/v2/admin/customization-settings", nil),
		runner.NewHTTPer(api, "/api/v2/admin/general-settings", nil),
		runner.NewHTTPer(api, "/api/v2/admin/organizations", nil),
		runner.NewHTTPer(api, "/api/v2/admin/terraform-versions", nil),
		runner.NewHTTPer(api, "/api/v2/admin/twilio-settings", nil),
		// page size 1 because we only actually care about total workspace count in the `meta` field
		runner.NewHTTPer(api, "/api/v2/admin/workspaces?page[size]=1", nil),

		runner.NewCommander("docker -v", "string", nil),
		runner.NewCommander("replicatedctl app status --output json", "json", nil),
		runner.NewCommander("lsblk --json", "json", nil),

		runner.NewSheller("getenforce", nil),
		runner.NewSheller("echo $proxy_ip", nil),
	}, nil
}
