# hcdiag
`hcdiag` simplifies debugging HashiCorp products by automating shared and product-specific diagnostics data collection on individual nodes. Running the binary issues a set of operations that read the current state of the system then write the results to a tar.gz bundle.

The utility is optimized for transparency and frugality. We believe users should be fully informed on how `hcdiag` works and what it collects and that `hcdiag` collects no more data than is necessary.

Features like `-dryrun`, declarative HCL config, and filters give users visibility and agency into what will run, and 
the open, non-proprietary, bundle format makes the results, manifest, and included files available for inspection by users.

We are constantly refining the utility to be safe, robust, and speedy. If you have any concerns please voice them via the GitHub issues so we may address them. 

---

**The documentation in this README corresponds to the main branch of hcdiag. It may contain references to new features that the most recently released version doesn't include.**

**Please see the [Git tag](https://github.com/hashicorp/hcdiag/releases) that corresponds to your version of hcdiag for the proper documentation.**

---

## Table of Contents

- [Installation](docs/installation.md)
- [Usage](#usage)
- [Prerequisites](#prerequisites)
- [Examples](#example-runs)
- [Flags](#flags)
- [Custom Configuration](docs/custom-config.md)
- [Runner Types](docs/runner-types.md)
- [Redactions](./docs/redactions.md)
- [FAQ](./docs/faq.md)
- [Contributing](./CONTRIBUTING.md)
- [Learn Tutorials](#learn-tutorials)

## Usage
To reliably debug the widest variety of issues with the lowest impact on each machine, `hcdiag` runs on one node at a time and gathers the view of the current node whenever possible. 

### Prerequisites
The `hcdiag` binary often issues commands using HashiCorp's product clients. Therefore, the utility must have access to a fully configured client in its environment for product diagnostics. Specifics are offered below per client.

#### Consul
- [Consul CLI documentation](https://www.consul.io/commands/index)
- [Consul API documentation](https://www.consul.io/api-docs)
- **Requirements**
  - The [Consul binary](https://www.consul.io/downloads) must be available on the local machine
  - Environment variable [CONSUL_HTTP_ADDR](https://www.consul.io/commands#consul_http_addr) must be set to the HTTP address to the local Consul agent
  - Environment variable [CONSUL_TOKEN](https://www.consul.io/commands#consul_http_token) must be set to the API access token if ACLs are enabled.

#### Nomad
- [Nomad CLI documentation](https://www.nomadproject.io/docs/commands)
- [Nomad API documentation](https://www.nomadproject.io/api-docs)
- **Requirements**
  - The [Nomad binary](https://www.nomadproject.io/downloads) must be available on the local machine
  - Environment variable [NOMAD_ADDR](https://www.nomadproject.io/docs/commands#nomad_addr) must be set to the HTTP address of the Nomad server
  - Environment variable [NOMAD_TOKEN](https://www.nomadproject.io/docs/commands#nomad_token) may be set to the SecretID of an ACL token for API requests if ACLs are enabled.

#### Terraform Enterprise/Cloud
Terraform Enterprise historically uses replicated to provide similar functionality to `hcdiag` and for ease of use and compatibility during the migration period we include its results in the hcdiag bundle.
- [Terraform Enterprise/Cloud API documentation](https://www.terraform.io/docs/cloud/api/index.html)
- [Replicated CLI documentation](https://help.replicated.com/api/replicatedctl/)
- **Requirements**
  - Environment variable `TFE_HTTP_ADDR` must be set to the HTTP address of a Terraform Enterprise or Terraform Cloud environment
  - Environment variable [TFE_TOKEN](https://www.terraform.io/docs/cloud/api/index.html#authentication) must be set to an appropriate bearer token (user, team, or organization)
  - Unlike other products, the Terraform binary is not required on the target machine
  - CLI currently limited to self-managed TFE environments

#### Vault
- [Vault CLI documentation](https://www.vaultproject.io/docs/commands)
- [Vault API documentation](https://www.vaultproject.io/api)
- **Requirements**
  - [Vault binary](https://www.vaultproject.io/downloads) must be available on the local machine
  - Environment variable [VAULT_ADDR](https://www.vaultproject.io/docs/commands#vault_addr) must be set to the HTTP address of the Vault server
  - Environment variable [VAULT_TOKEN](https://www.vaultproject.io/docs/commands#vault_token) must be set to the Vault authentication token
  - Alternatively, a token may also exist at ~/.vault-token
    - If both are present, `VAULT_TOKEN` will be used.

### Example Runs
- Gather host/node and product diagnostics for all supported HashiCorp products installed on the system
  - `hcdiag`

- Log hcdiag run to console without reading or writing files. (Also checks client requirements setup)
  - `hcdiag -dryrun`

- Gather node and diagnostics for one or many products
  - `hcdiag -vault {-nomad, -consul}`

- Gather diagnostics with config
  - `hcdiag -vault -config cfg.hcl`

- Gather diagnostics from the last day, rather than the default 3 days
  - `hcdiag -vault -since 24hr`

- Gather diagnostics and write bundle to a specific location. (default is `$PWD`)
  - `hcdiag -vault -dest /tmp/hcdiag`

- Gather diagnostics and use the CLI to copy individual files or whole directories
  - `hcdiag -vault -includes "/var/log/dmesg,/var/log/vault-"`

- Gather only host diagnostics (prior to `0.4.0`, this was the behavior of running `hcdiag` with no flags).
  - `hcdiag -autodetect=false`
  - *Note:* The `=` is required here because it is a boolean flag.

### Flags
| Argument        | Description                                                                                                                                                         | Type   | Default Value |
|-----------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------|---------------|
| `dryrun`        | Perform a dry run to display commands without executing them                                                                                                        | bool   | false         |
| `os`            | Override operating system detection                                                                                                                                 | string | "auto"        |
| `consul`        | Run Consul diagnostics                                                                                                                                              | bool   | false         |
| `nomad`         | Run Nomad diagnostics                                                                                                                                               | bool   | false         |
| `terraform-ent` | Run Terraform Enterprise/Cloud diagnostics                                                                                                                          | bool   | false         |
| `vault`         | Run Vault diagnostics                                                                                                                                               | bool   | false         |
| `autodetect`    | Automatically detect which product CLIs are installed and gather diagnostics for each. If any product flags are provided, they override this one.                   | bool   | true          |
| `since`         | Collect information within this time. Takes a 'go-formatted' duration, usage examples: `72h`, `25m`, `45s`, `120h1m90s`                                             | string | "72h"         |
| `include-since` | Alias for -since, will be overridden if -since is also provided, usage examples: `72h`, `25m`, `45s`, `120h1m90s`                                                   | string | "72h"         |
| `includes`      | (DEPRECATED) Files or directories to include (comma-separated, file-*-globbing available if 'wrapped-*-in-single-quotes') e.g. '/var/log/consul-*,/var/log/nomad-*' | string | ""            |
| `destination`   | Path to the directory the bundle should be written in                                                                                                               | string | "."           |
| `dest`          | Shorthand for -destination                                                                                                                                          | string | "."           |
| `config`        | Path to HCL configuration file                                                                                                                                      | string | ""            |

### Installation

You can see detailed installation instructions in each of our [Learn Tutorials](#learn-tutorials).

hcdiag is available on our [releases page](https://releases.hashicorp.com/hcdiag/), as well as via many popular package managers on different operating systems. Please see [docs/Installation.md](docs/installation.md) for details.

### Adding and Filtering Runners with Configuration

To support a variety of troubleshooting use-cases, the runners that hcdiag executes on a system can be customized via a configuration file.

In addition to the defaults hcdiag offers, for the host and products, diagnostic runs can be tailored to specific
use-cases. With the `-config <FILE>` flag, users can execute HCL configuration saved to disk. Here's a simple example:

```
product "consul" {
  command {
    run = "consul version"
    format = "string"
  }
}
```

Executing `hcdiag -config example.hcl` with the HCL above means we add a Runner: the CLI command `consul version`. The
`format = "string"` attribute tells `hcdiag` how to parse the result. The `product "consul" {}` block ensures we configure
the HTTP client for TLS and store the results in the proper location behind the scenes.

For more in-depth examples, check out the [custom configuration documentation](docs/custom-config.md)

**Note** hcdiag is an execution tool, and custom runners allow you to execute arbitrary commands on a system. Please ensure that data privacy is taken into account in all situations, particularly when using custom configuration.

## Redactions

Beginning with version `0.4.0`, `hcdiag` supports redactions. Redactions enable users to tell `hcdiag` about patterns of
text that should be omitted from the results bundle. Redaction matching patterns are based around Regular Expressions.

As of version `0.4.0`, hcdiag includes a default set of redactions which omit email addresses from runner results across all products, as well as more specific redactions for Terraform Enterprise only. Users may also specify custom redactions within configuration files.

See the [Redactions section](./docs/redactions.md) of the custom configuration documentation for more information and examples.


## FAQs

Depending on the context you're using hcdiag in, you may have some questions about the tool. Please scan through our [FAQ documentation](./docs/faq.md) before opening an issue to make sure there's not already an answer, clarifying context, or workaround for your situation.

## Learn Tutorials

We've created detailed tutorials that teach the basics of using hcdiag to gather troubleshooting data for Hashicorp tools.

- [Vault](https://learn.hashicorp.com/tutorials/vault/hcdiag-with-vault)
- [Consul](https://learn.hashicorp.com/tutorials/consul/hcdiag-with-consul)
- [Terraform Enterprise](https://learn.hashicorp.com/tutorials/terraform/hcdiag-with-tfe)
- [Nomad](https://learn.hashicorp.com/tutorials/nomad/hcdiag-with-nomad)

## Contributing

If you have ideas for improvements, if you’ve found bugs in the tool, or if you would like to add
to the codebase, we welcome [your contributions](./CONTRIBUTING.md)!
