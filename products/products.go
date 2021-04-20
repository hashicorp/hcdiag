package products

import (
	"fmt"

	s "github.com/hashicorp/host-diagnostics/seeker"
)

const (
	DebugSeconds    = 10
	IntervalSeconds = 5
)

// GetSeekers provides product Seekers for gathering info.
func GetSeekers(product string, tmpDir string) (seekers []*s.Seeker, err error) {
	if product == "" {
		return seekers, err
	} else if product == "consul" {
		seekers = append(seekers, ConsulSeekers(tmpDir)...)
	} else if product == "nomad" {
		seekers = append(seekers, NomadSeekers(tmpDir)...)
	} else if product == "vault" {
		seekers = append(seekers, VaultSeekers(tmpDir)...)
	} else {
		err = fmt.Errorf("unsupported product '%s'", product)
	}
	return seekers, err
}
