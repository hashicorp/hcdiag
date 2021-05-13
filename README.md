# host-diagnostics

The purpose of this tool is to simplify the collection of relevant support data for HashiCorp products. The tool will execute a number of commands to collect system information (./hostdiag) and product information (./products) for any number of supported products. The output of these commands is stored in a temporary directory and bundled into a single archive as an output. You may perform a `-dryrun` to view these commands before executing them and you can also see them clearly defined in code in the aforementioned locations.

## Prerequisites

The following subsections cover product specific prerequisite items such as environment variables required to successfully communicate with a given environment.

### Consul

- CLI
    - Consul CLI documentation is available [here](https://www.consul.io/commands/index)
    - The [Consul binary](https://www.consul.io/downloads) must be available on the local machine
- API
    - Consul API documentation is available [here](https://www.consul.io/api-docs)
    - Environment variable [CONSUL_HTTP_ADDR](https://www.consul.io/commands#consul_http_addr) must be set to the HTTP address to the local Consul agent
    - Environment variable [CONSUL_TOKEN](https://www.consul.io/commands#consul_http_token) must be set to the API access token

### Nomad

- CLI
    - Nomad CLI documentation is available [here](https://www.nomadproject.io/docs/commands)
    - The [Nomad binary](https://www.nomadproject.io/downloads) must be available on the local machine
- API
    - Nomad API documentation is available [here](https://www.nomadproject.io/api-docs)
    - Environment variable [NOMAD_ADDR](https://www.nomadproject.io/docs/commands#nomad_addr) must be set to the HTTP address of the Nomad server
    - Environment variable [NOMAD_TOKEN](https://www.nomadproject.io/docs/commands#nomad_token) must be set to the SecretID of an ACL token for API requests

### Terraform Enterprise/Cloud

- CLI
    - Replicated CLI documentation is available [here](https://help.replicated.com/api/replicatedctl/)
    - The Terraform binary is not currently required on the local machine
    - CLI currently limited to self-managed TFE environments
- API
    - Terraform Enterprise/Cloud API documentation is available [here](https://www.terraform.io/docs/cloud/api/index.html)
    - Environment variable [TFE_HTTP_ADDR]() ...
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

## Usage

### Input

| Argument | Description | Type | Default Value |
|------|-------------|------| ------------- |
| `dryrun` | Perform a dry run to display commands without executing them | bool | false |
| `os` | Override operating system detection | string | "auto" |
| `consul` | Run Consul diagnostics | bool | false |
| `nomad` | Run Nomad diagnostics | bool | false |
| `tfe` | Run Terraform Enterprise/Cloud diagnostics | bool | false |
| `vault` | Run Vault diagnostics | bool | false |
| `all` | Run all available product diagnostics | bool | false |
| `includes` | Files or directories to include | string | "" |
| `outfile` | Output file name | string | "support.tar.gz" |

###  Examples

- Host diagnostics only  
    - `<future_binary_name>`

- Host and Vault diagnostics  
    - `<future_binary_name> -vault`

- Host, Consul, and Nomad diagnostics  
    - `<future_binary_name> -consul -nomad`

- Host and all available product diagnostics  
    - `<future_binary_name> -all`
