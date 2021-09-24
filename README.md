# hcdiag

The purpose of this tool is to simplify the collection of relevant support data for HashiCorp products. The tool will execute a number of commands to collect system information (./hcdiag) and product information for any number of supported products. The output of these commands is stored in a temporary directory and bundled into a single archive as an output. You may perform a `-dryrun` to view these commands before executing them and you can also see them clearly defined in code in the aforementioned locations.

## Prerequisites

The following subsections cover product specific prerequisite items such as environment variables required to successfully communicate with a given environment.

### Consul

- CLI
    - Consul CLI documentation is available [here](https://www.consul.io/commands/index)
    - The [Consul binary](https://www.consul.io/downloads) must be available on the local machine
- API
    - Consul API documentation is available [here](https://www.consul.io/api-docs)
    - Environment variable [CONSUL_HTTP_ADDR](https://www.consul.io/commands#consul_http_addr) must be set to the HTTP address to the local Consul agent
    - Environment variable [CONSUL_TOKEN](https://www.consul.io/commands#consul_http_token) may be set to the API access token if ACLs are enabled.

### Nomad

- CLI
    - Nomad CLI documentation is available [here](https://www.nomadproject.io/docs/commands)
    - The [Nomad binary](https://www.nomadproject.io/downloads) must be available on the local machine
- API
    - Nomad API documentation is available [here](https://www.nomadproject.io/api-docs)
    - Environment variable [NOMAD_ADDR](https://www.nomadproject.io/docs/commands#nomad_addr) must be set to the HTTP address of the Nomad server
    - Environment variable [NOMAD_TOKEN](https://www.nomadproject.io/docs/commands#nomad_token) may be set to the SecretID of an ACL token for API requests if ACLs are enabled.

### Terraform Enterprise/Cloud

- CLI
    - Replicated CLI documentation is available [here](https://help.replicated.com/api/replicatedctl/)
    - The Terraform binary is not currently required on the local machine
    - CLI currently limited to self-managed TFE environments
- API
    - Terraform Enterprise/Cloud API documentation is available [here](https://www.terraform.io/docs/cloud/api/index.html)
    - Environment variable `TFE_HTTP_ADDR` must be set to the HTTP address of a Terraform Enterprise or Terraform Cloud environment
    - Environment variable [TFE_TOKEN](https://www.terraform.io/docs/cloud/api/index.html#authentication) must be set to an appropriate bearer token (user, team, or organization)

### Vault

- CLI
    - Vault CLI documentation is available [here](https://www.vaultproject.io/docs/commands)
    - The [Vault binary](https://www.vaultproject.io/downloads) must be available on the local machine
- API
    - Vault API documentation is available [here](https://www.vaultproject.io/api)
    - Environment variable [VAULT_ADDR](https://www.vaultproject.io/docs/commands#vault_addr) must be set to the HTTP address of the Vault server
    - Environment variable [VAULT_TOKEN](https://www.vaultproject.io/docs/commands#vault_token) must be set to the Vault authentication token
        - Alternatively, a token may also exist at ~/.vault-token
        - If both are present, `VAULT_TOKEN` will be used.

## Usage

### Input

| Argument | Description | Type | Default Value |
|----------|-------------|------| ------------- |
| `dryrun` | Perform a dry run to display commands without executing them | bool | false |
| `os` | Override operating system detection | string | "auto" |
| `consul` | Run Consul diagnostics | bool | false |
| `nomad` | Run Nomad diagnostics | bool | false |
| `terraform-ent` | Run Terraform Enterprise/Cloud diagnostics | bool | false |
| `vault` | Run Vault diagnostics | bool | false |
| `all` | Run all available product diagnostics | bool | false |
| `includes` | files or directories to include (comma-separated, file-*-globbing available if 'wrapped-*-in-single-quotes') e.g. '/var/log/consul-*,/var/log/nomad-*' | string | "" |
| `include-since` | Time range to include files, counting back from now. Takes a 'go-formatted' duration, usage examples: `72h`, `25m`, `45s`, `120h1m90s` | string | "72h" |
| `destination` | Path to the directory the bundle should be written in | string | "." |
| `dest` | Shorthand for -destination | string | "." |
| `config` | Path to HCL configuration file | string | "" |
| `serial` | Run products in sequence rather than concurrently. Mostly for dev - use only if you want to be especially delicate with system load. | bool | false |

###  Examples

- Host diagnostics only  
    - `hcdiag`

- Host and Vault diagnostics  
    - `hcdiag -vault`

- Host, Consul, and Nomad diagnostics  
    - `hcdiag -consul -nomad`

- Host and all available product diagnostics  
    - `hcdiag -all`


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

We offer 3 seeker types: `command`, `GET`, and `copy`. `command` describes a CLI command to execute. `GET` describes an
HTTP request to the corresponding product at the path extension. Finally, `copy` copies files over into the support
bundle. 

Seekers must be described in a `product` or `host` block. These contain our seekers and tell hcdiag where the store the
results. If a command or file copy is not product specific, `host { ... }` scopes the seeker to the machine the seeker
is run on. The supported product blocks are: `"consul", "vault", "nomad",` and `"terraform-ent"`.

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
    since = "10d"
  }

  excludes = ["consul some-verbose-command"]
  selects = ["consul include this", "consul and this"]
}
```
