package product

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/client"
	"github.com/hashicorp/hcdiag/runner"
)

// NewTFE takes a product config and creates a Product containing all of TFE's ops.
func NewTFE(logger hclog.Logger, cfg Config) (*Product, error) {
	api, err := client.NewTFEAPI()
	if err != nil {
		return nil, err
	}

	runners, err := TFERunners(cfg, api)
	if err != nil {
		return nil, err
	}
	return &Product{
		l:       logger.Named("product"),
		Name:    TFE,
		Runners: runners,
		Config:  cfg,
	}, nil
}

// TFERunners configures a set of default runners for TFE.
func TFERunners(cfg Config, api *client.APIClient) ([]runner.Runner, error) {
	return []runner.Runner{
		runner.NewCommander("replicatedctl support-bundle", "string"),

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
