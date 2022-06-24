package product

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/hcdiag/client"
	s "github.com/hashicorp/hcdiag/seeker"
	logs "github.com/hashicorp/hcdiag/seeker/log"
)

const (
	WaypointClientCheck = "waypoint version"
	WaypointAgentCheck  = "waypoint info"
)

// NewWaypoint takes a product config and creates a Product with all of Waypoint's default seekers
func NewWaypoint(logger hclog.Logger, cfg Config) (*Product, error) {
	api, err := client.NewWaypointAPI()
	if err != nil {
		return nil, err
	}

	seekers, err := WaypointSeekers(cfg, api)
	if err != nil {
		return nil, err
	}
	return &Product{
		l:       logger.Named("product"),
		Name:    Waypoint,
		Seekers: seekers,
		Config:  cfg,
	}, nil
}

// WaypointSeekers seek information about Waypoint.
func WaypointSeekers(cfg Config, api *client.APIClient) ([]*s.Seeker, error) {
	seekers := []*s.Seeker{
		s.NewCommander("waypoint version", "string"),
		s.NewCommander(fmt.Sprintf("waypoint debug -output=%s/WaypointDebug -duration=%s -interval=%s", cfg.TmpDir, cfg.DebugDuration, cfg.DebugInterval), "string"),

		s.NewHTTPer(api, "/v1/agent/self"),
		s.NewHTTPer(api, "/v1/agent/metrics"),
		s.NewHTTPer(api, "/v1/catalog/datacenters"),
		s.NewHTTPer(api, "/v1/catalog/services"),
		s.NewHTTPer(api, "/v1/namespace"),
		s.NewHTTPer(api, "/v1/status/leader"),
		s.NewHTTPer(api, "/v1/status/peers"),

		logs.NewDocker("waypoint", cfg.TmpDir, cfg.Since),
		logs.NewJournald("waypoint", cfg.TmpDir, cfg.Since, cfg.Until),
	}

	// try to detect log location to copy
	if logPath, err := client.GetWaypointLogPath(api); err == nil {
		dest := filepath.Join(cfg.TmpDir, "logs/waypoint")
		logCopier := s.NewCopier(logPath, dest, cfg.Since, cfg.Until)
		seekers = append([]*s.Seeker{logCopier}, seekers...)
	}

	return seekers, nil
}
