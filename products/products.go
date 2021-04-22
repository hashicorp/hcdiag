package products

import (
	s "github.com/hashicorp/host-diagnostics/seeker"
)

const (
	DebugSeconds    = 10
	IntervalSeconds = 5
)

// GetSeekers provides product Seekers for gathering info.
func GetSeekers(consul bool, nomad bool, vault bool, tmpDir string) (seekers []*s.Seeker, err error) {
	if consul {
		seekers = append(seekers, ConsulSeekers(tmpDir)...)
	}
	if nomad {
		seekers = append(seekers, NomadSeekers(tmpDir)...)
	}
	if vault {
		seekers = append(seekers, VaultSeekers(tmpDir)...)
	}
	return seekers, err
}
