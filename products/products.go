package products

import (
	"fmt"

	s "github.com/hashicorp/host-diagnostics/seeker"
)

const DebugSeconds = 3

// GetSeekers provides product Seekers for gathering info.
func GetSeekers(product string, tmpDir string) (seekers []*s.Seeker, err error) {
	if product == "" {
		return seekers, err
	} else if product == "nomad" {
		seekers = append(seekers, NomadSeekers(tmpDir)...)
	} else if product == "vault" {
		seekers = append(seekers, VaultSeekers(tmpDir)...)
	} else if product == "terraform" {
		seekers = append(seekers, TerraformSeekers(tmpDir)...)
	} else {
		err = fmt.Errorf("unsupported product '%s'", product)
	}
	return seekers, err
}
