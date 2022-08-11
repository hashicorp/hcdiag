package hcl

import (
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/client"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/runner"
	"github.com/hashicorp/hcdiag/runner/host"
	"github.com/hashicorp/hcdiag/runner/log"
	"github.com/hashicorp/hcl/v2/hclsimple"
)

type HCL struct {
	Host     *Host      `hcl:"host,block" json:"host"`
	Products []*Product `hcl:"product,block" json:"products"`
	Agent    *Agent     `hcl:"agent,block" json:"agent"`
}

type Blocks interface {
	*Host | *Product | *Agent
}

// NOTE(dcohen) this is currently a separate config block, as opposed to a parent block of the others
type Agent struct {
	Redactions []Redact `hcl:"redact,block"`
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
	Redactions   []Redact      `hcl:"redact,block"`
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
	Redactions   []Redact      `hcl:"redact,block"`
}

type Redact struct {
	Label   string `hcl:"name,label"`
	ID      string `hcl:"id,optional"`
	Match   string `hcl:"match"`
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
	Container  string   `hcl:"container"`
	Since      string   `hcl:"since,optional"`
	Redactions []Redact `hcl:"redact,block"`
}

type JournaldLog struct {
	Service    string   `hcl:"service"`
	Since      string   `hcl:"since,optional"`
	Redactions []Redact `hcl:"redact,block"`
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
func BuildRunners[T Blocks](config T, tmpDir string, c *client.APIClient, since, until time.Time, redactions []*redact.Redact) ([]runner.Runner, error) {
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
		gets, err := mapProductGETs(cfg.GETs, redactions, c)
		if err != nil {
			return nil, err
		}
		runners = append(runners, gets...)

		// Identical code between Product and Host, but cfg's type must be resolved via the switch to access the fields
		// Build copiers
		copiers, err := mapCopies(cfg.Copies, redactions, dest)
		if err != nil {
			return nil, err
		}
		runners = append(runners, copiers...)

		// Build docker and journald logs
		dockerLogs, err := mapDockerLogs(cfg.DockerLogs, dest, since, redactions)
		if err != nil {
			return nil, err
		}
		runners = append(runners, dockerLogs...)

		journaldLogs, err := mapJournaldLogs(cfg.JournaldLogs, dest, since, until, redactions)
		if err != nil {
			return nil, err
		}
		runners = append(runners, journaldLogs...)

		// Build commanders and shellers
		commands, err := mapCommands(cfg.Commands, redactions)
		if err != nil {
			return nil, err
		}
		runners = append(runners, commands...)

		shells, err := mapShells(cfg.Shells, redactions)
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
		gets, err := mapHostGets(cfg.GETs, redactions)
		if err != nil {
			return nil, err
		}
		runners = append(runners, gets...)

		// Identical code between Product and Host, but cfg's type must be resolved via the switch
		// Build copiers
		copiers, err := mapCopies(cfg.Copies, redactions, dest)
		if err != nil {
			return nil, err
		}
		runners = append(runners, copiers...)

		// Build docker and journald logs
		dockerLogs, err := mapDockerLogs(cfg.DockerLogs, dest, since, redactions)
		if err != nil {
			return nil, err
		}
		runners = append(runners, dockerLogs...)

		journaldLogs, err := mapJournaldLogs(cfg.JournaldLogs, dest, since, until, redactions)
		if err != nil {
			return nil, err
		}
		runners = append(runners, journaldLogs...)

		// Build commanders and shellers
		commands, err := mapCommands(cfg.Commands, redactions)
		if err != nil {
			return nil, err
		}
		runners = append(runners, commands...)

		shells, err := mapShells(cfg.Shells, redactions)
		if err != nil {
			return nil, err
		}
		runners = append(runners, shells...)
	}
	return runners, nil
}

func mapCommands(cfgs []Command, redactions []*redact.Redact) ([]runner.Runner, error) {
	runners := make([]runner.Runner, len(cfgs))
	for i, c := range cfgs {
		runnerRedacts, err := MapRedacts(c.Redactions)
		if err != nil {
			return nil, err
		}
		// Prepend runner-level redactions to those passed in
		runnerRedacts = append(runnerRedacts, redactions...)
		runners[i] = runner.NewCommander(c.Run, c.Format, runnerRedacts)
	}
	return runners, nil
}

func mapShells(cfgs []Shell, redactions []*redact.Redact) ([]runner.Runner, error) {
	runners := make([]runner.Runner, len(cfgs))
	for i, c := range cfgs {
		runnerRedacts, err := MapRedacts(c.Redactions)
		if err != nil {
			return nil, err
		}
		// Prepend runner-level redactions to those passed in
		runnerRedacts = append(runnerRedacts, redactions...)
		runners[i] = runner.NewSheller(c.Run, runnerRedacts)
	}
	return runners, nil
}

func mapCopies(cfgs []Copy, redactions []*redact.Redact, dest string) ([]runner.Runner, error) {
	runners := make([]runner.Runner, len(cfgs))
	for i, c := range cfgs {
		var since time.Time
		runnerRedacts, err := MapRedacts(c.Redactions)
		if err != nil {
			return nil, err
		}
		// Prepend runner-level redactions to those passed in
		runnerRedacts = append(runnerRedacts, redactions...)

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
		runners[i] = runner.NewCopier(c.Path, dest, since, time.Time{}, runnerRedacts)
	}
	return runners, nil
}

func mapProductGETs(cfgs []GET, redactions []*redact.Redact, c *client.APIClient) ([]runner.Runner, error) {
	runners := make([]runner.Runner, len(cfgs))
	for i, g := range cfgs {
		runnerRedacts, err := MapRedacts(g.Redactions)
		if err != nil {
			return nil, err
		}
		// Prepend runner-level redactions to those passed in
		runnerRedacts = append(runnerRedacts, redactions...)
		runners[i] = runner.NewHTTPer(c, g.Path, runnerRedacts)
	}
	return runners, nil
}

func mapHostGets(cfgs []GET, redactions []*redact.Redact) ([]runner.Runner, error) {
	runners := make([]runner.Runner, len(cfgs))
	for i, g := range cfgs {
		runnerRedacts, err := MapRedacts(g.Redactions)
		if err != nil {
			return nil, err
		}
		// Prepend runner-level redactions to those passed in
		runnerRedacts = append(runnerRedacts, redactions...)

		// TODO(mkcp): add redactions to host Get
		runners[i] = host.NewGetter(g.Path, runnerRedacts)
	}
	return runners, nil
}

func mapDockerLogs(cfgs []DockerLog, dest string, since time.Time, redactions []*redact.Redact) ([]runner.Runner, error) {
	runners := make([]runner.Runner, len(cfgs))

	for i, d := range cfgs {
		runnerRedacts, err := MapRedacts(d.Redactions)
		if err != nil {
			return nil, err
		}
		// Prepend runner-level redactions to those passed in
		runnerRedacts = append(runnerRedacts, redactions...)

		if d.Since != "" {
			dur, err := time.ParseDuration(d.Since)
			if err != nil {
				return nil, err
			}
			// TODO(mkcp): Adding an agent.Now would help us pass an absolute now into this function. There's a subtle
			//  bug: different runners have different now values but it's unlikely to ever cause an issues for users.
			since = time.Now().Add(-dur)
		}
		runners[i] = log.NewDocker(d.Container, dest, since, runnerRedacts)
	}
	return runners, nil
}

func mapJournaldLogs(cfgs []JournaldLog, dest string, since, until time.Time, redactions []*redact.Redact) ([]runner.Runner, error) {
	runners := make([]runner.Runner, len(cfgs))

	for i, j := range cfgs {
		runnerRedacts, err := MapRedacts(j.Redactions)
		if err != nil {
			return nil, err
		}
		// Prepend runner-level redactions to those passed in
		runnerRedacts = append(runnerRedacts, redactions...)

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
		runners[i] = log.NewJournald(j.Service, dest, since, until, runnerRedacts)
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

// MapRedacts maps HCL redactions to "real" `redact.Redact`s
func MapRedacts(redactions []Redact) ([]*redact.Redact, error) {
	err := ValidateRedactions(redactions)
	if err != nil {
		return nil, err
	}

	s := make([]*redact.Redact, len(redactions))
	for i, r := range redactions {
		cfg := redact.Config{
			Matcher: r.Match,
			ID:      r.ID,
			Replace: r.Replace,
		}
		red, err := redact.New(cfg)
		if err != nil {
			return nil, err
		}
		s[i] = red
	}
	return s, nil
}

// ValidateRedactions takes a slice of redactions and ensures they match valid names.
func ValidateRedactions(redactions []Redact) error {
	hclog.L().Trace("hcl.ValidateRedactions()", "redactions", redactions)
	for _, r := range redactions {
		switch r.Label {
		case "regex":
			_, err := regexp.Compile(r.Match)
			if err != nil {
				return fmt.Errorf("could not compile regex, matcher=%s, err=%s", r.Match, err)
			}
		case "literal":
			continue
		default:
			return fmt.Errorf("invalid redact name, name=%s", r.Label)
		}
	}
	return nil
}
