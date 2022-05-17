package product

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/hashicorp/hcdiag/client"
	s "github.com/hashicorp/hcdiag/seeker"
	logs "github.com/hashicorp/hcdiag/seeker/log"
)

const (
	NomadClientCheck   = "nomad version"
	NomadAgentCheck    = "nomad server members"
	NomadDebugDuration = 2 * time.Minute
	NomadDebugInterval = 30 * time.Second
)

// NewNomad takes a product config and creates a Product with all of Nomad's default seekers
func NewNomad(cfg Config) (*Product, error) {
	api, err := client.NewNomadAPI()
	if err != nil {
		return nil, err
	}

	// Apply nomad duration and interval default if CLI is using global defaults
	// NOTE(mkcp): This isn't ideal because we're using magic numbers here to match the CLI defaults. We could pass a
	//  tuple or something in from the CLI that describes both the default and the user-set value... but i'm timeboxing this.
	if cfg.DebugDuration == 10*time.Second {
		cfg.DebugDuration = NomadDebugDuration
	}
	if cfg.DebugInterval == 5*time.Second {
		cfg.DebugInterval = NomadDebugInterval
	}
	seekers, err := NomadSeekers(cfg, api)
	if err != nil {
		return nil, err
	}

	return &Product{
		Seekers: seekers,
	}, nil
}

// NomadSeekers seek information about Nomad.
func NomadSeekers(cfg Config, api *client.APIClient) ([]*s.Seeker, error) {
	seekers := []*s.Seeker{
		s.NewCommander("nomad version", "string"),
		s.NewCommander("nomad node status -self -json", "json"),
		s.NewCommander("nomad agent-info -json", "json"),
		s.NewCommander(fmt.Sprintf("nomad operator debug -log-level=TRACE -node-id=all -max-nodes=10 -output=%s -duration=%s -interval=%s", cfg.TmpDir, cfg.DebugDuration, cfg.DebugInterval), "string"),

		s.NewHTTPer(api, "/v1/agent/members?stale=true"),
		s.NewHTTPer(api, "/v1/operator/autopilot/configuration?stale=true"),
		s.NewHTTPer(api, "/v1/operator/raft/configuration?stale=true"),

		logs.NewDocker("nomad", cfg.TmpDir, cfg.Since),
	}

	// try to detect log location to copy
	if logPath, err := client.GetNomadLogPath(api); err == nil {
		dest := filepath.Join(cfg.TmpDir, "logs", "nomad")
		logCopier := s.NewCopier(logPath, dest, cfg.Since, cfg.Until)
		seekers = append([]*s.Seeker{logCopier}, seekers...)
	}
	// get logs from journald if available
	if journald := s.JournaldGetter("nomad", cfg.TmpDir, cfg.Since, cfg.Until); journald != nil {
		seekers = append(seekers, journald)
	}

	return seekers, nil
}
