package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/agent"
	"github.com/hashicorp/hcdiag/product"
	"github.com/hashicorp/hcdiag/version"
)

// SeventyTwoHours represents the duration "72h" parsed in nanoseconds
const SeventyTwoHours = 72 * time.Hour

func main() {
	os.Exit(realMain())
}

func realMain() (returnCode int) {
	l := configureLogging("hcdiag")

	// Parse our CLI flags
	flags := Flags{}
	err := flags.parseFlags(os.Args[1:])
	if err != nil {
		return 64
	}
	// DEPRECATED(mkcp): Warn users if they're utilizing a deprecated flag
	if flags.AllProducts {
		l.Warn("DEPRECATED: -all will be removed in the future. Instead, provide multiple product flags.")
	}

	// If -version, skip agent setup and print version
	if flags.Version {
		v := version.GetVersion()
		fmt.Println(v.FullVersionNumber(true))
		return
	}

	// Build agent configuration from flags, HCL, and system time
	var config agent.Config
	// Set host and product config
	if flags.Config != "" {
		config, err = agent.ParseHCL(flags.Config)
		if err != nil {
			log.Fatalf("Failed to load configuration: %s", err)
		}
		l.Debug("Config is", "config", config)
	}
	// Assign flag vals to our agent.Config
	cfg := mergeAgentConfig(config, flags)

	// Set config timestamps based on durations
	now := time.Now()
	since := pickSinceVsIncludeSince(l, flags.Since, flags.IncludeSince)
	cfg = setTime(cfg, now, since)
	l.Debug("merged cfg", "cfg", fmt.Sprintf("%+v", cfg))

	// Create agent
	a := agent.NewAgent(cfg, l)

	// Run the agent
	// NOTE(mkcp): Are there semantic returnCodes we can send based on the agent error type?
	errs := a.Run()
	if 0 < len(errs) {
		return 1
	}
	return 0
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

// Flags stores our CLI inputs.
// TODO(mkcp): Add doccomments for flag fields (and organize them)
type Flags struct {
	OS     string
	Serial bool
	Dryrun bool

	// Products
	Consul      bool
	Nomad       bool
	TFE         bool
	Vault       bool
	AllProducts bool

	// Since provides a time range for seekers to work from
	Since time.Duration

	// IncludeSince provides a time range for seekers to work from
	IncludeSince time.Duration

	// Includes
	Includes []string

	// Bundle write location
	Destination string

	// HCL file location
	Config string

	// Get hcdiag version
	Version bool

	// Duration param for product debug bundles
	DebugDuration time.Duration

	// Interval param for product debug bundles
	DebugInterval time.Duration
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

func (f *Flags) parseFlags(args []string) error {
	flags := flag.NewFlagSet("hcdiag", flag.ExitOnError)
	flags.BoolVar(&f.Dryrun, "dryrun", false, "Performing a dry run will display all commands without executing them")
	flags.BoolVar(&f.Serial, "serial", false, "Run products in sequence rather than concurrently")
	flags.BoolVar(&f.Consul, "consul", false, "Run Consul diagnostics")
	flags.BoolVar(&f.Nomad, "nomad", false, "Run Nomad diagnostics")
	flags.BoolVar(&f.TFE, "terraform-ent", false, "(Experimental) Run Terraform Enterprise diagnostics")
	flags.BoolVar(&f.Vault, "vault", false, "Run Vault diagnostics")
	flags.BoolVar(&f.AllProducts, "all", false, "DEPRECATED: Run all available product diagnostics")
	flags.BoolVar(&f.Version, "version", false, "Print the current version of hcdiag")
	flags.DurationVar(&f.IncludeSince, "include-since", SeventyTwoHours, "Alias for -since, will be overridden if -since is also provided, usage examples: `72h`, `25m`, `45s`, `120h1m90s`")
	flags.DurationVar(&f.Since, "since", SeventyTwoHours, "Collect information within this time. Takes a 'go-formatted' duration, usage examples: `72h`, `25m`, `45s`, `120h1m90s`")
	flags.DurationVar(&f.DebugDuration, "debug-duration", product.DefaultDuration, "How long to run product debug bundle commands. Provide a duration ex: `00h00m00s`. See: -duration in `vault debug`, `consul debug`, and `nomad operator debug`")
	flags.DurationVar(&f.DebugInterval, "debug-interval", product.DefaultInterval, "How long metrics collection intervals in product debug commands last. Provide a duration ex: `00h00m00s`. See: -interval in `vault debug`, `consul debug`, and `nomad operator debug`")
	flags.StringVar(&f.OS, "os", "auto", "Override operating system detection")
	flags.StringVar(&f.Destination, "destination", ".", "Path to the directory the bundle should be written in")
	flags.StringVar(&f.Destination, "dest", ".", "Shorthand for -destination")
	flags.StringVar(&f.Config, "config", "", "Path to HCL configuration file")
	flags.Var(&CSVFlag{&f.Includes}, "includes", "files or directories to include (comma-separated, file-*-globbing available if 'wrapped-*-in-single-quotes')\ne.g. '/var/log/consul-*,/var/log/nomad-*'")

	// Ensure f.Destination points to some kind of directory by its notation
	// FIXME(mkcp): trailing slashes should be trimmed in path.Dir... why does a double slash end in a slash?
	f.Destination = path.Dir(f.Destination)

	return flags.Parse(args)
}

// mergeAgentConfig merges flags into the agent.Config, prioritizing flags over HCL config.
func mergeAgentConfig(config agent.Config, flags Flags) agent.Config {
	config.OS = flags.OS
	config.Serial = flags.Serial
	config.Dryrun = flags.Dryrun

	// DEPRECATED(mkcp): flags.AllProducts
	config.Consul = flags.AllProducts || flags.Consul
	// DEPRECATED(mkcp): flags.AllProducts
	config.Nomad = flags.AllProducts || flags.Nomad
	// DEPRECATED(mkcp): flags.AllProducts
	config.TFE = flags.AllProducts || flags.TFE
	// DEPRECATED(mkcp): flags.AllProducts
	config.Vault = flags.AllProducts || flags.Vault

	// Params for --includes
	config.Includes = flags.Includes

	// Bundle write location
	config.Destination = flags.Destination

	// Apply Debug{Duration,Interval}
	config.DebugDuration = flags.DebugDuration
	config.DebugInterval = flags.DebugInterval

	return config
}

// pickSinceVsIncludeSince if Since is default and IncludeSince is NOT default, use IncludeSince
func pickSinceVsIncludeSince(l hclog.Logger, since, includeSince time.Duration) time.Duration {
	if since == SeventyTwoHours && includeSince != SeventyTwoHours {
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
