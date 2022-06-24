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
	BoundaryClientCheck = "boundary version"
)

// NewBoundary takes a product config and creates a Product with all of Boundary's default seekers
func NewBoundary(logger hclog.Logger, cfg Config) (*Product, error) {
	api, err := client.NewBoundaryAPI()
	if err != nil {
		return nil, err
	}

	seekers, err := BoundarySeekers(cfg, api)
	if err != nil {
		return nil, err
	}
	return &Product{
		l:       logger.Named("product"),
		Name:    Boundary,
		Seekers: seekers,
		Config:  cfg,
	}, nil
}

// BoundarySeekers seek information about Boundary.
func BoundarySeekers(cfg Config, api *client.APIClient) ([]*s.Seeker, error) {
	seekers := []*s.Seeker{
		s.NewCommander("boundary version", "string"),
		s.NewCommander(fmt.Sprintf("boundary debug -output=%s/BoundaryDebug -duration=%s -interval=%s", cfg.TmpDir, cfg.DebugDuration, cfg.DebugInterval), "string"),
		// https://github.com/hashicorp/boundary/blob/main/internal/gen/controller.swagger.json
		s.NewHTTPer(api, "/v1/accounts"),
		s.NewHTTPer(api, "/v1/auth-methods"),
		s.NewHTTPer(api, "/v1/auth-tokens"),
		s.NewHTTPer(api, "/v1/credential-libraries"),
		s.NewHTTPer(api, "/v1/credential-stores"),
		s.NewHTTPer(api, "/v1/credentials"),
		s.NewHTTPer(api, "/v1/groups"),
		s.NewHTTPer(api, "/v1/host-catalogs"),
		s.NewHTTPer(api, "/v1/host-sets"),
		s.NewHTTPer(api, "/v1/hosts"),
		s.NewHTTPer(api, "/v1/managed-groups"),
		s.NewHTTPer(api, "/v1/roles"),
		s.NewHTTPer(api, "/v1/sessions"),
		s.NewHTTPer(api, "/v1/targets"),
		s.NewHTTPer(api, "/v1/users"),

		logs.NewDocker("boundary", cfg.TmpDir, cfg.Since),
		logs.NewJournald("boundary", cfg.TmpDir, cfg.Since, cfg.Until),
	}

	// try to detect log location to copy
	if logPath, err := client.GetBoundaryLogPath(api); err == nil {
		dest := filepath.Join(cfg.TmpDir, "logs/boundary")
		logCopier := s.NewCopier(logPath, dest, cfg.Since, cfg.Until)
		seekers = append([]*s.Seeker{logCopier}, seekers...)
	}

	return seekers, nil
}
