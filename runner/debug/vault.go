// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package debug

import (
	"fmt"
	"os"
	"path"
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

func (VaultDebug) ID() string {
	return "VaultDebug"
}

func NewVaultDebug(cfg VaultDebugConfig, tmpDir string, debugDuration time.Duration, debugInterval time.Duration) (*VaultDebug, error) {

	dbg := VaultDebug{
		// No compression because the hcdiag bundle will get compressed anyway
		Compress: "true",
		// Use debug duration and interval
		Duration:        debugDuration.String(),
		Interval:        debugInterval.String(),
		LogFormat:       "standard",
		MetricsInterval: "10s",
		Targets:         cfg.Targets,
		Redactions:      cfg.Redactions,
		output:          tmpDir,
	}

	if cfg.Compress != "" {
		dbg.Compress = cfg.Compress
	}
	if cfg.Duration != "" {
		dbg.Duration = cfg.Duration
	}
	if cfg.Interval != "" {
		dbg.Interval = cfg.Interval
	}
	if cfg.LogFormat != "" {
		dbg.LogFormat = cfg.LogFormat
	}
	if cfg.MetricsInterval != "" {
		dbg.MetricsInterval = cfg.MetricsInterval
	}

	return &dbg, nil
}

func (dbg VaultDebug) Run() op.Op {
	startTime := time.Now()

	// Allow more than one VaultDebug to create output directories during the same run
	dir, err := os.MkdirTemp(dbg.output, "VaultDebug*")
	if err != nil {
		return op.New(dbg.ID(), nil, op.Fail, err, runner.Params(dbg), startTime, time.Now())
	}

	// Assemble the vault debug command to execute
	filterString := filterArgs("target", dbg.Targets)
	cmdStr := vaultCmdString(dbg, filterString, dir)

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

// vaultCmdString takes a VaultDebug, a filterString, and a tmpDir string that's safe to write files into, and creates a valid Vault debug command string
func vaultCmdString(dbg VaultDebug, filterString, tmpDir string) string {
	var fileEnding string

	if dbg.Compress == "true" {
		fileEnding = ".tar.gz"
	}
	dbg.output = path.Join(tmpDir, fmt.Sprintf("VaultDebug%s", fileEnding))

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
