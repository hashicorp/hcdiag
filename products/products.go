package products

import (
	"fmt"
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

// NewConfigAllEnabled returns a Config struct with every product enabled
func NewConfigAllEnabled () Config {
	return Config{
		Consul: true,
		Nomad:  true,
		TFE:    true,
		Vault:  true,
	}
}

// CheckAvailable runs healthchecks for each enabled product
func CheckAvailable(cfg Config) error {
	if cfg.Consul {
		err := CommanderHealthCheck(ConsulClientCheck, ConsulAgentCheck)
		if err != nil {
			return err
		}
	}
	if cfg.Nomad {
		err := CommanderHealthCheck(NomadClientCheck, NomadAgentCheck)
		if err != nil {
			return err
		}
	}
	// NOTE(mkcp): We don't have a TFE healthcheck because we don't support API checks yet.
	// if cfg.TFE {
	// }
	if cfg.Vault {
		err := CommanderHealthCheck(VaultClientCheck, VaultAgentCheck)
		if err != nil {
			return err
		}
	}
	return nil
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

// CommanderHealthCheck employs the the CLI to check if the client and then the agent are available.
func CommanderHealthCheck(client, agent string) error {
	isClientAvailable := seeker.NewCommander(client, "string")
	if result, err := isClientAvailable.Run(); err != nil {
		return fmt.Errorf("client not available, healthcheck=%v, result=%v, error=%v", client, result, err)
	}
	isAgentAvailable := seeker.NewCommander(agent, "string")
	if result, err := isAgentAvailable.Run(); err != nil {
		return fmt.Errorf("agent not available, healthcheck=%v, result=%v, error=%v", agent, result, err)
	}
	return nil
}
