package hcl

import (
	"fmt"
	"github.com/hashicorp/hcdiag/client"
	"github.com/hashicorp/hcdiag/runner"
	"github.com/hashicorp/hcdiag/runner/host"
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
	Commands []Command `hcl:"command,block"`
	Shells   []Shell   `hcl:"shell,block"`
	GETs     []GET     `hcl:"GET,block"`
	Copies   []Copy    `hcl:"copy,block"`
	Excludes []string  `hcl:"excludes,optional"`
	Selects  []string  `hcl:"selects,optional"`
}

type Product struct {
	Name     string    `hcl:"name,label"`
	Commands []Command `hcl:"command,block"`
	Shells   []Shell   `hcl:"shell,block"`
	GETs     []GET     `hcl:"GET,block"`
	Copies   []Copy    `hcl:"copy,block"`
	Excludes []string  `hcl:"excludes,optional"`
	Selects  []string  `hcl:"selects,optional"`
}

type Command struct {
	Run    string `hcl:"run"`
	Format string `hcl:"format"`
}

type Shell struct {
	Run string `hcl:"run"`
}

type GET struct {
	Path string `hcl:"path"`
}

type Copy struct {
	Path  string `hcl:"path"`
	Since string `hcl:"since,optional"`
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
func BuildRunners[T Blocks](config T, tmpDir string, c *client.APIClient) ([]runner.Runner, error) {
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

		// Build commanders and shellers
		runners = append(runners, mapCommands(cfg.Commands)...)
		runners = append(runners, mapShells(cfg.Shells)...)

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

		// Build commanders and shellers
		runners = append(runners, mapCommands(cfg.Commands)...)
		runners = append(runners, mapShells(cfg.Shells)...)
	}
	return runners, nil
}

func mapCommands(cfgs []Command) []runner.Runner {
	runners := make([]runner.Runner, len(cfgs))
	for i, c := range cfgs {
		runners[i] = runner.NewCommander(c.Run, c.Format)
	}
	return runners
}

func mapShells(cfgs []Shell) []runner.Runner {
	runners := make([]runner.Runner, len(cfgs))
	for i, c := range cfgs {
		runners[i] = runner.NewSheller(c.Run)
	}
	return runners
}

func mapCopies(cfgs []Copy, dest string) ([]runner.Runner, error) {
	runners := make([]runner.Runner, len(cfgs))
	for i, c := range cfgs {
		var since time.Time

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
