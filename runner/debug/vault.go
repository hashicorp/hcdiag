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

// Functional Options pattern
type vaultDebugOption func(*VaultDebug)

func WithCompress(s string) vaultDebugOption {
	return func(d *VaultDebug) {
		d.Compress = s
		if s == "true" {
			d.Output = d.Output + ".tar.gz"
		}
	}
}

func WithDuration(s string) vaultDebugOption {
	return func(d *VaultDebug) {
		d.Duration = s
	}
}

func WithInterval(s string) vaultDebugOption {
	return func(d *VaultDebug) {
		d.Interval = s
	}
}

func WithLogFormat(s string) vaultDebugOption {
	return func(d *VaultDebug) {
		d.LogFormat = s
	}
}

func WithMetricsInterval(s string) vaultDebugOption {
	return func(d *VaultDebug) {
		d.MetricsInterval = s
	}
}

func WithTargets(s []string) vaultDebugOption {
	return func(d *VaultDebug) {
		d.Targets = s
	}
}

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

// NewVaultDebug takes a product config, a slice of redactions, and any number of vaultDebugOptions and returns a valid VaultDebug runner
func NewVaultDebug(cfg product.Config, redactions []*redact.Redact, opts ...vaultDebugOption) *VaultDebug {
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

	// Apply functional options
	for _, opt := range opts {
		opt(&dbg)
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
	if dbg.LogFormat == "standard" {
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
