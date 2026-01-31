// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package hcl

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/client"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/runner"
	"github.com/hashicorp/hcdiag/runner/debug"
	"github.com/hashicorp/hcdiag/runner/host"
	"github.com/hashicorp/hcdiag/runner/log"
	"github.com/hashicorp/hcl/v2/hclsimple"
)

type HCL struct {
	Host     *Host      `hcl:"host,block" json:"host,omitempty"`
	Products []*Product `hcl:"product,block" json:"products,omitempty"`
	Agent    *Agent     `hcl:"agent,block" json:"agent,omitempty"`
}

type Blocks interface {
	*Host | *Product | *Agent
}

type Agent struct {
	// NOTE(dcohen) this is currently a separate config block, as opposed to a parent block of the others
	Redactions []Redact `hcl:"redact,block" json:"redactions,omitempty"`
}

type Host struct {
	// Do
	Do  []Do  `hcl:"do,block" json:"do,omitempty"`
	Seq []Seq `hcl:"seq,block" json:"seq,omitempty"`

	// Runners
	Commands     []Command     `hcl:"command,block" json:"commands,omitempty"`
	Shells       []Shell       `hcl:"shell,block" json:"shells,omitempty"`
	GETs         []GET         `hcl:"GET,block" json:"gets,omitempty"`
	Copies       []Copy        `hcl:"copy,block" json:"copies,omitempty"`
	DockerLogs   []DockerLog   `hcl:"docker-log,block" json:"docker_log,omitempty"`
	JournaldLogs []JournaldLog `hcl:"journald-log,block" json:"journald_log,omitempty"`

	// Filters
	Excludes []string `hcl:"excludes,optional" json:"excludes,omitempty"`
	Selects  []string `hcl:"selects,optional" json:"selects,omitempty"`

	// Params
	Redactions []Redact `hcl:"redact,block" json:"redactions,omitempty"`
}

type Product struct {
	Name string `hcl:"name,label" json:"name"`

	// Do
	Do  []Do  `hcl:"do,block" json:"do,omitempty"`
	Seq []Seq `hcl:"seq,block" json:"seq,omitempty"`

	// Runners
	Commands     []Command     `hcl:"command,block" json:"commands,omitempty"`
	Shells       []Shell       `hcl:"shell,block" json:"shells,omitempty"`
	GETs         []GET         `hcl:"GET,block" json:"gets,omitempty"`
	Copies       []Copy        `hcl:"copy,block" json:"copies,omitempty"`
	DockerLogs   []DockerLog   `hcl:"docker-log,block" json:"docker_log,omitempty"`
	JournaldLogs []JournaldLog `hcl:"journald-log,block" json:"journald_log,omitempty"`
	VaultDebugs  []VaultDebug  `hcl:"vault-debug,block" json:"vault_debug,omitempty"`
	ConsulDebugs []ConsulDebug `hcl:"consul-debug,block" json:"consul_debug,omitempty"`
	NomadDebugs  []NomadDebug  `hcl:"nomad-debug,block" json:"nomad_debug,omitempty"`
	Excludes     []string      `hcl:"excludes,optional" json:"excludes,omitempty"`
	Selects      []string      `hcl:"selects,optional" json:"selects,omitempty"`
	Redactions   []Redact      `hcl:"redact,block" json:"redactions,omitempty"`
}

type Do struct {
	Label       string `hcl:"name,label" json:"label"`
	Description string `hcl:"description,optional" json:"since"`

	Do  []Do  `hcl:"do,block" json:"do,omitempty"`
	Seq []Seq `hcl:"seq,block" json:"seq,omitempty"`

	// Runners
	Commands     []Command     `hcl:"command,block" json:"commands,omitempty"`
	Shells       []Shell       `hcl:"shell,block" json:"shells,omitempty"`
	GETs         []GET         `hcl:"GET,block" json:"gets,omitempty"`
	Copies       []Copy        `hcl:"copy,block" json:"copies,omitempty"`
	DockerLogs   []DockerLog   `hcl:"docker-log,block" json:"docker_log,omitempty"`
	JournaldLogs []JournaldLog `hcl:"journald-log,block" json:"journald_log,omitempty"`
	VaultDebugs  []VaultDebug  `hcl:"vault-debug,block" json:"vault_debug,omitempty"`
	ConsulDebugs []ConsulDebug `hcl:"consul-debug,block" json:"consul_debug,omitempty"`
	NomadDebugs  []NomadDebug  `hcl:"nomad-debug,block" json:"nomad_debug,omitempty"`

	// Filters
	// Excludes     []string      `hcl:"excludes,optional" json:"excludes,omitempty"`
	// Selects      []string      `hcl:"selects,optional" json:"selects,omitempty"`

	// Params
	// Redactions   []Redact      `hcl:"redact,block" json:"redactions,omitempty"`
}

type Seq struct {
	Label       string `hcl:"name,label" json:"label"`
	Description string `hcl:"description,optional" json:"since"`
	Timeout     string `hcl:"timeout,optional" json:"timeout,omitempty"`

	// Do
	Do  []Do  `hcl:"do,block" json:"do,omitempty"`
	Seq []Seq `hcl:"seq,block" json:"seq,omitempty"`

	// Runners
	Commands     []Command     `hcl:"command,block" json:"commands,omitempty"`
	Shells       []Shell       `hcl:"shell,block" json:"shells,omitempty"`
	GETs         []GET         `hcl:"GET,block" json:"gets,omitempty"`
	Copies       []Copy        `hcl:"copy,block" json:"copies,omitempty"`
	DockerLogs   []DockerLog   `hcl:"docker-log,block" json:"docker_log,omitempty"`
	JournaldLogs []JournaldLog `hcl:"journald-log,block" json:"journald_log,omitempty"`
	VaultDebugs  []VaultDebug  `hcl:"vault-debug,block" json:"vault_debug,omitempty"`
	ConsulDebugs []ConsulDebug `hcl:"consul-debug,block" json:"consul_debug,omitempty"`
	NomadDebugs  []NomadDebug  `hcl:"nomad-debug,block" json:"nomad_debug,omitempty"`

	// Filters
	// Excludes     []string      `hcl:"excludes,optional" json:"excludes,omitempty"`
	// Selects      []string      `hcl:"selects,optional" json:"selects,omitempty"`

	// Params
	// Redactions   []Redact      `hcl:"redact,block" json:"redactions,omitempty"`
}

type Redact struct {
	Label   string `hcl:"name,label" json:"label"`
	ID      string `hcl:"id,optional" json:"id"`
	Match   string `hcl:"match" json:"-"`
	Replace string `hcl:"replace,optional" json:"replace"`
}

type Command struct {
	Run        string   `hcl:"run" json:"run"`
	Format     string   `hcl:"format" json:"format"`
	Redactions []Redact `hcl:"redact,block" json:"redactions,omitempty"`
	Timeout    string   `hcl:"timeout,optional" json:"timeout,omitempty"`
}

type Shell struct {
	Run        string   `hcl:"run" json:"run"`
	Timeout    string   `hcl:"timeout,optional" json:"timeout,omitempty"`
	Redactions []Redact `hcl:"redact,block" json:"redactions,omitempty"`
}

type GET struct {
	Path       string   `hcl:"path" json:"path"`
	Redactions []Redact `hcl:"redact,block" json:"redactions,omitempty"`
	Timeout    string   `hcl:"timeout,optional" json:"timeout,omitempty"`
}

type Copy struct {
	Path       string   `hcl:"path" json:"path"`
	Since      string   `hcl:"since,optional" json:"since"`
	Redactions []Redact `hcl:"redact,block" json:"redactions,omitempty"`
	Timeout    string   `hcl:"timeout,optional" json:"timeout,omitempty"`
}

type DockerLog struct {
	Container  string   `hcl:"container" json:"container"`
	Since      string   `hcl:"since,optional" json:"since"`
	Redactions []Redact `hcl:"redact,block" json:"redactions,omitempty"`
	Timeout    string   `hcl:"timeout,optional" json:"timeout,omitempty"`
}

type JournaldLog struct {
	Service    string   `hcl:"service" json:"service"`
	Since      string   `hcl:"since,optional" json:"since"`
	Redactions []Redact `hcl:"redact,block" json:"redactions,omitempty"`
	Timeout    string   `hcl:"timeout,optional" json:"timeout,omitempty"`
}

type VaultDebug struct {
	Compress        string   `hcl:"compress" json:"compress"`
	Duration        string   `hcl:"duration,optional" json:"duration"`
	Interval        string   `hcl:"interval,optional" json:"interval"`
	LogFormat       string   `hcl:"log-format,optional" json:"log_format"`
	MetricsInterval string   `hcl:"metrics-interval,optional" json:"metrics_interval"`
	Targets         []string `hcl:"targets,optional" json:"targets"`
	Redactions      []Redact `hcl:"redact,block" json:"redactions"`
}

type ConsulDebug struct {
	Archive    string   `hcl:"archive" json:"archive"`
	Duration   string   `hcl:"duration,optional" json:"duration"`
	Interval   string   `hcl:"interval,optional" json:"interval"`
	Captures   []string `hcl:"captures,optional" json:"captures"`
	Redactions []Redact `hcl:"redact,block" json:"redactions"`
}

type NomadDebug struct {
	Duration      string `hcl:"duration,optional" json:"duration"`
	Interval      string `hcl:"interval,optional" json:"interval"`
	LogLevel      string `hcl:"log-level,optional" json:"log_level"`
	MaxNodes      int    `hcl:"max-nodes,optional" json:"max_nodes"`
	NodeClass     string `hcl:"node-class,optional" json:"node_class"`
	NodeID        string `hcl:"node-id,optional" json:"node_id"`
	PprofDuration string `hcl:"pprof-duration,optional" json:"pprof_duration"`
	PprofInterval string `hcl:"pprof-interval,optional" json:"pprof_interval"`
	ServerID      string `hcl:"server-id,optional" json:"server_id"`
	Stale         bool   `hcl:"stale,optional" json:"stale"`
	Verbose       bool   `hcl:"verbose,optional" json:"verbose"`

	EventTopic []string `hcl:"targets,optional" json:"targets"`
	Redactions []Redact `hcl:"redact,block" json:"redactions"`
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
func BuildRunners[T Blocks](config T, tmpDir string, debugDuration time.Duration, debugInterval time.Duration, c *client.APIClient, since, until time.Time, redactions []*redact.Redact) ([]runner.Runner, error) {
	return BuildRunnersWithContext(context.Background(), config, tmpDir, debugDuration, debugInterval, c, since, until, redactions)
}

// BuildRunnersWithContext is similar to BuildRunners but accepts a context.Context that will be passed into the runners.
func BuildRunnersWithContext[T Blocks](ctx context.Context, config T, tmpDir string, debugDuration time.Duration, debugInterval time.Duration, c *client.APIClient, since, until time.Time, redactions []*redact.Redact) ([]runner.Runner, error) {
	var dest string
	runners := make([]runner.Runner, 0)

	switch cfg := any(config).(type) {
	case *Product:
		// Set and validate the params that are different between Product and Host
		dest = tmpDir + "/" + cfg.Name
		if c == nil {
			return nil, fmt.Errorf("hcl.BuildRunners product received unexpected nil client, product=%s", cfg.Name)
		}

		// Build product's HTTPs
		gets, err := mapProductGETs(ctx, cfg.GETs, redactions, c)
		if err != nil {
			return nil, err
		}
		runners = append(runners, gets...)

		// Identical code between Product and Host, but cfg's type must be resolved via the switch to access the fields
		// Build Copy runners
		copies, err := mapCopies(ctx, cfg.Copies, redactions, dest)
		if err != nil {
			return nil, err
		}
		runners = append(runners, copies...)

		// Build docker and journald logs
		dockerLogs, err := mapDockerLogs(ctx, cfg.DockerLogs, dest, since, redactions)
		if err != nil {
			return nil, err
		}
		runners = append(runners, dockerLogs...)

		journaldLogs, err := mapJournaldLogs(ctx, cfg.JournaldLogs, dest, since, until, redactions)
		if err != nil {
			return nil, err
		}
		runners = append(runners, journaldLogs...)

		// Build debug runners
		vaultDebugs, err := mapVaultDebugs(ctx, cfg.VaultDebugs, tmpDir, debugDuration, debugInterval, redactions)
		if err != nil {
			return nil, err
		}
		runners = append(runners, vaultDebugs...)

		consulDebugs, err := mapConsulDebugs(ctx, cfg.ConsulDebugs, tmpDir, debugDuration, debugInterval, redactions)
		if err != nil {
			return nil, err
		}
		runners = append(runners, consulDebugs...)

		nomadDebugs, err := mapNomadDebugs(ctx, cfg.NomadDebugs, tmpDir, debugDuration, debugInterval, redactions)
		if err != nil {
			return nil, err
		}
		runners = append(runners, nomadDebugs...)

		// Build commands and shells
		commands, err := mapCommands(ctx, cfg.Commands, redactions)
		if err != nil {
			return nil, err
		}
		runners = append(runners, commands...)

		shells, err := mapShells(ctx, cfg.Shells, redactions)
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

		// Build host's HTTPs
		gets, err := mapHostGets(ctx, cfg.GETs, redactions)
		if err != nil {
			return nil, err
		}
		runners = append(runners, gets...)

		// Identical code between Product and Host, but cfg's type must be resolved via the switch
		// Build Copy runners
		copies, err := mapCopies(ctx, cfg.Copies, redactions, dest)
		if err != nil {
			return nil, err
		}
		runners = append(runners, copies...)

		// Build docker and journald logs
		dockerLogs, err := mapDockerLogs(ctx, cfg.DockerLogs, dest, since, redactions)
		if err != nil {
			return nil, err
		}
		runners = append(runners, dockerLogs...)

		journaldLogs, err := mapJournaldLogs(ctx, cfg.JournaldLogs, dest, since, until, redactions)
		if err != nil {
			return nil, err
		}
		runners = append(runners, journaldLogs...)

		// Build commands and shells
		commands, err := mapCommands(ctx, cfg.Commands, redactions)
		if err != nil {
			return nil, err
		}
		runners = append(runners, commands...)

		shells, err := mapShells(ctx, cfg.Shells, redactions)
		if err != nil {
			return nil, err
		}
		runners = append(runners, shells...)
	}
	return runners, nil
}

func mapCommands(ctx context.Context, cfgs []Command, redactions []*redact.Redact) ([]runner.Runner, error) {
	runners := make([]runner.Runner, len(cfgs))
	for i, c := range cfgs {
		runnerRedacts, err := MapRedacts(c.Redactions)
		if err != nil {
			return nil, err
		}
		// Prepend runner-level redactions to those passed in
		runnerRedacts = append(runnerRedacts, redactions...)
		var timeout time.Duration
		if c.Timeout != "" {
			timeout, err = time.ParseDuration(c.Timeout)
			if err != nil {
				return nil, err
			}
		}
		r, err := runner.NewCommandWithContext(ctx, runner.CommandConfig{
			Command:    c.Run,
			Format:     c.Format,
			Timeout:    timeout,
			Redactions: runnerRedacts,
		})
		if err != nil {
			return nil, err
		}
		runners[i] = r
	}
	return runners, nil
}

func mapShells(ctx context.Context, cfgs []Shell, redactions []*redact.Redact) ([]runner.Runner, error) {
	runners := make([]runner.Runner, len(cfgs))
	for i, c := range cfgs {
		runnerRedacts, err := MapRedacts(c.Redactions)
		if err != nil {
			return nil, err
		}
		// Prepend runner-level redactions to those passed in
		runnerRedacts = append(runnerRedacts, redactions...)
		var timeout time.Duration
		if c.Timeout != "" {
			timeout, err = time.ParseDuration(c.Timeout)
			if err != nil {
				return nil, err
			}
		}
		if err != nil {
			return nil, err
		}
		s, err := runner.NewShellWithContext(ctx, runner.ShellConfig{
			Command:    c.Run,
			Redactions: runnerRedacts,
			Timeout:    timeout,
		})
		if err != nil {
			return nil, err
		}
		runners[i] = s
	}
	return runners, nil
}

func mapCopies(ctx context.Context, cfgs []Copy, redactions []*redact.Redact, dest string) ([]runner.Runner, error) {
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
		var timeout time.Duration
		if c.Timeout != "" {
			timeout, err = time.ParseDuration(c.Timeout)
			if err != nil {
				return nil, err
			}
		}
		r, err := runner.NewCopyWithContext(ctx, runner.CopyConfig{
			Path:       c.Path,
			DestDir:    dest,
			Since:      since,
			Until:      time.Time{},
			Redactions: runnerRedacts,
			Timeout:    timeout,
		})
		if err != nil {
			return nil, err
		}
		runners[i] = r
	}
	return runners, nil
}

func mapProductGETs(ctx context.Context, cfgs []GET, redactions []*redact.Redact, c *client.APIClient) ([]runner.Runner, error) {
	runners := make([]runner.Runner, len(cfgs))
	for i, g := range cfgs {
		runnerRedacts, err := MapRedacts(g.Redactions)
		if err != nil {
			return nil, err
		}
		// Prepend runner-level redactions to those passed in
		runnerRedacts = append(runnerRedacts, redactions...)
		var timeout time.Duration
		if g.Timeout != "" {
			timeout, err = time.ParseDuration(g.Timeout)
			if err != nil {
				return nil, err
			}
		}
		r, err := runner.NewHTTPWithContext(ctx, runner.HttpConfig{
			Client:     c,
			Path:       g.Path,
			Timeout:    timeout,
			Redactions: runnerRedacts,
		})
		if err != nil {
			return nil, err
		}
		runners[i] = r
	}
	return runners, nil
}

func mapHostGets(ctx context.Context, cfgs []GET, redactions []*redact.Redact) ([]runner.Runner, error) {
	runners := make([]runner.Runner, len(cfgs))
	for i, g := range cfgs {
		runnerRedacts, err := MapRedacts(g.Redactions)
		if err != nil {
			return nil, err
		}
		// Prepend runner-level redactions to those passed in
		runnerRedacts = append(runnerRedacts, redactions...)
		var timeout time.Duration
		if g.Timeout != "" {
			timeout, err = time.ParseDuration(g.Timeout)
			if err != nil {
				return nil, err
			}
		}
		r, err := host.NewGetWithContext(ctx, host.GetConfig{
			Path:       g.Path,
			Timeout:    timeout,
			Redactions: runnerRedacts,
		})
		if err != nil {
			return nil, err
		}

		runners[i] = r
	}
	return runners, nil
}

func mapDockerLogs(ctx context.Context, cfgs []DockerLog, dest string, since time.Time, redactions []*redact.Redact) ([]runner.Runner, error) {
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
			// bug: different runners have different now values but it's unlikely to ever cause an issues for users.
			since = time.Now().Add(-dur)
		}
		var timeout time.Duration
		if d.Timeout != "" {
			timeout, err = time.ParseDuration(d.Timeout)
			if err != nil {
				return nil, err
			}
		}
		runners[i] = log.NewDockerWithContext(ctx, log.DockerConfig{
			Container:  d.Container,
			DestDir:    dest,
			Since:      since,
			Redactions: runnerRedacts,
			Timeout:    timeout,
		})
	}
	return runners, nil
}

func mapJournaldLogs(ctx context.Context, cfgs []JournaldLog, dest string, since, until time.Time, redactions []*redact.Redact) ([]runner.Runner, error) {
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
			// bug: different runners have different now values but it's unlikely to ever cause an issues for users.
			now := time.Now()
			since = now.Add(dur)
			until = time.Time{}
		}
		var timeout time.Duration
		if j.Timeout != "" {
			timeout, err = time.ParseDuration(j.Timeout)
			if err != nil {
				return nil, err
			}
		}
		runners[i] = log.NewJournaldWithContext(ctx, log.JournaldConfig{
			Service:    j.Service,
			DestDir:    dest,
			Since:      since,
			Until:      until,
			Redactions: runnerRedacts,
			Timeout:    timeout,
		})
	}
	return runners, nil
}

func mapVaultDebugs(ctx context.Context, cfgs []VaultDebug, tmpDir string, debugDuration time.Duration, debugInterval time.Duration, redactions []*redact.Redact) ([]runner.Runner, error) {
	runners := make([]runner.Runner, len(cfgs))

	for i, d := range cfgs {
		runnerRedacts, err := MapRedacts(d.Redactions)
		if err != nil {
			return nil, err
		}
		// Prepend runner-level redactions to those passed in
		runnerRedacts = append(runnerRedacts, redactions...)

		// Create the runner
		cfg := debug.VaultDebugConfig{
			Compress:        d.Compress,
			Duration:        d.Duration,
			Interval:        d.Interval,
			LogFormat:       d.LogFormat,
			MetricsInterval: d.MetricsInterval,
			Targets:         d.Targets,
			Redactions:      runnerRedacts,
		}

		dbg, err := debug.NewVaultDebug(cfg, tmpDir, debugDuration, debugInterval)
		if err != nil {
			return nil, err
		}
		runners[i] = dbg
	}
	return runners, nil
}

func mapConsulDebugs(ctx context.Context, cfgs []ConsulDebug, tmpDir string, debugDuration time.Duration, debugInterval time.Duration, redactions []*redact.Redact) ([]runner.Runner, error) {
	runners := make([]runner.Runner, len(cfgs))

	for i, d := range cfgs {
		runnerRedacts, err := MapRedacts(d.Redactions)
		if err != nil {
			return nil, err
		}
		// Prepend runner-level redactions to those passed in
		runnerRedacts = append(runnerRedacts, redactions...)

		// Create the runner
		cfg := debug.ConsulDebugConfig{
			Archive:    d.Archive,
			Duration:   d.Duration,
			Interval:   d.Interval,
			Captures:   d.Captures,
			Redactions: runnerRedacts,
		}
		dbg, err := debug.NewConsulDebug(cfg, tmpDir, debugDuration, debugInterval)
		if err != nil {
			return nil, err
		}
		runners[i] = dbg
	}
	return runners, nil
}

func mapNomadDebugs(ctx context.Context, cfgs []NomadDebug, tmpDir string, debugDuration time.Duration, debugInterval time.Duration, redactions []*redact.Redact) ([]runner.Runner, error) {
	runners := make([]runner.Runner, len(cfgs))

	for i, d := range cfgs {
		runnerRedacts, err := MapRedacts(d.Redactions)
		if err != nil {
			return nil, err
		}
		// Prepend runner-level redactions to those passed in
		runnerRedacts = append(runnerRedacts, redactions...)

		// Create the runner
		cfg := debug.NomadDebugConfig{
			Duration:      d.Duration,
			Interval:      d.Interval,
			LogLevel:      d.LogLevel,
			MaxNodes:      d.MaxNodes,
			NodeClass:     d.NodeClass,
			NodeID:        d.NodeID,
			PprofDuration: d.PprofDuration,
			PprofInterval: d.PprofInterval,
			ServerID:      d.ServerID,
			Stale:         d.Stale,
			Verbose:       d.Verbose,
			EventTopic:    d.EventTopic,
			Redactions:    runnerRedacts,
		}
		dbg, err := debug.NewNomadDebug(cfg, tmpDir, debugDuration, debugInterval)
		if err != nil {
			return nil, err
		}
		runners[i] = dbg
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
		// TODO(mkcp): Implement literals and `switch r.Label {}`
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
		// TODO(mkcp): Validate literals when they are implemented
		// case "literal":
		// 	continue
		default:
			return fmt.Errorf("invalid redact name, name=%s", r.Label)
		}
	}
	return nil
}
