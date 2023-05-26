// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package product

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/client"
	"github.com/hashicorp/hcdiag/hcl"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/runner"
	"github.com/hashicorp/hcdiag/runner/do"
)

// NewTFE takes a logger and product config, and it creates a Product with all of TFE's default runners.
func NewTFE(logger hclog.Logger, cfg Config) (*Product, error) {
	return NewTFEWithContext(context.Background(), logger, cfg)
}

// NewTFEWithContext takes a context, a logger, and product config, and it creates a Product with all of TFE's default runners.
func NewTFEWithContext(ctx context.Context, logger hclog.Logger, cfg Config) (*Product, error) {
	// Prepend product-specific redactions to agent-level redactions from cfg
	defaultRedactions, err := tfeRedactions()
	if err != nil {
		return nil, err
	}
	cfg.Redactions = redact.Flatten(defaultRedactions, cfg.Redactions)

	product := &Product{
		l:      logger.Named("product"),
		Name:   TFE,
		Config: cfg,
	}
	api, err := client.NewTFEAPI()
	if err != nil {
		return nil, err
	}

	if cfg.HCL != nil {
		// Map product-specific redactions from our config
		hclProductRedactions, err := hcl.MapRedacts(cfg.HCL.Redactions)
		if err != nil {
			return nil, err
		}
		// Prepend product HCL redactions to our product defaults
		cfg.Redactions = redact.Flatten(hclProductRedactions, cfg.Redactions)

		hclRunners, err := hcl.BuildRunnersWithContext(ctx, cfg.HCL, cfg.TmpDir, cfg.DebugDuration, cfg.DebugInterval, api, cfg.Since, cfg.Until, nil)
		if err != nil {
			return nil, err
		}
		product.Runners = append(product.Runners, hclRunners...)
		product.Excludes = cfg.HCL.Excludes
		product.Selects = cfg.HCL.Selects
	}

	// Add built-in runners
	builtInRunners, err := tfeRunners(ctx, cfg, api, logger)
	if err != nil {
		return nil, err
	}
	product.Runners = append(product.Runners, builtInRunners...)

	return product, nil
}

// tfeRunners configures a set of default runners for TFE.
func tfeRunners(ctx context.Context, cfg Config, api *client.APIClient, l hclog.Logger) ([]runner.Runner, error) {
	var r []runner.Runner

	// Set up the Replicated Support Bundle runners
	supportBundleCmd, err := runner.NewCommandWithContext(ctx, runner.CommandConfig{
		Command:    "replicatedctl support-bundle",
		Redactions: cfg.Redactions,
	})
	if err != nil {
		return nil, err
	}
	supportBundleCopy, err := runner.NewCopyWithContext(ctx, runner.CopyConfig{
		Path:       "/var/lib/replicated/support-bundles/replicated-support*.tar.gz",
		DestDir:    cfg.TmpDir,
		Since:      cfg.Since,
		Until:      cfg.Until,
		Redactions: cfg.Redactions,
	})
	if err != nil {
		return nil, err
	}

	// The support bundle that we copy is built by the `replicated support-bundle` command, so we need to ensure
	// that these run in sequence.
	replicatedSeq := do.NewSeq(do.SeqConfig{
		Runners: []runner.Runner{
			supportBundleCmd,
			supportBundleCopy,
		},
		Label:       "support-bundle",
		Description: "replicated support bundle",
		Logger:      l,
	})
	r = append(r, replicatedSeq)

	// Set up HTTP runners
	for _, hc := range []runner.HttpConfig{
		{Client: api, Path: "/api/v2/admin/customization-settings", Redactions: cfg.Redactions},
		{Client: api, Path: "/api/v2/admin/general-settings", Redactions: cfg.Redactions},
		{Client: api, Path: "/api/v2/admin/organizations", Redactions: cfg.Redactions},
		{Client: api, Path: "/api/v2/admin/terraform-versions", Redactions: cfg.Redactions},
		{Client: api, Path: "/api/v2/admin/twilio-settings", Redactions: cfg.Redactions},
		// page size 1 because we only actually care about total workspace count in the `meta` field
		{Client: api, Path: "/api/v2/admin/workspaces?page[size]=1", Redactions: cfg.Redactions},
		{Client: api, Path: "/api/v2/admin/users?page[size]=1", Redactions: cfg.Redactions},
		{Client: api, Path: "/api/v2/admin/runs?page[size]=1", Redactions: cfg.Redactions},
	} {
		c, err := runner.NewHTTPWithContext(ctx, hc)
		if err != nil {
			return nil, err
		}
		r = append(r, c)
	}

	// Set up Command runners
	for _, cc := range []runner.CommandConfig{
		{Command: "docker -v", Redactions: cfg.Redactions},
		{Command: "replicatedctl app status --output json", Format: "json", Redactions: cfg.Redactions},
		{Command: "lsblk --json", Format: "json", Redactions: cfg.Redactions},
		{Command: "replicatedctl app-config view -o json --group capacity", Format: "json", Redactions: cfg.Redactions},
		{Command: "replicatedctl app-config view -o json --group production_type", Format: "json", Redactions: cfg.Redactions},
		{Command: "replicatedctl app-config view -o json --group log_forwarding", Format: "json", Redactions: cfg.Redactions},
		{Command: "replicatedctl app-config view -o json --group blob", Format: "json", Redactions: cfg.Redactions},
		{Command: "replicatedctl app-config view -o json --group worker_image", Format: "json", Redactions: cfg.Redactions},
		{Command: "replicatedctl params export --template '{{.Airgap}}'", Redactions: cfg.Redactions},
		{Command: "replicated --no-tty admin list-nodes", Format: "json", Redactions: cfg.Redactions},
	} {
		c, err := runner.NewCommandWithContext(ctx, cc)
		if err != nil {
			return nil, err
		}
		r = append(r, c)
	}

	// Set up Shell runners
	for _, sc := range []runner.ShellConfig{
		{Command: "getenforce", Redactions: cfg.Redactions},
		{Command: "env | grep -i proxy", Redactions: cfg.Redactions},
	} {
		s, err := runner.NewShellWithContext(ctx, sc)
		if err != nil {
			return nil, err
		}
		r = append(r, s)
	}

	runners := []runner.Runner{
		do.New(l, "tfe", "tfe runners", r),
	}
	return runners, nil
}

// tfeRedactions returns a slice of default redactions for this product
func tfeRedactions() ([]*redact.Redact, error) {
	configs := []redact.Config{
		{
			Matcher: `(postgres://)[^@{]+`,
			Replace: "${1}REDACTED",
		},
		{
			Matcher: `(SECRET0=)[^ ]+`,
			Replace: "${1}REDACTED",
		},
		{
			Matcher: `(SECRET=)[^ ]+`,
			Replace: "${1}REDACTED",
		},
		{
			Matcher: `(\s+")[a-zA-Z0-9]{32}("\s+)`,
			Replace: "${1}REDACTED${2}",
		},
	}
	redactions, err := redact.MapNew(configs)
	if err != nil {
		return nil, err
	}
	return redactions, nil
}
