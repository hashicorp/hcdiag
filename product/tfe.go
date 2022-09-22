package product

import (
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/client"
	"github.com/hashicorp/hcdiag/hcl"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/runner"
)

// NewTFE takes a logger and product config, and it creates a Product with all of TFE's default runners.
func NewTFE(logger hclog.Logger, cfg Config) (*Product, error) {
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

		hclRunners, err := hcl.BuildRunners(cfg.HCL, cfg.TmpDir, api, cfg.Since, cfg.Until, nil)
		if err != nil {
			return nil, err
		}
		product.Runners = append(product.Runners, hclRunners...)
		product.Excludes = cfg.HCL.Excludes
		product.Selects = cfg.HCL.Selects
	}

	// Add built-in runners
	builtInRunners, err := tfeRunners(cfg, api)
	if err != nil {
		return nil, err
	}
	product.Runners = append(product.Runners, builtInRunners...)

	return product, nil
}

// tfeRunners configures a set of default runners for TFE.
func tfeRunners(cfg Config, api *client.APIClient) ([]runner.Runner, error) {
	return []runner.Runner{
		runner.NewCommander("replicatedctl support-bundle", "string", cfg.Redactions),

		runner.NewCopier("/var/lib/replicated/support-bundles/replicated-support*.tar.gz", cfg.TmpDir, cfg.Since, cfg.Until, cfg.Redactions),

		runner.NewHTTPer(api, "/api/v2/admin/customization-settings", cfg.Redactions),
		runner.NewHTTPer(api, "/api/v2/admin/general-settings", cfg.Redactions),
		runner.NewHTTPer(api, "/api/v2/admin/organizations", cfg.Redactions),
		runner.NewHTTPer(api, "/api/v2/admin/terraform-versions", cfg.Redactions),
		runner.NewHTTPer(api, "/api/v2/admin/twilio-settings", cfg.Redactions),
		// page size 1 because we only actually care about total workspace count in the `meta` field
		runner.NewHTTPer(api, "/api/v2/admin/workspaces?page[size]=1", cfg.Redactions),
		runner.NewHTTPer(api, "/api/v2/admin/users?page[size]=1", cfg.Redactions),
		runner.NewHTTPer(api, "/api/v2/admin/runs?page[size]=1", cfg.Redactions),

		runner.NewCommander("docker -v", "string", cfg.Redactions),
		runner.NewCommander("replicatedctl app status --output json", "json", cfg.Redactions),
		runner.NewCommander("lsblk --json", "json", cfg.Redactions),
		runner.NewCommander("replicatedctl app-config view -o json --group capacity", "json", cfg.Redactions),
		runner.NewCommander("replicatedctl app-config view -o json --group production_type", "json", cfg.Redactions),
		runner.NewCommander("replicatedctl app-config view -o json --group log_forwarding", "json", cfg.Redactions),
		runner.NewCommander("replicatedctl app-config view -o json --group blob", "json", cfg.Redactions),
		runner.NewCommander("replicatedctl app-config view -o json --group worker_image", "json", cfg.Redactions),
		runner.NewCommander("replicatedctl params export --template '{{.Airgap}}'", "string", cfg.Redactions),

		runner.NewSheller("getenforce", cfg.Redactions),
		runner.NewSheller("env | grep -i proxy", cfg.Redactions),

		//
		// NOTE(dcohen): Replicated bundle building, by the numbers
		//

		// TODO(dcohen) bundle/app/container-logs

		// bundle/app/logs
		// NOTE(dcohen temp:) ignoring since and until, since this is a simple file copy
		runner.NewCopier("/var/lib/docker/containers/*/*-json.log", "app/logs", time.Time{}, time.Now(), cfg.Redactions),

		// bundle/app/containers
		// Run a docker inspect for all containers on the host
		runner.NewCommander("docker ps -aq | xargs docker inspect", "json", cfg.Redactions),

		// bundle/default/commands
		runner.NewCommander("date", "string", cfg.Redactions),
		runner.NewCommander("df", "string", cfg.Redactions),
		runner.NewCommander("df -ali", "string", cfg.Redactions),
		runner.NewCommander("dmesg", "string", cfg.Redactions),
		runner.NewCommander("free", "string", cfg.Redactions),
		runner.NewCommander("hostname", "string", cfg.Redactions),
		runner.NewCommander("ip -o addr show", "string", cfg.Redactions),
		runner.NewCommander("ip -o link show", "string", cfg.Redactions),
		runner.NewCommander("ip -o route show", "string", cfg.Redactions),
		// Note(dcohen): 'loadavg' is not always installed, and always available from `uptime`
		runner.NewCommander("uptime", "string", cfg.Redactions),
		runner.NewCommander("ps fauxwww", "string", cfg.Redactions),

		// bundle/default/docker
		// container_ls.json
		runner.NewCommander("docker container ls --format '{{json .}}'", "json", cfg.Redactions),
		// docker_info.json
		runner.NewCommander("docker info --format '{{json .}}'", "json", cfg.Redactions),
		// docker_version.json
		runner.NewCommander("docker version --format '{{json .}}'", "json", cfg.Redactions),
		// image_ls.json
		runner.NewCommander("docker image ls --format '{{json .}}'", "json", cfg.Redactions),

		// bundle/default/etc
		runner.NewCopier("/etc/firewalld", "default/etc", time.Time{}, time.Now(), cfg.Redactions),
		runner.NewCopier("/etc/dnsmasq.conf", "default/etc", time.Time{}, time.Now(), cfg.Redactions),
		runner.NewCopier("/etc/sysconfig/iptables-config", "default/etc", time.Time{}, time.Now(), cfg.Redactions),
		runner.NewCopier("/etc/fstab", "default/etc", time.Time{}, time.Now(), cfg.Redactions),
		runner.NewCopier("/etc/hostname", "default/etc", time.Time{}, time.Now(), cfg.Redactions),
		runner.NewCopier("/etc/hosts", "default/etc", time.Time{}, time.Now(), cfg.Redactions),
		runner.NewCopier("/etc/os-release", "default/etc", time.Time{}, time.Now(), cfg.Redactions),
		// NOTE(dcohen): added /etc/issue in case os-release isn't present
		runner.NewCopier("/etc/issue", "default/etc", time.Time{}, time.Now(), cfg.Redactions),
		runner.NewCopier("/etc/resolv.conf", "default/etc", time.Time{}, time.Now(), cfg.Redactions),

		// TODO(dcohen): journald: (raw docker logs from journald)

		// TODO(dcohen): kubernetes kubelet/logs/logs.raw

		// bundle/default/proc
		runner.NewCommander("cat /proc/cpuinfo", "string", cfg.Redactions),
		runner.NewCommander("cat /proc/loadavg", "string", cfg.Redactions),
		runner.NewCommander("cat /proc/meminfo", "string", cfg.Redactions),
		runner.NewCommander("cat /proc/mounts", "string", cfg.Redactions),
		runner.NewCommander("cat /proc/uptime", "string", cfg.Redactions),
		runner.NewCommander("cat /proc/version", "string", cfg.Redactions),
		runner.NewCommander("cat /proc/vmstat", "string", cfg.Redactions),

		// NOTE(dcohen): omitted: bundle/default/{replicated,support-bundle}

		// NOTE(dcohen): omitted: bundle/docker/* (dupes)
		// NOTE(dcohen): omitted: bundle/kubernetes/logs (dupes)

		// TODO(dcohen): bundle/kubernetes/inspect

		// NOTE(dcohen): omitted: bundle/os/* (dupes)
		// NOTE(dcohen): omitted: bundle/replicated/*
		// NOTE(dcohen): omitted: bundle/retraced/*

	}, nil
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
