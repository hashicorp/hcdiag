package product

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/client"
	"github.com/hashicorp/hcdiag/op"
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

// FIXME(mkcp): doccment
// TFERunners...
func TFERunners(cfg Config, api *client.APIClient) ([]op.Runner, error) {
	return []op.Runner{
		op.NewCommander("replicatedctl support-bundle", "string"),

		op.NewCopier("/var/lib/replicated/support-bundles/replicated-support*.tar.gz", cfg.TmpDir, cfg.Since, cfg.Until),

		op.NewHTTPer(api, "/api/v2/admin/customization-settings"),
		op.NewHTTPer(api, "/api/v2/admin/general-settings"),
		op.NewHTTPer(api, "/api/v2/admin/organizations"),
		op.NewHTTPer(api, "/api/v2/admin/terraform-versions"),
		op.NewHTTPer(api, "/api/v2/admin/twilio-settings"),
		// page size 1 because we only actually care about total workspace count in the `meta` field
		op.NewHTTPer(api, "/api/v2/admin/workspaces?page[size]=1"),
	}, nil
}
