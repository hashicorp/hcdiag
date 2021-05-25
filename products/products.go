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

// ConfigAllEnabled returns a Config struct with every product enabled
func NewConfigAllEnabled () Config {
	return Config{
		Consul: true,
		Nomad:  true,
		TFE:    true,
		Vault:  true,
	}
}

// GetSeekers returns a map of enabled products to their seekers. If any of the products' healthchecks fail, we abort
//   the run. We want to abort the run here so we don't encourage users to send us incomplete diagnostics.
func GetSeekers(cfg Config) (map[string][]*seeker.Seeker, error) {
	sets := make(map[string][]*seeker.Seeker)
	if cfg.Consul {
		err := CommanderHealthCheck(ConsulClientCheck, ConsulAgentCheck)
		if err != nil {
			return nil, err
		}
		sets["consul"] = ConsulSeekers(cfg.TmpDir)
	}
	if cfg.Nomad {
		err := CommanderHealthCheck(NomadClientCheck, NomadAgentCheck)
		if err != nil {
			return nil, err
		}
		sets["nomad"] = NomadSeekers(cfg.TmpDir)
	}
	if cfg.TFE {
		// NOTE(mkcp): We don't have a TFE healthcheck because we don't support API checks yet.
		sets["terraform-ent"] = TFESeekers(cfg.TmpDir)
	}
	if cfg.Vault {
		err := CommanderHealthCheck(VaultClientCheck, VaultAgentCheck)
		if err != nil {
			return nil, err
		}
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
	if _, err := isClientAvailable.Run(); err != nil {
		return fmt.Errorf("client not available, healthcheck=%s, error=%s", client, err)
	}
	isAgentAvailable := seeker.NewCommander(agent, "string")
	if _, err := isAgentAvailable.Run(); err != nil {
		return fmt.Errorf("agent not available, healthcheck=%s, error=%s", agent, err)
	}
	return nil
}
