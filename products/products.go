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
	Name   string
	TmpDir string
	From   time.Time
	To     time.Time
	OS     string
}

type Name string

var (
	Consul Name = "consul"
	Host   Name = "host"
	Nomad  Name = "nomad"
	TFE    Name = "tfe"
	Vault  Name = "vault"
)

type Product struct {
	Name             Name
	Seekers          []*seeker.Seeker
	customCommanders []seeker.Commander
	customHTTPers    []seeker.HTTPer
	customCopiers    []seeker.Copier
	excludes         []string
	selects          []string
}

// Filter applies our slices of exclude and select seeker.Identifier matchers to the set of the product's seekers
func (p *Product) Filter() {
	if p.Seekers == nil {
		p.Seekers = []*seeker.Seeker{}
	}
	// The presence of selects takes precedence over excludes
	if p.selects != nil && 0 < len(p.selects) {
		p.Seekers = seeker.Select(p.selects, p.Seekers)
		// Skip any excludes
		return
	}
	// No selects, we can apply excludes
	if p.excludes != nil {
		p.Seekers = seeker.Exclude(p.excludes, p.Seekers)
	}
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
