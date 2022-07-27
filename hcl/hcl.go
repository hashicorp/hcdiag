package hcl

import (
	"fmt"
	"github.com/hashicorp/hcdiag/client"
	"github.com/hashicorp/hcdiag/runner"
	"github.com/hashicorp/hcdiag/runner/host"
	"github.com/hashicorp/hcdiag/runner/log"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"time"
)

type HCL struct {
	Host     *Host      `hcl:"host,block" json:"host"`
	Products []*Product `hcl:"product,block" json:"products"`
}

type Blocks interface {
	*Host | *Product
}

type Host struct {
	Commands     []Command     `hcl:"command,block"`
	Shells       []Shell       `hcl:"shell,block"`
	GETs         []GET         `hcl:"GET,block"`
	Copies       []Copy        `hcl:"copy,block"`
	DockerLogs   []DockerLog   `hcl:"docker-log,block"`
	JournaldLogs []JournaldLog `hcl:"journald-log,block"`
	Excludes     []string      `hcl:"excludes,optional"`
	Selects      []string      `hcl:"selects,optional"`
}

type Product struct {
	Name         string        `hcl:"name,label"`
	Commands     []Command     `hcl:"command,block"`
	Shells       []Shell       `hcl:"shell,block"`
	GETs         []GET         `hcl:"GET,block"`
	Copies       []Copy        `hcl:"copy,block"`
	DockerLogs   []DockerLog   `hcl:"docker-log,block"`
	JournaldLogs []JournaldLog `hcl:"journald-log,block"`
	Excludes     []string      `hcl:"excludes,optional"`
	Selects      []string      `hcl:"selects,optional"`
}

type Redact struct {
	Name    string `hcl:"name,label"`
	Replace string `hcl:"replace,optional"`
}

type Command struct {
	Run        string   `hcl:"run"`
	Format     string   `hcl:"format"`
	Redactions []Redact `hcl:"redact,block"`
}

type Shell struct {
	Run        string   `hcl:"run"`
	Redactions []Redact `hcl:"redact,block"`
}

type GET struct {
	Path       string   `hcl:"path"`
	Redactions []Redact `hcl:"redact,block"`
}

type Copy struct {
	Path       string   `hcl:"path"`
	Since      string   `hcl:"since,optional"`
	Redactions []Redact `hcl:"redact,block"`
}

type DockerLog struct {
	Container string `hcl:"container"`
	Since     string `hcl:"since,optional"`
}

type JournaldLog struct {
	Service string `hcl:"service"`
	Since   string `hcl:"since,optional"`
}

// Parse takes a file path and decodes the file from disk into HCL types.
func Parse(path string) (HCL, error) {
	var h HCL
	err := hclsimple.DecodeFile(path, nil, &h)
	if err != nil {
		return HCL{}, err
	}
	return h, nil
}

// BuildRunners steps through the HCLConfig structs and maps each runner config type to the corresponding New<Runner> function.
// All custom runners are reduced into a linear slice of runners and served back up to the product.
// No runners are returned if any config is invalid.
func BuildRunners[T Blocks](config T, tmpDir string, c *client.APIClient, since, until time.Time) ([]runner.Runner, error) {
	var dest string
	runners := make([]runner.Runner, 0)
	switch cfg := any(config).(type) {
	case *Product:
		// Set and validate the params that are different between Product and Host
		dest = tmpDir + "/" + cfg.Name
		if c == nil {
			return nil, fmt.Errorf("hcl.BuildRunners product received unexpected nil client, product=%s", cfg.Name)
		}

		// Build product's HTTPers
		runners = append(runners, mapProductGETs(cfg.GETs, c)...)

		// Identical code between Product and Host, but cfg's type must be resolved via the switch to access the fields
		// Build copiers
		copiers, err := mapCopies(cfg.Copies, dest)
		if err != nil {
			return nil, err
		}
		runners = append(runners, copiers...)

		// Build docker and journald logs
		dockerLogs, err := mapDockerLogs(cfg.DockerLogs, dest, since)
		if err != nil {
			return nil, err
		}
		runners = append(runners, dockerLogs...)
		journaldLogs, err := mapJournaldLogs(cfg.JournaldLogs, dest, since, until)
		if err != nil {
			return nil, err
		}
		runners = append(runners, journaldLogs...)

		// Build commanders and shellers
		commands, err := mapCommands(cfg.Commands)
		if err != nil {
			return nil, err
		}
		runners = append(runners, commands...)

		shells, err := mapShells(cfg.Shells)
		if err != nil {
			return nil, err
		}
		runners = append(runners, shells...)

	case *Host:
		// Set and validate the params that are different between Product and Host
		dest = tmpDir + "/host"
		if c != nil {
			return nil, fmt.Errorf("hcl.BuildRunners host received a client when nil expected, client=%v", c)
		}

		// Build host's HTTPers
		for _, g := range cfg.GETs {
			runners = append(runners, host.NewGetter(g.Path))
		}

		// Identical code between Product and Host, but cfg's type must be resolved via the switch
		// Build copiers
		copiers, err := mapCopies(cfg.Copies, dest)
		if err != nil {
			return nil, err
		}
		runners = append(runners, copiers...)

		// Build docker and journald logs
		dockerLogs, err := mapDockerLogs(cfg.DockerLogs, dest, since)
		if err != nil {
			return nil, err
		}
		runners = append(runners, dockerLogs...)
		journaldLogs, err := mapJournaldLogs(cfg.JournaldLogs, dest, since, until)
		if err != nil {
			return nil, err
		}
		runners = append(runners, journaldLogs...)

		// Build commanders and shellers
		commands, err := mapCommands(cfg.Commands)
		if err != nil {
			return nil, err
		}
		runners = append(runners, commands...)
		shells, err := mapShells(cfg.Shells)
		if err != nil {
			return nil, err
		}
		runners = append(runners, shells...)
	}
	return runners, nil
}

func mapCommands(cfgs []Command) ([]runner.Runner, error) {
	runners := make([]runner.Runner, len(cfgs))
	for i, c := range cfgs {
		err := ValidateRedactions(c.Redactions)
		if err != nil {
			return nil, err
		}
		runners[i] = runner.NewCommander(c.Run, c.Format)
	}
	return runners, nil
}

func mapShells(cfgs []Shell) ([]runner.Runner, error) {
	runners := make([]runner.Runner, len(cfgs))
	for i, c := range cfgs {
		err := ValidateRedactions(c.Redactions)
		if err != nil {
			return nil, err
		}
		runners[i] = runner.NewSheller(c.Run)
	}
	return runners, nil
}

func mapCopies(cfgs []Copy, dest string) ([]runner.Runner, error) {
	runners := make([]runner.Runner, len(cfgs))
	for i, c := range cfgs {
		var since time.Time

		err := ValidateRedactions(c.Redactions)
		if err != nil {
			return nil, err
		}

		// Set `from` with a timestamp
		if c.Since != "" {
			sinceDur, err := time.ParseDuration(c.Since)
			if err != nil {
				return nil, err
			}
			// Get the timestamp which marks the start of our duration
			// FIXME(mkcp): "Now" should be sourced from the agent to avoid clock sync issues
			since = time.Now().Add(-sinceDur)
		}
		runners[i] = runner.NewCopier(c.Path, dest, since, time.Time{})
	}
	return runners, nil
}

func mapProductGETs(cfgs []GET, c *client.APIClient) []runner.Runner {
	runners := make([]runner.Runner, len(cfgs))
	for i, g := range cfgs {
		runners[i] = runner.NewHTTPer(c, g.Path)
	}
	return runners
}

func mapDockerLogs(cfgs []DockerLog, dest string, since time.Time) ([]runner.Runner, error) {
	runners := make([]runner.Runner, len(cfgs))

	for i, d := range cfgs {
		if d.Since != "" {
			dur, err := time.ParseDuration(d.Since)
			if err != nil {
				return nil, err
			}
			// TODO(mkcp): Adding an agent.Now would help us pass an absolute now into this function. There's a subtle
			//  bug: different runners have different now values but it's unlikely to ever cause an issues for users.
			since = time.Now().Add(-dur)
		}
		runners[i] = log.NewDocker(d.Container, dest, since)
	}
	return runners, nil
}

func mapJournaldLogs(cfgs []JournaldLog, dest string, since, until time.Time) ([]runner.Runner, error) {
	runners := make([]runner.Runner, len(cfgs))

	for i, j := range cfgs {
		if j.Since != "" {
			dur, err := time.ParseDuration(j.Since)
			if err != nil {
				return nil, err
			}
			// TODO(mkcp): Adding an agent.Now would help us pass an absolute now into this function. There's a subtle
			//  bug: different runners have different now values but it's unlikely to ever cause an issues for users.
			now := time.Now()
			since = now.Add(dur)
			until = time.Time{}
		}
		runners[i] = log.NewJournald(j.Service, dest, since, until)
	}
	return runners, nil
}

// ProductsMap takes the collection of products and returns a map that keys each product to its Name.
func ProductsMap(products []*Product) map[string]*Product {
	m := make(map[string]*Product)
	for _, p := range products {
		m[p.Name] = p
	}
	return m
}

// ValidateRedactions takes a slice of redactions and ensures they match valid names.
func ValidateRedactions(redactions []Redact) error {
	set := map[string]bool{
		"regex":   true,
		"literal": true,
	}
	for _, r := range redactions {
		if !set[r.Name] {
			return fmt.Errorf("invalid redact, name=%s", r.Name)
		}
	}
	return nil
}
