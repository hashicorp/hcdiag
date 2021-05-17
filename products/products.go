package products

import (
	s "github.com/hashicorp/host-diagnostics/seeker"
)

const (
	DefaultDebugSeconds    = 10
	DefaultIntervalSeconds = 5
)

// TODO(kit): refactor later, see https://hashicorp.atlassian.net/browse/ENGSYS-1199
// GetSeekers provides product Seekers for gathering info.
func GetSeekers(consul bool, nomad bool, tfe bool, vault bool, all bool, tmpDir string) (seekers []*s.Seeker, err error) {
	if consul || all {
		seekers = append(seekers, ConsulSeekers(tmpDir)...)
	}
	if nomad || all {
		seekers = append(seekers, NomadSeekers(tmpDir)...)
	}
	if tfe || all {
		seekers = append(seekers, TFESeekers(tmpDir)...)
	}
	if vault || all {
		vaultSeekers, err := VaultSeekers(tmpDir)
		if err != nil {
			return seekers, err
		}
		seekers = append(seekers, vaultSeekers...)
	}
	return seekers, err
}
