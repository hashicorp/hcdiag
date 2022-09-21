package command

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/cli"

	"github.com/hashicorp/hcdiag/agent"
	"github.com/hashicorp/hcdiag/hcl"
	"github.com/hashicorp/hcdiag/product"
	"github.com/hashicorp/hcdiag/util"
)

// seventyTwoHours represents the duration "72h" parsed in nanoseconds
const seventyTwoHours = 72 * time.Hour

var _ cli.Command = &RunCommand{}

type RunCommand struct {
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

func (c *RunCommand) init() {
	const (
		consulUsageText        = "Run Consul diagnostics"
		nomadUsageText         = "Run Nomad diagnostics"
		terraformEntUsageText  = "Run Terraform Enterprise diagnostics"
		vaultUsageText         = "Run Vault diagnostics"
		autodetectUsageText    = "Auto-Detect installed products; any provided product flags will override this setting"
		dryrunUsageText        = "Displays all runners that would be executed during a normal run without actually executing them."
		serialUsageText        = "Run products in sequence rather than concurrently"
		includeSinceUsageText  = "Alias for -since, will be overridden if -since is also provided, usage examples: '72h', '25m', '45s', '120h1m90s'"
		sinceUsageText         = "Collect information within this time. Takes a 'go-formatted' duration, usage examples: '72h', '25m', '45s', '120h1m90s'"
		debugDurationUsageText = "How long to run product debug bundle commands. Provide a duration ex: '00h00m00s'. See: -duration in 'vault debug', 'consul debug', and 'nomad operator debug'"
		debugIntervalUsageText = "How long metrics collection intervals in product debug commands last. Provide a duration ex: '00h00m00s'. See: -interval in 'vault debug', 'consul debug', and 'nomad operator debug'"
		osUsageText            = "Override operating system detection"
		destinationUsageText   = "Path to the directory the bundle should be written in"
		destUsageText          = "Shorthand for -destination"
		configUsageText        = "Path to HCL configuration file"
		includesUsageText      = "Files or directories to include (comma-separated, file-*-globbing available if 'wrapped-*-in-single-quotes'); e.g. '/var/log/consul-*,/var/log/nomad-*'"
	)

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

// NewRunCommand produces a new *command pointer, initialized for use in a CLI application.
func NewRunCommand(ui cli.Ui) *RunCommand {
	c := &RunCommand{ui: ui}
	c.init()
	return c
}

// RunCommandFactory provides a cli.CommandFactory that will produce an appropriately-initiated *command.
func RunCommandFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		return NewRunCommand(ui), nil
	}
}

// Help provides help text to users who pass in the --help flag or who enter invalid options.
func (c *RunCommand) Help() string {
	helpText := `Usage: hcdiag run [options]

Executes an hcdiag diagnostics run on a local machine. Options are available to customize the execution.
`

	return Usage(helpText, c.flags)
}

// Synopsis provides a brief description of the command, for inclusion in the application's primary --help.
func (c *RunCommand) Synopsis() string {
	return "Execute an hcdiag diagnostic run"
}

// Run executes the command.
func (c *RunCommand) Run(args []string) int {
	if err := c.parseFlags(args); err != nil {
		// Output the specific error to help the user understand what went wrong.
		c.ui.Warn(err.Error())
		// Since there was an issue in input, let's show our Help to try and assist the user.
		c.ui.Warn(c.Help())
		return FlagParseError
	}

	l := configureLogging("hcdiag")

	// Build agent configuration from flags, HCL, and system time
	var config agent.Config
	// Parse and store HCL struct on agent.
	if c.config != "" {
		hclCfg, err := hcl.Parse(c.config)
		if err != nil {
			l.Error("Failed to load configuration", "config", c.config, "error", err)
			return ConfigError
		}
		l.Debug("HCL config is", "hcl", hclCfg)
		config.HCL = hclCfg
	}
	// Assign flag values to our agent.Config
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
		return AgentSetupError
	}

	// Run the agent
	errs := a.Run()
	if 0 < len(errs) {
		return AgentExecutionError
	}

	// TODO (nwchandler): Pass in the client UI for output
	if err = a.WriteSummary(os.Stdout); err != nil {
		l.Warn("failed to generate report summary; please review output files to ensure everything expected is present", "err", err)
		return AgentExecutionError
	}

	return Success
}

// configureLogging takes a logger name, sets the default configuration, grabs the LOG_LEVEL from our ENV vars, and
//
//	returns a configured and usable logger.
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

func (c *RunCommand) parseFlags(args []string) error {
	return c.flags.Parse(args)
}

// mergeAgentConfig merges flags into the agent.Config, prioritizing flags over HCL config.
func (c *RunCommand) mergeAgentConfig(config agent.Config) agent.Config {
	config.OS = c.os
	config.Serial = c.serial
	config.Dryrun = c.dryrun

	config.Consul = c.consul
	config.Nomad = c.nomad
	config.TFE = c.tfe
	config.Vault = c.vault

	// If any products have been set manually, then we do not care about product auto-detection
	if c.autoDetectProducts && !checkProductsSet(config) {
		config.Consul, _ = util.HostCommandExists("consul")
		config.Nomad, _ = util.HostCommandExists("nomad")
		config.TFE, _ = util.HostCommandExists("terraform")
		config.Vault, _ = util.HostCommandExists("vault")

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
