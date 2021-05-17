package products

import (
	"github.com/hashicorp/host-diagnostics/apiclients"
	s "github.com/hashicorp/host-diagnostics/seeker"
)

// TFESeekers seek information about Terraform Enterprise/Cloud.
func TFESeekers(tmpDir string) []*s.Seeker {
	api := apiclients.NewTFEAPI()

	return []*s.Seeker{
		s.NewCommander("replicatedctl support-bundle", "string", false),

		s.NewCopier("/var/lib/replicated/support-bundles/replicated-support*.tar.gz", tmpDir, false),

		s.NewHTTPer(api, "/api/v2/admin/cost-estimation-settings", false),
		s.NewHTTPer(api, "/api/v2/admin/customization-settings", false),
		s.NewHTTPer(api, "/api/v2/admin/general-settings", false),
		s.NewHTTPer(api, "/api/v2/admin/organizations", false),
		s.NewHTTPer(api, "/api/v2/admin/saml-settings", false),
		s.NewHTTPer(api, "/api/v2/admin/smtp-settings", false),
		s.NewHTTPer(api, "/api/v2/admin/terraform-versions", false),
		s.NewHTTPer(api, "/api/v2/admin/twilio-settings", false),
	}
}
