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

var _ runner.Runner = NomadDebug{}

// NomadDebugConfig is a config struct for NomadDebug runners
type NomadDebugConfig struct {
	Duration      string
	Interval      string
	LogLevel      string
	MaxNodes      int
	NodeClass     string
	NodeID        string
	PprofDuration string
	PprofInterval string
	ServerID      string
	Stale         bool
	Verbose       bool
	EventTopic    []string

	Redactions []*redact.Redact
}

// NomadDebug represents a NomadDebug runner
type NomadDebug struct {
	Duration      string   `json:"duration"`
	Interval      string   `json:"interval"`
	LogLevel      string   `json:"log_level"`
	MaxNodes      int      `json:"max_nodes"`
	NodeClass     string   `json:"node_class"`
	NodeID        string   `json:"node_id"`
	PprofDuration string   `json:"pprof_duration"`
	PprofInterval string   `json:"pprof_interval"`
	ServerID      string   `json:"server_id"`
	Stale         bool     `json:"stale"`
	Verbose       bool     `json:"verbose"`
	EventTopic    []string `json:"event_topic"`

	Redactions []*redact.Redact `json:"redactions"`
	output     string
}

func (NomadDebug) ID() string {
	return "NomadDebug"
}

func NewNomadDebug(cfg NomadDebugConfig, tmpDir string, debugDuration time.Duration, debugInterval time.Duration) (*NomadDebug, error) {
	dbg := NomadDebug{
		// Use debug duration and interval
		Duration:      debugDuration.String(),
		Interval:      debugInterval.String(),
		LogLevel:      "TRACE",
		MaxNodes:      10,
		NodeClass:     "",
		NodeID:        "all",
		PprofDuration: "1s",
		PprofInterval: "250ms",
		ServerID:      "all",
		Stale:         false,
		Verbose:       false,

		// Creates a subdirectory inside output dir
		output:     tmpDir,
		EventTopic: cfg.EventTopic,
		Redactions: cfg.Redactions,
	}

	if cfg.Duration != "" {
		dbg.Duration = cfg.Duration
	}
	if cfg.Interval != "" {
		dbg.Interval = cfg.Interval
	}
	if cfg.LogLevel != "" {
		dbg.LogLevel = cfg.LogLevel
	}
	// disregard negative values
	if cfg.MaxNodes > 0 {
		dbg.MaxNodes = cfg.MaxNodes
	}
	if cfg.NodeClass != "" {
		dbg.NodeClass = cfg.NodeClass
	}
	if cfg.NodeID != "" {
		dbg.NodeID = cfg.NodeID
	}
	if cfg.PprofDuration != "" {
		dbg.PprofDuration = cfg.PprofDuration
	}
	if cfg.PprofInterval != "" {
		dbg.PprofInterval = cfg.PprofInterval
	}
	if cfg.ServerID != "" {
		dbg.ServerID = cfg.ServerID
	}
	// Bool zero-values from NomadDebugConfig already match defaults
	dbg.Stale = cfg.Stale
	dbg.Verbose = cfg.Verbose

	return &dbg, nil
}

func (dbg NomadDebug) Run() op.Op {
	startTime := time.Now()

	// Allow more than one NomadDebug to create output directories during the same run
	dir, err := os.MkdirTemp(dbg.output, "NomadDebug*")
	if err != nil {
		return op.New(dbg.ID(), nil, op.Fail, err, runner.Params(dbg), startTime, time.Now())
	}

	// Assemble the Nomad debug command to execute
	filterString := filterArgs("event-topic", dbg.EventTopic)
	cmdStr := nomadCmdString(dbg, filterString, dir)

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

// nomadCmdString takes a NomadDebug and a filterString, and creates a valid nomad debug command string
func nomadCmdString(dbg NomadDebug, filterString, tmpDir string) string {
	dbg.output = path.Join(tmpDir, "NomadDebug")

	// elide this option entirely if it's not set via config, because the docs aren't 100% clear about the default value
	var nodeClassOpt string
	if dbg.NodeClass != "" {
		nodeClassOpt = fmt.Sprintf(" -node-class=%s", dbg.NodeClass)
	}

	return fmt.Sprintf(
		"nomad debug -no-color -duration=%s -interval=%s -log-level=%s -max-nodes=%d%s -node-id=%s -pprof-duration=%s -pprof-interval=%s -server-id=%s -stale=%t -verbose=%t -output=%s%s",
		dbg.Duration,
		dbg.Interval,
		dbg.LogLevel,
		dbg.MaxNodes,

		nodeClassOpt,

		dbg.NodeID,
		dbg.PprofDuration,
		dbg.PprofInterval,
		dbg.ServerID,
		dbg.Stale,
		dbg.Verbose,

		dbg.output,
		filterString,
	)
}
