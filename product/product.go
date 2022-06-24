package product

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
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
	Runners  []op.Runner
	Excludes []string
	Selects  []string
	Config   Config
}

// Run iterates over the list of ops in a product and stores each op into a map.
func (p *Product) Run() map[string]op.Op {
	p.l.Info("Running operations for", "product", p.Name)
	results := make(map[string]op.Op)
	for _, r := range p.Runners {
		p.l.Info("running operation", "product", p.Name, "op", r.ID())
		o := r.Run()
		// NOTE(mkcp): There's nothing stopping Run() from being called multiple times, so we'll copy the op off the product once it's done.
		// TODO(mkcp): It would be nice if we got an immutable op result type back from op runs instead.
		results[r.ID()] = o
		// Note op errors to users and keep going.
		if o.Error != nil {
			p.l.Warn("result",
				"op", r.ID(),
				"result", fmt.Sprintf("%s", o.Result),
				"error", o.Error,
			)
		}
	}
	return results
}

// Filter applies our slices of exclude and select op.Identifier matchers to the set of the product's ops
func (p *Product) Filter() error {
	if p.Runners == nil {
		p.Runners = []op.Runner{}
	}
	var err error
	// The presence of Selects takes precedence over Excludes
	if p.Selects != nil && 0 < len(p.Selects) {
		p.Runners, err = op.Select(p.Selects, p.Runners)
		// Skip any Excludes
		return err
	}
	// No Selects, we can apply Excludes
	if p.Excludes != nil {
		p.Runners, err = op.Exclude(p.Excludes, p.Runners)
	}
	return err
}

// CommanderHealthCheck employs the CLI to check if the client and then the agent are available.
func CommanderHealthCheck(client, agent string) error {
	checkClient := op.NewCommander(client, "string").Run()
	if checkClient.Error != nil {
		return fmt.Errorf("client not available, healthcheck=%v, result=%v, error=%v", client, checkClient.Result, checkClient.Error)
	}
	checkAgent := op.NewCommander(agent, "string").Run()
	if checkAgent.Error != nil {
		return fmt.Errorf("agent not available, healthcheck=%v, result=%v, error=%v", agent, checkAgent.Result, checkAgent.Error)
	}
	return nil
}

func CountOps(products map[string]*Product) int {
	var count int
	for _, product := range products {
		count += len(product.Runners)
	}
	return count
}
