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

func (ConsulDebug) ID() string {
	return "ConsulDebug"
}

func NewConsulDebug(cfg ConsulDebugConfig, tmpDir string, debugDuration time.Duration, debugInterval time.Duration) (*ConsulDebug, error) {
	dbg := ConsulDebug{
		// No compression because the hcdiag bundle will get compressed anyway
		Archive: "true",
		// Use debug duration and interval
		Duration: debugDuration.String(),
		Interval: debugInterval.String(),
		// Creates a subdirectory inside output dir
		output:     tmpDir,
		Captures:   cfg.Captures,
		Redactions: cfg.Redactions,
	}

	if cfg.Archive != "" {
		dbg.Archive = cfg.Archive
	}

	if cfg.Duration != "" {
		dbg.Duration = cfg.Duration
	}
	if cfg.Interval != "" {
		dbg.Interval = cfg.Interval
	}

	return &dbg, nil
}

func (dbg ConsulDebug) Run() op.Op {
	startTime := time.Now()

	// Allow more than one ConsulDebug to create output directories during the same run
	dir, err := os.MkdirTemp(dbg.output, "ConsulDebug*")
	if err != nil {
		return op.New(dbg.ID(), nil, op.Fail, err, runner.Params(dbg), startTime, time.Now())
	}

	// Assemble the Consul debug command to execute
	filterString := filterArgs("capture", dbg.Captures)
	cmdStr := consulCmdString(dbg, filterString, dir)

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
func consulCmdString(dbg ConsulDebug, filterString, tmpDir string) string {
	dbg.output = path.Join(tmpDir, "ConsulDebug")

	return fmt.Sprintf(
		"consul debug -archive=%s -duration=%s -interval=%s -output=%s%s",
		dbg.Archive,
		dbg.Duration,
		dbg.Interval,
		dbg.output,
		filterString,
	)
}
