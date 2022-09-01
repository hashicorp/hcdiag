package run

const (
	dryrunUsageText        = "Displays all runners that would be executed during a normal run without actually executing them."
	serialUsageText        = "Run products in sequence rather than concurrently"
	consulUsageText        = "Run Consul diagnostics"
	nomadUsageText         = "Run Nomad diagnostics"
	terraformEntUsageText  = "Run Terraform Enterprise diagnostics"
	vaultUsageText         = "Run Vault diagnostics"
	autodetectUsageText    = "Auto-Detect installed products; any provided product flags will override this setting"
	includeSinceUsageText  = "Alias for -since, will be overridden if -since is also provided, usage examples: `72h`, `25m`, `45s`, `120h1m90s`"
	sinceUsageText         = "Collect information within this time. Takes a 'go-formatted' duration, usage examples: `72h`, `25m`, `45s`, `120h1m90s`"
	debugDurationUsageText = "How long to run product debug bundle commands. Provide a duration ex: `00h00m00s`. See: -duration in `vault debug`, `consul debug`, and `nomad operator debug`"
	debugIntervalUsageText = "How long metrics collection intervals in product debug commands last. Provide a duration ex: `00h00m00s`. See: -interval in `vault debug`, `consul debug`, and `nomad operator debug`"
	osUsageText            = "Override operating system detection"
	destinationUsageText   = "Path to the directory the bundle should be written in"
	destUsageText          = "Shorthand for -destination"
	configUsageText        = "Path to HCL configuration file"
	includesUsageText      = "Files or directories to include (comma-separated, file-*-globbing available if 'wrapped-*-in-single-quotes')\ne.g. '/var/log/consul-*,/var/log/nomad-*'"
)
