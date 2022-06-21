package product

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/seeker"
)

const (
	Consul = "consul"
	Host   = "host"
	Nomad  = "nomad"
	TFE    = "terraform-ent"
	Vault  = "vault"
)

const (
	DefaultDuration = 10 * time.Second
	DefaultInterval = 5 * time.Second
)

type Config struct {
	Name          string
	TmpDir        string
	Since         time.Time
	Until         time.Time
	OS            string
	DebugDuration time.Duration
	DebugInterval time.Duration
}

type Product struct {
	l        hclog.Logger
	Name     string
	Seekers  []*seeker.Seeker
	Excludes []string
	Selects  []string
	Config   Config
}

// Run iterates over the list of seekers in a product and stores each seeker into a map.
func (p *Product) Run() map[string]seeker.Seeker {
	p.l.Info("Running seekers for", "product", p.Name)
	results := make(map[string]seeker.Seeker)
	for _, s := range p.Seekers {
		p.l.Info("running operation", "product", p.Name, "seeker", s.Identifier)
		result, err := s.Run()
		// NOTE(mkcp): There's nothing stopping Run() from being called multiple times, so we'll copy the seeker off the product once it's done.
		// TODO(mkcp): It would be nice if we got an immutable seeker result type back from seeker runs instead.
		results[s.Identifier] = *s
		// Note seeker errors to users and keep going.
		if err != nil {
			p.l.Warn("result",
				"seeker", s.Identifier,
				"result", fmt.Sprintf("%s", result),
				"error", err,
			)
		}
	}
	return results
}

// Filter applies our slices of exclude and select seeker.Identifier matchers to the set of the product's seekers
func (p *Product) Filter() error {
	if p.Seekers == nil {
		p.Seekers = []*seeker.Seeker{}
	}
	var err error
	// The presence of Selects takes precedence over Excludes
	if p.Selects != nil && 0 < len(p.Selects) {
		p.Seekers, err = seeker.Select(p.Selects, p.Seekers)
		// Skip any Excludes
		return err
	}
	// No Selects, we can apply Excludes
	if p.Excludes != nil {
		p.Seekers, err = seeker.Exclude(p.Excludes, p.Seekers)
	}
	return err
}

// CommanderHealthCheck employs the CLI to check if the client and then the agent are available.
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
