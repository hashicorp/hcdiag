package debug

import (
	"fmt"
	"time"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = ConsulDebug{}

// ConsulDebugConfig is a config struct for ConsulDebug runners
type ConsulDebugConfig struct {
	Archive    string
	Duration   string
	Interval   string
	Captures   []string
	Redactions []*redact.Redact
}

// ConsulDebug represents a ConsulDebug runner
type ConsulDebug struct {
	Archive    string           `json:"archive"`
	Duration   string           `json:"duration"`
	Interval   string           `json:"interval"`
	Captures   []string         `json:"captures"`
	Redactions []*redact.Redact `json:"redactions"`

	output string
}

func (d ConsulDebug) ID() string {
	return "ConsulDebug"
}

func NewConsulDebug(cfg ConsulDebugConfig, tmpDir string, debugDuration time.Duration, debugInterval time.Duration) *ConsulDebug {
	// Create a pseudorandom string of characters to allow >1 ConsulDebug runner without filename collisions
	randStr := randAlphanumString(4)

	dbg := ConsulDebug{
		// No compression because the hcdiag bundle will get compressed anyway
		Archive: "true",
		// Use debug duration and interval
		Duration: debugDuration.String(),
		Interval: debugInterval.String(),
		// Creates a subdirectory inside output dir
		output:     debugOutputPath(tmpDir, "ConsulDebug", randStr),
		Captures:   cfg.Captures,
		Redactions: cfg.Redactions,
	}

	if cfg.Archive != "" {
		dbg.Archive = cfg.Archive
	}

	if len(cfg.Duration) > 0 {
		dbg.Duration = cfg.Duration
	}
	if len(cfg.Interval) > 0 {
		dbg.Interval = cfg.Interval
	}

	return &dbg
}

func (dbg ConsulDebug) Run() op.Op {
	startTime := time.Now()

	filterString := filterArgs("capture", dbg.Captures)

	// Assemble the Consul debug command to execute
	cmdStr := consulCmdString(dbg, filterString)

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

// consulCmdString takes a ConsulDebug and a filterString, and creates a valid Consul debug command string
func consulCmdString(dbg ConsulDebug, filterString string) string {
	return fmt.Sprintf(
		"consul debug -archive=%s -duration=%s -interval=%s -output=%s%s",
		dbg.Archive,
		dbg.Duration,
		dbg.Interval,
		dbg.output,
		filterString,
	)
}
