package product

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/hashicorp/hcdiag/client"
	s "github.com/hashicorp/hcdiag/seeker"
)

const (
	NomadClientCheck   = "nomad version"
	NomadAgentCheck    = "nomad server members"
	NomadDebugDuration = "2m"
	NomadDebugInterval = "30s"
)

// NewNomad takes a product config and creates a Product with all of Nomad's default seekers
func NewNomad(cfg Config) (*Product, error) {
	seekers, err := NomadSeekers(cfg.TmpDir, cfg.Since, cfg.Until)
	if err != nil {
		return nil, err
	}
	return &Product{
		Seekers: seekers,
	}, nil
}

// NomadSeekers seek information about Nomad.
func NomadSeekers(tmpDir string, since, until time.Time) ([]*s.Seeker, error) {
	api, err := client.NewNomadAPI()
	if err != nil {
		return nil, err
	}

	seekers := []*s.Seeker{
		s.NewCommander("nomad version", "string"),
		s.NewCommander("nomad node status -self -json", "json"),
		s.NewCommander("nomad agent-info -json", "json"),
		s.NewCommander(fmt.Sprintf("nomad operator debug -log-level=TRACE -node-id=all -max-nodes=10 -output=%s -duration=%s -interval=%s", tmpDir, NomadDebugDuration, NomadDebugInterval), "string"),

		s.NewHTTPer(api, "/v1/agent/members?stale=true"),
		s.NewHTTPer(api, "/v1/operator/autopilot/configuration?stale=true"),
		s.NewHTTPer(api, "/v1/operator/raft/configuration?stale=true"),
	}

	// try to detect log location to copy
	if logPath, err := client.GetNomadLogPath(api); err == nil {
		dest := filepath.Join(tmpDir, "logs", "nomad")
		logCopier := s.NewCopier(logPath, dest, since, until)
		seekers = append([]*s.Seeker{logCopier}, seekers...)
	}
	// get logs from journald if available
	if journald := s.JournaldGetter("nomad", tmpDir, since, until); journald != nil {
		seekers = append(seekers, journald)
	}

	return seekers, nil
}
