package debug

import (
	"fmt"
	"time"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = VaultDebug{}

// VaultDebugConfig is a config struct for VaultDebug runners
type VaultDebugConfig struct {
	Compress        string
	Duration        string
	Interval        string
	LogFormat       string
	MetricsInterval string
	Targets         []string
	Redactions      []*redact.Redact
}

// VaultDebug represents a VaultDebug runner
type VaultDebug struct {
	Compress        string           `json:"compress"`
	Duration        string           `json:"duration"`
	Interval        string           `json:"interval"`
	LogFormat       string           `json:"log_format"`
	MetricsInterval string           `json:"metrics_interval"`
	Targets         []string         `json:"targets"`
	Redactions      []*redact.Redact `json:"redactions"`

	output string
}

func (d VaultDebug) ID() string {
	return "VaultDebug"
}

func NewVaultDebug(cfg VaultDebugConfig, tmpDir string, debugDuration time.Duration, debugInterval time.Duration) *VaultDebug {
	// Create a pseudorandom string of characters to allow >1 VaultDebug runner without filename collisions
	randStr := randAlphanumString(4)

	dbg := VaultDebug{
		// No compression because the hcdiag bundle will get compressed anyway
		Compress: "true",
		// Use debug duration and interval
		Duration:        debugDuration.String(),
		Interval:        debugInterval.String(),
		LogFormat:       "standard",
		MetricsInterval: "10s",
		// Creates a subdirectory inside output dir
		output:     debugOutputPath(tmpDir, "VaultDebug", randStr),
		Targets:    cfg.Targets,
		Redactions: cfg.Redactions,
	}

	if len(cfg.Compress) > 0 {
		dbg.Compress = cfg.Compress
	}
	if dbg.Compress == "true" {
		dbg.output = dbg.output + ".tar.gz"
	}

	if len(cfg.Duration) > 0 {
		dbg.Duration = cfg.Duration
	}
	if len(cfg.Interval) > 0 {
		dbg.Interval = cfg.Interval
	}
	if len(cfg.LogFormat) > 0 {
		dbg.LogFormat = cfg.LogFormat
	}
	if len(cfg.MetricsInterval) > 0 {
		dbg.MetricsInterval = cfg.MetricsInterval
	}

	return &dbg
}

func (dbg VaultDebug) Run() op.Op {
	startTime := time.Now()

	filterString, err := productFilterString("vault", dbg.Targets)
	if err != nil {
		return op.New(dbg.ID(), map[string]any{}, op.Fail, err, runner.Params(dbg), startTime, time.Now())
	}

	// Assemble the vault debug command to execute
	cmdStr := vaultCmdString(dbg, filterString)

	// Create and set the Command
	cmd := runner.Command{
		Command:    cmdStr,
		Format:     "string",
		Redactions: dbg.Redactions,
	}

	o := cmd.Run()
	if o.Error != nil {
		return op.New(dbg.ID(), o.Result, op.Fail, o.Error, runner.Params(dbg), startTime, time.Now())
	}

	return op.New(dbg.ID(), o.Result, op.Success, nil, runner.Params(dbg), startTime, time.Now())
}

// vaultCmdString takes a VaultDebug and a filterString, and creates a valid Vault debug command string
func vaultCmdString(dbg VaultDebug, filterString string) string {
	return fmt.Sprintf(
		"vault debug -compress=%s -duration=%s -interval=%s -log-format=%s -metrics-interval=%s -output=%s%s",
		dbg.Compress,
		dbg.Duration,
		dbg.Interval,
		dbg.LogFormat,
		dbg.MetricsInterval,
		dbg.output,
		filterString,
	)
}

// debugOutputPath splices together an output path from a given tmpDir, debug output directory, and a random string
func debugOutputPath(tmpDir, dirname, randStr string) string {
	return fmt.Sprintf("%s/%s-%s", tmpDir, dirname, randStr)
}
