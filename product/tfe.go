package product

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/client"
	s "github.com/hashicorp/hcdiag/seeker"
)

// NewTFE takes a product config and creates a Product containing all of TFE's seekers.
func NewTFE(logger hclog.Logger, cfg Config) (*Product, error) {
	api, err := client.NewTFEAPI()
	if err != nil {
		return nil, err
	}

	seekers, err := TFESeekers(cfg, api)
	if err != nil {
		return nil, err
	}
	return &Product{
		l:       logger.Named("product"),
		Name:    TFE,
		Seekers: seekers,
		Config:  cfg,
	}, nil
}

// TFESeekers seek information about Terraform Enterprise/Cloud.
func TFESeekers(cfg Config, api *client.APIClient) ([]*s.Seeker, error) {
	return []*s.Seeker{
		// https://support.hashicorp.com/hc/en-us/articles/360047764514-Terraform-Enterprise-Support-Bundles-Are-Empty
		s.NewCommander("replicatedctl -version", "string"),
		s.NewCommander("journalctl --disk-usage", "string"),
		s.NewCommander("journalctl --vacuum-size=500M", "string"),
		s.NewCommander("docker system df", "string"),
		s.NewCommander("df -h", "string"),
		s.NewCommander("replicatedctl app status", "string"),
		s.NewCommander("systemctl status replicated replicated-ui replicated-operator", "string"),
		s.NewCommander("docker system prune -f", "string"),
		s.NewCommander("replicatedctl support-bundle", "string"),
		s.NewCopier("/var/lib/replicated/support-bundles/replicated-support*.tar.gz", cfg.TmpDir, cfg.Since, cfg.Until),
		s.NewHTTPer(api, "/api/v2/admin/customization-settings"),
		s.NewHTTPer(api, "/api/v2/admin/general-settings"),
		s.NewHTTPer(api, "/api/v2/admin/organizations"),
		// page size 1 because we only actually care about total run count in the `meta` field
		s.NewHTTPer(api, "/api/v2/admin/runs?page[size]=1"),
		s.NewHTTPer(api, "/api/v2/admin/terraform-versions"),
		s.NewHTTPer(api, "/api/v2/admin/twilio-settings"),
		// page size 1 because we only actually care about total user count in the `meta` field
		s.NewHTTPer(api, "/api/v2/admin/users?page[size]=1"),
		// page size 1 because we only actually care about total workspace count in the `meta` field
		s.NewHTTPer(api, "/api/v2/admin/workspaces?page[size]=1"),
	}, nil
}
