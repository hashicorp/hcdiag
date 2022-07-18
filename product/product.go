package product

import (
	"fmt"
	"time"

	"github.com/hashicorp/hcdiag/hcl"

	"github.com/hashicorp/hcdiag/op"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/runner"
)

type Name string

const (
	Consul Name = "consul"
	Host   Name = "host"
	Nomad  Name = "nomad"
	TFE    Name = "terraform-ent"
	Vault  Name = "vault"
)

const (
	DefaultDuration = 10 * time.Second
	DefaultInterval = 5 * time.Second
)

type Config struct {
	Name          Name
	TmpDir        string
	Since         time.Time
	Until         time.Time
	OS            string
	DebugDuration time.Duration
	DebugInterval time.Duration
	HCL           *hcl.Product
}

type Product struct {
	l        hclog.Logger
	Name     Name
	Runners  []runner.Runner
	Excludes []string
	Selects  []string
	Config   Config
}

// Run iterates over the list of ops in a product and stores each runner into a map.
func (p *Product) Run() map[string]op.Op {
	p.l.Info("Running operations for", "product", p.Name)
	results := make(map[string]op.Op)
	for _, r := range p.Runners {
		p.l.Info("running operation", "product", p.Name, "runner", r.ID())
		o := r.Run()
		// NOTE(mkcp): There's nothing stopping Run() from being called multiple times, so we'll copy the runner off the product once it's done.
		// TODO(mkcp): It would be nice if we got an immutable runner result type back from runner runs instead.
		results[r.ID()] = o
		// Note runner errors to users and keep going.
		if o.Error != nil {
			p.l.Warn("result",
				"runner", r.ID(),
				"result", fmt.Sprintf("%s", o.Result),
				"error", o.Error,
			)
		}
	}
	return results
}

// Filter applies our slices of exclude and select runner.Identifier matchers to the set of the product's ops
func (p *Product) Filter() error {
	if p.Runners == nil {
		p.Runners = []runner.Runner{}
	}
	var err error
	// The presence of Selects takes precedence over Excludes
	if p.Selects != nil && 0 < len(p.Selects) {
		p.Runners, err = runner.Select(p.Selects, p.Runners)
		// Skip any Excludes
		return err
	}
	// No Selects, we can apply Excludes
	if p.Excludes != nil {
		p.Runners, err = runner.Exclude(p.Excludes, p.Runners)
	}
	return err
}

// CommanderHealthCheck employs the CLI to check if the client and then the agent are available.
func CommanderHealthCheck(client, agent string) error {
	checkClient := runner.NewCommander(client, "string").Run()
	if checkClient.Error != nil {
		return fmt.Errorf("client not available, healthcheck=%v, result=%v, error=%v", client, checkClient.Result, checkClient.Error)
	}
	checkAgent := runner.NewCommander(agent, "string").Run()
	if checkAgent.Error != nil {
		return fmt.Errorf("agent not available, healthcheck=%v, result=%v, error=%v", agent, checkAgent.Result, checkAgent.Error)
	}
	return nil
}

// CountRunners takes a map of product references and returns a count of all the runners
func CountRunners(products map[Name]*Product) int {
	var count int
	for _, product := range products {
		count += len(product.Runners)
	}
	return count
}

// DestructureHCL takes the collection of products and assigns them to vars
func DestructureHCL(products []*hcl.Product) (consulHCL, nomadHCL, tfeHCL, vaultHCL *hcl.Product) {
	for _, p := range products {
		switch p.Name {
		case string(Consul):
			consulHCL = p
		case string(Nomad):
			nomadHCL = p
		case string(TFE):
			tfeHCL = p
		case string(Vault):
			vaultHCL = p
		}
	}
	return consulHCL, nomadHCL, tfeHCL, vaultHCL
}
