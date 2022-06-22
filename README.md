# hcdiag
`hcdiag` simplifies debugging HashiCorp products by automating shared and product-specific diagnostics data collection on individual nodes. Running the binary issues a set of operations that read the current state of the system then write the results to a tar.gz bundle.

The utility is optimized for transparency and frugality. We believe users should be fully informed on how `hcdiag` works and what it collects and that `hcdiag` collects no more data than is necessary.

Features like `-dryrun`, declarative HCL config, and filters give users visibility and agency into what will run, and 
the open, non-proprietary, bundle format makes the results, manifest, and included files available for inspection by users.

We are constantly refining the utility to be safe, robust, and speedy. If you have any concerns please voice them via the GitHub issues so we may address them. 

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
- Log hcdiag run to console without reading or writing. (Also checks client requirements setup)
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
  - `hcdiag -vault -include "/var/log/dmesg,/var/log/vault-"`

### Flags
| Argument        | Description                                                                                                                                            | Type   | Default Value |
|-----------------|--------------------------------------------------------------------------------------------------------------------------------------------------------|--------|---------------|
| `dryrun`        | Perform a dry run to display commands without executing them                                                                                           | bool   | false         |
| `os`            | Override operating system detection                                                                                                                    | string | "auto"        |
| `consul`        | Run Consul diagnostics                                                                                                                                 | bool   | false         |
| `nomad`         | Run Nomad diagnostics                                                                                                                                  | bool   | false         |
| `terraform-ent` | Run Terraform Enterprise/Cloud diagnostics                                                                                                             | bool   | false         |
| `vault`         | Run Vault diagnostics                                                                                                                                  | bool   | false         |
| `all`           | DEPRECATED: Run all available product diagnostics                                                                                                      | bool   | false         |
| `since`         | Collect information within this time. Takes a 'go-formatted' duration, usage examples: `72h`, `25m`, `45s`, `120h1m90s`                                | string | "72h"         |
| `include-since` | Alias for -since, will be overridden if -since is also provided, usage examples: `72h`, `25m`, `45s`, `120h1m90s`                                      | string | "72h"         |
| `includes`      | files or directories to include (comma-separated, file-*-globbing available if 'wrapped-*-in-single-quotes') e.g. '/var/log/consul-*,/var/log/nomad-*' | string | ""            |
| `destination`   | Path to the directory the bundle should be written in                                                                                                  | string | "."           |
| `dest`          | Shorthand for -destination                                                                                                                             | string | "."           |
| `config`        | Path to HCL configuration file                                                                                                                         | string | ""            |
| `serial`        | Run products in sequence rather than concurrently. Mostly for dev - use only if you want to be especially delicate with system load.                   | bool   | false         |


### Custom Seekers with Configuration
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

Running `hcdiag` with this HCL means we run one seeker: the CLI command `consul version` and the `format = "string"`
tells hcdiag how to parse the result. The `product "consul" {}` block ensures we store the results in the proper location.

Let's go over how to write the custom configuration:

Seekers must be described in a `product` or `host` block. These contain our seekers and tell hcdiag where the store the
results. If a command or file copy is not product specific, `host { ... }` scopes the seeker to the machine the seeker
is run on. The supported product blocks are: `"consul", "vault", "nomad",` and `"terraform-ent"`. A full reference table
of seekers is available in a table below.

Lastly, we'll cover filters. Filters optionally let you remove results from the support bundle. The two options are
`excludes` and `selects`. Each is an array that takes a list of seeker IDs. `exclude` removes matching seekers from the
results and `selects` removes everything that _doesn't_ match the seeker IDs. `selects` take precedence if a seeker matches
for both.

Here's a complete example that describes each of the seekers for one of the product blocks, and host.

```hcl
host {
  command {
    run = "ps aux"
    format = "string"
  }
}

product "consul" {
  command {
    run = "consul version"
    format = "json"
  }

  command {
    run = "consul operator raft list-peers"
    format = "json"
  }

  GET {
    path = "/v1/api/metrics?format=prometheus"
  }

  copy {
    path = "/another/test/log"
    since = "240h"
  }

  excludes = ["consul some-verbose-command"]
  selects = ["consul include this", "consul and this"]
}
```

#### Table of Available Seekers

More will be added as they are made available

| Constructor                | Config Block | Description                                                                                                                                                                            | Parameters                                                       |
|----------------------------|--------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------|
| `seeker.NewCommander(...)` | `command`    | Issues a CLI command and optionally parses the result if format JSON is specified. Otherwise use string.                                                                               | `command = <string,required>` <br/> `format = <string,required>` |
| `seeker.NewCopier(...)`    | `copy`       | Copies the file or directory and all of its contents into the bundle using the same name. Since will check the last modified time of the file and ignore if it's outside the duration. | `path = <string,required>` <br/> `since = <duration,optional>`   |
| `seeker.NewHTTPer(...)`    | `GET`        | Makes an HTTP get request to the path                                                                                                                                                  | `path = <string,required>`                                       |
| `seeker.NewSheller(...)`   | `shell`      | An "escape hatch" allowing arbitrary shell strings to be executed.                                                                                                                     | `run = <string,required>`                                        |


## FAQs

### How do I use hcdiag with Kubernetes?

Although Kubernetes is a complex topic with many configuration options, the key to remember is that hcdiag must be able
to communicate with your pod(s) via a network connection in order to get diagnostic details from the products running in
those pods. If the management interface is not already exposed externally from k8s, you may consider setting up a port-forward
when collecting diagnostics. The command would be similar to `kubectl -n consul port-forward consul-server-0 8500:8500`; 
in this example, we are in the namespace `consul` (noted by `-n consul`), and we are forwarding the external port `8500`
(the port before the `:`) to port `8500` (the port after the `:`) on the pod `consul-server-0`.

If you would like to experiment with a setup like this, assuming you have both minikube and helm installed on your machine,
you could use the following as a reference:

```shell
minikube start
helm repo add hashicorp https://helm.releases.hashicorp.com
helm search repo hashicorp/consul
helm install consul hashicorp/consul --set global.name=consul --create-namespace -n consul
kubectl -n consul port-forward consul-server-0 8500:8500

# Now, run hcdiag
hcdiag -consul
```
