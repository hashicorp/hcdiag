package products

import (
	s "github.com/hashicorp/host-diagnostics/seeker"
)

// TerraformSeekers seek information about Terraform.
func TerraformSeekers(tmpDir string) []*s.Seeker {
	return []*s.Seeker{
		s.NewCommander("replicatedctl support-bundle", "string", false),
		s.NewCommander("tfe-admin support-bundle", "string", false),
	}
}
