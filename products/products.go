package products

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/host-diagnostics/seeker"
	"time"
)

const (
	DefaultDebugSeconds    = 10
	DefaultIntervalSeconds = 5
)

type Config struct {
	Logger *hclog.Logger
	Consul bool
	Nomad  bool
	TFE    bool
	Vault  bool
	TmpDir string
	From   time.Time
	To     time.Time
	OS     string
}

type Product struct {
	Seekers     []*seeker.Seeker
}

// NewConfigAllEnabled returns a Config struct with every product enabled
func NewConfigAllEnabled (tmpDir string, from, to time.Time) Config {
	return Config{
		Consul: true,
		Nomad:  true,
		TFE:    true,
		Vault:  true,
		TmpDir: tmpDir,
		From:   from,
		To:     to,
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
