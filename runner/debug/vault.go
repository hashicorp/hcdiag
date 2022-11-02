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

// VaultDebugConfig wraps all command options except 'output' (which is always the hcdiag directory): https://developer.hashicorp.com/vault/docs/commands/debug#command-options
// type VaultDebugConfig struct {
// 	// No compression because the hcdiag bundle will get compressed anyway
// 	Compress string
// 	Duration string
// 	Interval string
// 	LogFormat string
// 	MetricsInterval string
// 	Output string
// 	Targets []string
// 	Command runner.Command
// 	Redactions []*redact.Redact
// }

// func newVaultDebugConfig() *VaultDebugConfig {
// 	return &VaultDebugConfig{
// 		// No compression because the hcdiag bundle will get compressed anyway
// 		Compress:        "false",
// 		Duration:        "2m",
// 		Interval:        "30s",
// 		LogFormat:       "standard",
// 		MetricsInterval: "10s",
// 		Output:          fmt.Sprintf("%s/VaultDebug", cfg.TmpDir),
// 		Targets:         []string{},
// 		Command:         runner.Command{},
// 		Redactions:      []*redact.Redact{},
// 	}
// }

type VaultDebug struct {
	Compress        string           `json:"compress"`
	Duration        string           `json:"duration"`
	Interval        string           `json:"interval"`
	LogFormat       string           `json:"logformat"`
	MetricsInterval string           `json:"metricsinterval"`
	Output          string           `json:"output"`
	Targets         []string         `json:"targets"`
	Command         runner.Command   `json:"command"`
	Redactions      []*redact.Redact `json:"redactions"`
}

func (d VaultDebug) ID() string {
	return "VaultDebug"
}

// NewVaultDebug TODO TODO
func NewVaultDebug(cfg product.Config, compress string, duration string, interval string, logformat string, metricsinterval string, targets []string, redactions []*redact.Redact) *VaultDebug {
	dbg := VaultDebug{
		// No compression because the hcdiag bundle will get compressed anyway
		Compress:        "false",
		Duration:        "2m",
		Interval:        "30s",
		LogFormat:       "standard",
		MetricsInterval: "10s",
		Output:          fmt.Sprintf("%s/VaultDebug", cfg.TmpDir),
		Targets:         []string{},
		Command:         runner.Command{},
		Redactions:      redactions,
	}

	if len(compress) > 0 {
		dbg.Compress = compress
		if dbg.Compress == "true" {
			dbg.Output = dbg.Output + ".tar.gz"
		}
	}
	if len(duration) > 0 {
		dbg.Duration = duration
	}
	if len(interval) > 0 {
		dbg.Interval = interval
	}
	if len(logformat) > 0 {
		dbg.LogFormat = logformat
	}
	if len(metricsinterval) > 0 {
		dbg.MetricsInterval = metricsinterval
	}

	filterString, err := productFilterString(cfg.Name, dbg.Targets)
	if err != nil {
		// TODO figure out error handling inside of a runner constructor -- no other runners need this
		panic(err)
	}

	// Create the commandstring
	var cmdStr = fmt.Sprintf(
		"vault debug -compress=%s -duration=%s -interval=%s -logformat=%s -metricsinterval=%s -output=%s%s",
		dbg.Compress,
		dbg.Duration,
		dbg.Interval,
		dbg.LogFormat,
		dbg.MetricsInterval,
		dbg.Output,
		filterString,
	)
	// Vault's 'format' and runner.Command's 'format' are different
	cmdFormat := "json"
	if logformat == "standard" {
		cmdFormat = "string"
	}

	// Create and set the Command
	dbg.Command = runner.Command{
		Command:    cmdStr,
		Format:     cmdFormat,
		Redactions: redactions,
	}

	return &dbg
}

func (d VaultDebug) Run() op.Op {
	startTime := time.Now()

	o := d.Command.Run()
	if o.Error != nil {
		return op.New(d.ID(), o.Result, op.Fail, o.Error, runner.Params(d), startTime, time.Now())
	}

	return op.New(d.ID(), o.Result, op.Success, nil, runner.Params(d), startTime, time.Now())
}
