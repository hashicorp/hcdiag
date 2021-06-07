package products

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/hashicorp/host-diagnostics/apiclients"
	s "github.com/hashicorp/host-diagnostics/seeker"
)

const (
	NomadClientCheck = "nomad version"
	NomadAgentCheck  = "nomad server members"
)

// FIXME(mkcp): doccomment
func NewNomad(cfg Config) *Product {
	return &Product{
		Seekers:     NomadSeekers(cfg.TmpDir, cfg.From, cfg.To),
	}
}

// NomadSeekers seek information about Nomad.
func NomadSeekers(tmpDir string, from, to time.Time) []*s.Seeker {
	api := apiclients.NewNomadAPI()

	seekers := []*s.Seeker{
		s.NewCommander("nomad version", "string"),
		s.NewCommander("nomad node status -json", "json"),
		s.NewCommander("nomad agent-info -json", "json"),
		s.NewCommander(fmt.Sprintf("nomad operator debug -output=%s -duration=%ds", tmpDir, DefaultDebugSeconds), "string"),
		// s.NewCommander("nomad operator metrics", "json", false), // TODO: uncomment (it's very verbose, so not noisy during testing)

		s.NewHTTPer(api, "/v1/agent/members"),
		s.NewHTTPer(api, "/v1/plugins?type=csi"),
		s.NewHTTPer(api, "/v1/operator/autopilot/configuration"),
		s.NewHTTPer(api, "/v1/operator/raft/configuration"),
	}

	// try to detect log location to copy
	if logPath, err := apiclients.GetNomadLogPath(api); err == nil {
		dest := filepath.Join(tmpDir, "logs/nomad")
		logCopier := s.NewCopier(logPath, dest, from, to)
		seekers = append([]*s.Seeker{logCopier}, seekers...)
	}

	return seekers
}
