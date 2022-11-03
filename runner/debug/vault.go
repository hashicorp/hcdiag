package debug

import (
	"fmt"
	"time"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/product"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = VaultDebug{}

// VaultDebugConfig is a config struct for VaultDebug runners
type VaultDebugConfig struct {
	ProductConfig product.Config
	// No compression because the hcdiag bundle will get compressed anyway
	Compress        string
	Duration        string
	Interval        string
	LogFormat       string
	MetricsInterval string
	Output          string
	Targets         []string
	Redactions      []*redact.Redact
}

// VaultDebug represents a VaultDebug runner
type VaultDebug struct {
	Compress        string           `json:"compress"`
	Duration        string           `json:"duration"`
	Interval        string           `json:"interval"`
	LogFormat       string           `json:"logformat"`
	MetricsInterval string           `json:"metricsinterval"`
	Output          string           `json:"output"`
	Targets         []string         `json:"targets"`
	Redactions      []*redact.Redact `json:"redactions"`
}

func (d VaultDebug) ID() string {
	return "VaultDebug"
}

func NewVaultDebug(cfg VaultDebugConfig) *VaultDebug {
	dbg := VaultDebug{
		// No compression because the hcdiag bundle will get compressed anyway
		Compress:        "false",
		Duration:        "2m",
		Interval:        "30s",
		LogFormat:       "standard",
		MetricsInterval: "10s",
		Output:          fmt.Sprintf("%s/VaultDebug", cfg.ProductConfig.TmpDir),
		Targets:         cfg.Targets,
		Redactions:      cfg.Redactions,
	}

	if len(cfg.Compress) > 0 {
		dbg.Compress = cfg.Compress
		if dbg.Compress == "true" {
			dbg.Output = dbg.Output + ".tar.gz"
		}
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

	// Vault's 'format' and runner.Command's 'format' are different
	cmdFormat := "json"
	if dbg.LogFormat == "standard" {
		cmdFormat = "string"
	}

	// Create and set the Command
	cmd := runner.Command{
		Command:    cmdStr,
		Format:     cmdFormat,
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
		"vault debug -compress=%s -duration=%s -interval=%s -logformat=%s -metricsinterval=%s -output=%s%s",
		dbg.Compress,
		dbg.Duration,
		dbg.Interval,
		dbg.LogFormat,
		dbg.MetricsInterval,
		dbg.Output,
		filterString,
	)
}
