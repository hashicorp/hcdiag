package products

import (
	s "github.com/hashicorp/host-diagnostics/seeker"
)

const (
	DefaultDebugSeconds    = 10
	DefaultIntervalSeconds = 5
)

type Config struct {
	Consul bool
	Nomad  bool
	TFE    bool
	Vault  bool
	TmpDir string
}

// ConfigAllEnabled returns a Config struct with every product enabled
func NewConfigAllEnabled () Config {
	return Config{
		Consul: true,
		Nomad:  true,
		TFE:    true,
		Vault:  true,
	}
}

// GetSeekers provides product Seekers for gathering info.
func GetSeekers(cfg Config) (seekers []*s.Seeker, err error) {
	if cfg.Consul{
		seekers = append(seekers, ConsulSeekers(cfg.TmpDir)...)
	}
	if cfg.Nomad {
		seekers = append(seekers, NomadSeekers(cfg.TmpDir)...)
	}
	if cfg.TFE {
		seekers = append(seekers, TFESeekers(cfg.TmpDir)...)
	}
	if cfg.Vault {
		vaultSeekers, err := VaultSeekers(cfg.TmpDir)
		if err != nil {
			return seekers, err
		}
		seekers = append(seekers, vaultSeekers...)
	}
	return seekers, err
}
