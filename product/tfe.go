package product

import (
	"github.com/hashicorp/hcdiag/client"
	s "github.com/hashicorp/hcdiag/seeker"
)

// NewTFE takes a product config and creates a Product containing all of TFE's seekers.
func NewTFE(cfg Config) (*Product, error) {
	api, err := client.NewTFEAPI()
	if err != nil {
		return nil, err
	}

	seekers, err := TFESeekers(cfg, api)
	if err != nil {
		return nil, err
	}
	return &Product{
		Seekers: seekers,
	}, nil
}

// TFESeekers seek information about Terraform Enterprise/Cloud.
func TFESeekers(cfg Config, api *client.APIClient) ([]*s.Seeker, error) {
	return []*s.Seeker{
		s.NewCommander("replicatedctl support-bundle", "string"),

		s.NewCopier("/var/lib/replicated/support-bundles/replicated-support*.tar.gz", cfg.TmpDir, cfg.Since, cfg.Until),

		s.NewHTTPer(api, "/api/v2/admin/customization-settings"),
		s.NewHTTPer(api, "/api/v2/admin/general-settings"),
		s.NewHTTPer(api, "/api/v2/admin/organizations"),
		s.NewHTTPer(api, "/api/v2/admin/terraform-versions"),
		s.NewHTTPer(api, "/api/v2/admin/twilio-settings"),
		// page size 1 because we only actually care about total workspace count in the `meta` field
		s.NewHTTPer(api, "/api/v2/admin/workspaces?page[size]=1"),
	}, nil
}
