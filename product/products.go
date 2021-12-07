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
	DefaultDebugSeconds    = 10
	DefaultIntervalSeconds = 5
)

type Config struct {
	Logger *hclog.Logger
	Name   string
	TmpDir string
	From   time.Time
	To     time.Time
	OS     string
}

type Product struct {
	Name     string
	Seekers  []*seeker.Seeker
	Excludes []string
	Selects  []string
}

// Filter applies our slices of exclude and select seeker.Identifier matchers to the set of the product's seekers
func (p *Product) Filter() error {
	if p.Seekers == nil {
		p.Seekers = []*seeker.Seeker{}
	}

	var filter string
	var set []string
	// The presence of Selects takes precedence over Excludes
	if p.Selects != nil && 0 < len(p.Selects) {
		filter = "select"
		set = p.Selects
	} else if p.Excludes != nil {
		filter = "exclude"
		set = p.Excludes
	} else {
		return nil
	}

	filtered, err := seeker.Filter(filter, set, p.Seekers)
	if err != nil {
		return err
	}
	p.Seekers = filtered

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

func CountSeekers(products map[string]*Product) int {
	var count int
	for _, product := range products {
		count += len(product.Seekers)
	}
	return count
}
