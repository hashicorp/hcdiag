package products

import (
	s "github.com/hashicorp/host-diagnostics/seeker"
)

const (
	DebugSeconds    = 10
	// PLSFIX(kit): Nitpicky, but we should change this to DefaultIntervalSeconds. It's only used in Consul right now,
	//  but if the intent to keeping it in this file is to provide a default between products, we should communicate
	//  that in the name.
	IntervalSeconds = 5
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
		seekers = append(seekers, VaultSeekers(tmpDir)...)
	}
	return seekers, err
}
