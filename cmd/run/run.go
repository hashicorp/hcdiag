package run

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/cmd/help"
	"github.com/hashicorp/hcdiag/cmd/returns"
	"github.com/mitchellh/cli"

	"github.com/hashicorp/hcdiag/agent"
	"github.com/hashicorp/hcdiag/hcl"
	"github.com/hashicorp/hcdiag/product"
)

// seventyTwoHours represents the duration "72h" parsed in nanoseconds
const seventyTwoHours = 72 * time.Hour

// helpText is the short usage guidance shown under --help.
const helpText = `Usage: hcdiag run [options]

Executes an hcdiag diagnostics run on a local machine. Options are available to customize the execution.
`

// synopsis is provided in the help output of the enclosing scope, for example `hcdiag --help`.
const synopsis = `Execute an hcdiag diagnostic run`

var _ cli.Command = &cmd{}

type cmd struct {
	ui    cli.Ui
	flags *flag.FlagSet

	os     string
	serial bool
	dryrun bool

	// Products
	autoDetectProducts bool
	consul             bool
	nomad              bool
	tfe                bool
	vault              bool

	// since provides a time range for ops to work from
	since time.Duration

	// includeSince provides a time range for ops to work from
	includeSince time.Duration

	// includes
	includes []string

	// Bundle write location
	destination string

	// HCL file location
	config string

	// debugDuration param for product debug bundles
	debugDuration time.Duration

	// debugInterval param for product debug bundles
	debugInterval time.Duration
}

func (c *cmd) init() {
	// flag.ContinueOnError allows flag.Parse to return an error if one comes up, rather than doing an `os.Exit(2)`
	// on its own.
	c.flags = flag.NewFlagSet("run", flag.ContinueOnError)

	c.flags.BoolVar(&c.dryrun, "dryrun", false, dryrunUsageText)
	c.flags.BoolVar(&c.serial, "serial", false, serialUsageText)
	c.flags.BoolVar(&c.consul, "consul", false, consulUsageText)
	c.flags.BoolVar(&c.nomad, "nomad", false, nomadUsageText)
	c.flags.BoolVar(&c.tfe, "terraform-ent", false, terraformEntUsageText)
	c.flags.BoolVar(&c.vault, "vault", false, vaultUsageText)
	c.flags.BoolVar(&c.autoDetectProducts, "autodetect", true, autodetectUsageText)
	c.flags.DurationVar(&c.includeSince, "include-since", seventyTwoHours, includeSinceUsageText)
	c.flags.DurationVar(&c.since, "since", seventyTwoHours, sinceUsageText)
	c.flags.DurationVar(&c.debugDuration, "debug-duration", product.DefaultDuration, debugDurationUsageText)
	c.flags.DurationVar(&c.debugInterval, "debug-interval", product.DefaultInterval, debugIntervalUsageText)
	c.flags.StringVar(&c.os, "os", "auto", osUsageText)
	c.flags.StringVar(&c.destination, "destination", ".", destinationUsageText)
	c.flags.StringVar(&c.destination, "dest", ".", destUsageText)
	c.flags.StringVar(&c.config, "config", "", configUsageText)
	c.flags.Var(&CSVFlag{&c.includes}, "includes", includesUsageText)

	// Ensure f.Destination points to some kind of directory by its notation
	// FIXME(mkcp): trailing slashes should be trimmed in path.Dir... why does a double slash end in a slash?
	c.destination = path.Dir(c.destination)

	// When invalid flags are provided, Go will output a usage message of its own. If we direct our flag set to
	// io.Discard, it will effectively be hidden, allowing us to print our own Help message upon failure.
	c.flags.SetOutput(io.Discard)
}

// New produces a new *cmd pointer, initialized for use in a CLI application.
func New(ui cli.Ui) *cmd {
	c := &cmd{ui: ui}
	c.init()
	return c
}

// CommandFactory provides a cli.CommandFactory that will produce an appropriately-initiated *cmd.
func CommandFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		return New(ui), nil
	}
}

// Help provides help text to users who pass in the --help flag or who enter invalid options.
func (c *cmd) Help() string {
	return help.Usage(helpText, c.flags)
}

// Synopsis provides a brief description of the command, for inclusion in the application's primary --help.
func (c *cmd) Synopsis() string {
	return synopsis
}

// Run executes the command. On successful execution, it returns 0. On unsuccessful execution, a non-zero integer
// is returned instead.
func (c *cmd) Run(args []string) int {
	if err := c.parseFlags(args); err != nil {
		// Output the specific error to help the user understand what went wrong.
		c.ui.Warn(err.Error())
		// Since there was an issue in input, let's show our Help to try and assist the user.
		c.ui.Warn(c.Help())
		return returns.FlagParseError
	}

	l := configureLogging("hcdiag")

	// Build agent configuration from flags, HCL, and system time
	var config agent.Config
	// Parse and store HCL struct on agent.
	if c.config != "" {
		hclCfg, err := hcl.Parse(c.config)
		if err != nil {
			log.Fatalf("Failed to load configuration: %s", err)
		}
		l.Debug("HCL config is", "hcl", hclCfg)
		config.HCL = hclCfg
	}
	// Assign flag vals to our agent.Config
	cfg := c.mergeAgentConfig(config)

	// Set config timestamps based on durations
	now := time.Now()
	since := pickSinceVsIncludeSince(l, c.since, c.includeSince)
	cfg = setTime(cfg, now, since)
	l.Debug("merged cfg", "cfg", fmt.Sprintf("%+v", cfg))

	// Create agent
	a, err := agent.NewAgent(cfg, l)
	if err != nil {
		l.Error("problem creating agent", err)
		return returns.AgentSetupError
	}

	// Run the agent
	// NOTE(mkcp): Are there semantic returnCodes we can send based on the agent error type?
	errs := a.Run()
	if 0 < len(errs) {
		return returns.AgentExecutionError
	}

	// TODO (nwchandler): Pass in the client UI for output
	if err = a.WriteSummary(os.Stdout); err != nil {
		l.Warn("failed to generate report summary; please review output files to ensure everything expected is present", "err", err)
		return returns.AgentExecutionError
	}

	return returns.Success
}

// configureLogging takes a logger name, sets the default configuration, grabs the LOG_LEVEL from our ENV vars, and
//  returns a configured and usable logger.
func configureLogging(loggerName string) hclog.Logger {
	// Create logger, set default and log level
	appLogger := hclog.New(&hclog.LoggerOptions{
		Name:  loggerName,
		Color: hclog.AutoColor,
	})
	hclog.SetDefault(appLogger)
	if logStr := os.Getenv("LOG_LEVEL"); logStr != "" {
		if level := hclog.LevelFromString(logStr); level != hclog.NoLevel {
			appLogger.SetLevel(level)
			appLogger.Debug("Logger configuration change", "LOG_LEVEL", hclog.Fmt("%s", logStr))
		}
	}
	return hclog.Default()
}

type CSVFlag struct {
	Values *[]string
}

func (s CSVFlag) String() string {
	if s.Values == nil {
		return ""
	}
	return strings.Join(*s.Values, ",")
}

func (s CSVFlag) Set(v string) error {
	*s.Values = strings.Split(v, ",")
	return nil
}

func (c *cmd) parseFlags(args []string) error {
	return c.flags.Parse(args)
}

// mergeAgentConfig merges flags into the agent.Config, prioritizing flags over HCL config.
func (c *cmd) mergeAgentConfig(config agent.Config) agent.Config {
	config.OS = c.os
	config.Serial = c.serial
	config.Dryrun = c.dryrun

	config.Consul = c.consul
	config.Nomad = c.nomad
	config.TFE = c.tfe
	config.Vault = c.vault

	// If any products have been set manually, then we do not care about product auto-detection
	if c.autoDetectProducts && !checkProductsSet(config) {
		config.Consul = autoDetectCommand("consul")
		config.Nomad = autoDetectCommand("nomad")
		config.TFE = autoDetectCommand("terraform")
		config.Vault = autoDetectCommand("vault")

		if checkProductsSet(config) {
			hclog.L().Info(
				"Auto-detected products; if you wish to limit hcdiag, please use the appropriate -<product> flag and run again",
				"consul", config.Consul,
				"nomad", config.Nomad,
				"terraform", config.TFE,
				"vault", config.Vault,
			)
		}
	}

	// Params for --includes
	config.Includes = c.includes

	// Bundle write location
	config.Destination = c.destination

	// Apply Debug{Duration,Interval}
	config.DebugDuration = c.debugDuration
	config.DebugInterval = c.debugInterval

	return config
}

// checkProductsSet returns true if any of the individual products are true in the provided config
func checkProductsSet(config agent.Config) bool {
	return config.Consul || config.Nomad || config.TFE || config.Vault
}

func autoDetectCommand(cmd string) bool {
	p, err := exec.LookPath(cmd)
	if err != nil {
		return false
	}
	hclog.L().Debug("Found command", "cmd", cmd, "path", p)
	return true
}

// pickSinceVsIncludeSince if Since is default and IncludeSince is NOT default, use IncludeSince
func pickSinceVsIncludeSince(l hclog.Logger, since, includeSince time.Duration) time.Duration {
	if since == seventyTwoHours && includeSince != seventyTwoHours {
		l.Debug("includeSince set and default since", "includeSince", includeSince)
		return includeSince
	}
	return since
}

func setTime(cfg agent.Config, now time.Time, since time.Duration) agent.Config {
	// Capture a now value and set timestamps based on the same Now value
	// Get the difference between now and the provided --since Duration
	cfg.Since = now.Add(-since)
	// NOTE(mkcp): In the future, cfg.Until may be set by a flag.
	cfg.Until = time.Time{}

	return cfg
}
