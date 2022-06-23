package product

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/hcdiag/client"
	s "github.com/hashicorp/hcdiag/op"
	logs "github.com/hashicorp/hcdiag/op/log"
)

const (
	NomadClientCheck   = "nomad version"
	NomadAgentCheck    = "nomad server members"
	NomadDebugDuration = 2 * time.Minute
	NomadDebugInterval = 30 * time.Second
)

// NewNomad takes a product config and creates a Product with all of Nomad's default ops
func NewNomad(logger hclog.Logger, cfg Config) (*Product, error) {
	api, err := client.NewNomadAPI()
	if err != nil {
		return nil, err
	}

	// Apply nomad duration and interval default if CLI is using global defaults.
	// NOTE(mkcp): The downside to this approach is that Nomad cannot be run with a 10s duration and 5s interval.
	//  passing in a zero value from the agent would allow us to do that, but the flags library requires a default value
	//  to be set in order to _show_ that default to the user, so we have to set the agent with that default.
	if DefaultDuration == cfg.DebugDuration {
		cfg.DebugDuration = NomadDebugDuration
	}
	if DefaultInterval == cfg.DebugInterval {
		cfg.DebugInterval = NomadDebugInterval
	}
	ops, err := NomadOps(cfg, api)
	if err != nil {
		return nil, err
	}

	return &Product{
		l:      logger.Named("product"),
		Name:   Nomad,
		Ops:    ops,
		Config: cfg,
	}, nil
}

// NomadOps seek information about Nomad.
func NomadOps(cfg Config, api *client.APIClient) ([]*s.Op, error) {
	ops := []*s.Op{
		s.NewCommander("nomad version", "string"),
		s.NewCommander("nomad node status -self -json", "json"),
		s.NewCommander("nomad agent-info -json", "json"),
		s.NewCommander(fmt.Sprintf("nomad operator debug -log-level=TRACE -node-id=all -max-nodes=10 -output=%s -duration=%s -interval=%s", cfg.TmpDir, cfg.DebugDuration, cfg.DebugInterval), "string"),

		s.NewHTTPer(api, "/v1/agent/members?stale=true"),
		s.NewHTTPer(api, "/v1/operator/autopilot/configuration?stale=true"),
		s.NewHTTPer(api, "/v1/operator/raft/configuration?stale=true"),

		logs.NewDocker("nomad", cfg.TmpDir, cfg.Since),
		logs.NewJournald("nomad", cfg.TmpDir, cfg.Since, cfg.Until),
	}

	// try to detect log location to copy
	if logPath, err := client.GetNomadLogPath(api); err == nil {
		dest := filepath.Join(cfg.TmpDir, "logs", "nomad")
		logCopier := s.NewCopier(logPath, dest, cfg.Since, cfg.Until)
		ops = append([]*s.Op{logCopier}, ops...)
	}

	return ops, nil
}
