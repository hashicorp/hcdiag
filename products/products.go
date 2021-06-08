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
	Seekers []*seeker.Seeker
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

func Setup(cfg Config) (map[string]*Product, error) {
	p := make(map[string]*Product)
	if cfg.Consul {
		p["consul"] = NewConsul(cfg)
	}
	if cfg.Nomad {
		p["nomad"] = NewNomad(cfg)
	}
	if cfg.TFE {
		p["terraform-ent"] = NewTFE(cfg)
	}
	if cfg.Vault {
		vaultSeekers, err := NewVault(cfg)
		if err != nil {
			return nil, err
		}
		p["vault"] = vaultSeekers
	}
	p["host"] = NewHost(cfg)
	return p, nil
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

func CountSeekers(products map[string]*Product) int {
	var count int
	for _, product := range products {
		count += len(product.Seekers)
	}
	return count
}
