package products

import (
	"github.com/hashicorp/hcdiag/apiclients"
	s "github.com/hashicorp/hcdiag/seeker"
	"time"
)

// NewTFE takes a product config and creates a Product containing all of TFE's seekers.
func NewTFE(cfg Config) *Product {
	return &Product{
		Seekers: TFESeekers(cfg.TmpDir, cfg.From, cfg.To),
	}
}

// TFESeekers seek information about Terraform Enterprise/Cloud.
func TFESeekers(tmpDir string, from, to time.Time) []*s.Seeker {
	api := apiclients.NewTFEAPI()

	return []*s.Seeker{
		s.NewCommander("replicatedctl support-bundle", "string"),

		s.NewCopier("/var/lib/replicated/support-bundles/replicated-support*.tar.gz", tmpDir, from, to),

		s.NewHTTPer(api, "/api/v2/admin/cost-estimation-settings"),
		s.NewHTTPer(api, "/api/v2/admin/customization-settings"),
		s.NewHTTPer(api, "/api/v2/admin/general-settings"),
		s.NewHTTPer(api, "/api/v2/admin/organizations"),
		s.NewHTTPer(api, "/api/v2/admin/saml-settings"),
		s.NewHTTPer(api, "/api/v2/admin/smtp-settings"),
		s.NewHTTPer(api, "/api/v2/admin/terraform-versions"),
		s.NewHTTPer(api, "/api/v2/admin/twilio-settings"),
	}
}
