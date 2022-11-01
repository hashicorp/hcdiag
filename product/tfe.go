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

		hclRunners, err := hcl.BuildRunnersWithContext(ctx, cfg.HCL, cfg.TmpDir, api, cfg.Since, cfg.Until, nil)
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
	r := []runner.Runner{
		do.NewSync(l, "support-bundle", "replicated support bundle",
			// The support bundle that we copy is built by the `replicated support-bundle` command, so we need to ensure
			// that these run serially.
			[]runner.Runner{
				runner.NewCommandWithContext(ctx, "replicatedctl support-bundle", "string", cfg.Redactions),
				runner.NewCopyWithContext(ctx, "/var/lib/replicated/support-bundles/replicated-support*.tar.gz", cfg.TmpDir, cfg.Since, cfg.Until, cfg.Redactions),
			}),

		runner.NewHTTP(api, "/api/v2/admin/customization-settings", cfg.Redactions),
		runner.NewHTTP(api, "/api/v2/admin/general-settings", cfg.Redactions),
		runner.NewHTTP(api, "/api/v2/admin/organizations", cfg.Redactions),
		runner.NewHTTP(api, "/api/v2/admin/terraform-versions", cfg.Redactions),
		runner.NewHTTP(api, "/api/v2/admin/twilio-settings", cfg.Redactions),
		// page size 1 because we only actually care about total workspace count in the `meta` field
		runner.NewHTTP(api, "/api/v2/admin/workspaces?page[size]=1", cfg.Redactions),
		runner.NewHTTP(api, "/api/v2/admin/users?page[size]=1", cfg.Redactions),
		runner.NewHTTP(api, "/api/v2/admin/runs?page[size]=1", cfg.Redactions),

		runner.NewCommandWithContext(ctx, "docker -v", "string", cfg.Redactions),
		runner.NewCommandWithContext(ctx, "replicatedctl app status --output json", "json", cfg.Redactions),
		runner.NewCommandWithContext(ctx, "lsblk --json", "json", cfg.Redactions),
		runner.NewCommandWithContext(ctx, "replicatedctl app-config view -o json --group capacity", "json", cfg.Redactions),
		runner.NewCommandWithContext(ctx, "replicatedctl app-config view -o json --group production_type", "json", cfg.Redactions),
		runner.NewCommandWithContext(ctx, "replicatedctl app-config view -o json --group log_forwarding", "json", cfg.Redactions),
		runner.NewCommandWithContext(ctx, "replicatedctl app-config view -o json --group blob", "json", cfg.Redactions),
		runner.NewCommandWithContext(ctx, "replicatedctl app-config view -o json --group worker_image", "json", cfg.Redactions),
		runner.NewCommandWithContext(ctx, "replicatedctl params export --template '{{.Airgap}}'", "string", cfg.Redactions),
		runner.NewCommandWithContext(ctx, "replicated --no-tty admin list-nodes", "json", cfg.Redactions),

		runner.NewShell("getenforce", cfg.Redactions),
		runner.NewShell("env | grep -i proxy", cfg.Redactions),
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
