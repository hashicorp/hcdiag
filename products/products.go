package products

import (
	"github.com/hashicorp/host-diagnostics/seeker"
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

// GetSeekers returns a map of enabled products to their seekers.
func GetSeekers(cfg Config) (map[string][]*seeker.Seeker, error) {
	sets := make(map[string][]*seeker.Seeker)
	if cfg.Consul {
		sets["consul"] = ConsulSeekers(cfg.TmpDir)
	}
	if cfg.Nomad {
		sets["nomad"] = NomadSeekers(cfg.TmpDir)
	}
	if cfg.TFE {
		sets["terraform-ent"] = TFESeekers(cfg.TmpDir)
	}
	if cfg.Vault {
		vaultSeekers, err := VaultSeekers(cfg.TmpDir)
		if err != nil {
			return sets, err
		}
		sets["vault"] = vaultSeekers
	}
	return sets, nil
}
