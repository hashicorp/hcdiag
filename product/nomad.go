package product

import (
	"fmt"
	"path/filepath"

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

	seekers, err := NomadSeekers(cfg, api)
	if err != nil {
		return nil, err
	}

	// Apply nomad duration and interval default if CLI is using global defaults
	dur := cfg.DebugDuration
	if dur == 10*time.Second {
		dur = NomadDebugDuration
	}
	interval := cfg.DebugInterval
	if interval == 5*time.Second {
		interval = NomadDebugInterval
	}

	return &Product{
		Seekers:       seekers,
		DebugDuration: dur,
		DebugInterval: interval,
	}, nil
}

// NomadSeekers seek information about Nomad.
func NomadSeekers(cfg Config, api *client.APIClient) ([]*s.Seeker, error) {
	seekers := []*s.Seeker{
		s.NewCommander("nomad version", "string"),
		s.NewCommander("nomad node status -self -json", "json"),
		s.NewCommander("nomad agent-info -json", "json"),
		s.NewCommander(fmt.Sprintf("nomad operator debug -log-level=TRACE -node-id=all -max-nodes=10 -output=%s -duration=%s -interval=%s", cfg.TmpDir, NomadDebugDuration, NomadDebugInterval), "string"),

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
