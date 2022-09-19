# Custom Configuration

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

Let's go over how to write configuration:

In `hcdiag`, Runners provide an abstraction for any kind of operation. The `command` block above represents a `Command`
Runner, and must be described in a `product` or `host` block. These contain our Runners and tell `hcdiag` where to store
the results. If a command or file copy is not product specific, `host { ... }` scopes the Runner to the local machine.
The supported product blocks are: `"consul", "vault", "nomad",` and `"terraform-ent"`. A full reference table
of Runners is available in a table below.

Filters optionally let you remove Runners before they're Run. Because they're never executed, the results aren't in the
support bundle. The two options are `excludes` and `selects`. Each is an array that takes a list of Runner IDs.
`exclude` removes matching Runners and `selects` removes everything that _doesn't_ match the Runner IDs. `selects`
take precedence if a Runner matches for both.

Here's a complete example that describes each of the Runners for one of the product blocks, and host.

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

## Redactions

Beginning with version `0.4.0`, `hcdiag` supports redactions. Redactions enable users to tell `hcdiag` about patterns of text that should be omitted from the results bundle.

Read more about them [here](./redactions.md).

## Custom Config Examples

### Collect additional Consul information

**Example:** we want a `-consul` run to collect additional information from several commands:

* The Consul memberlist: `consul members`
* Raft peer information: `consul operator raft list-peers`
* A custom journalctl command that's different from the builtin: `journalctl -u consul --since=yesterday --output=json`
* Another custom command (in this case, just running `cat` to collect a file's contents into the bundle `Results.json` file with no redactions)

To do this, we create a custom config file named `hcdiag.hcl` with the following content:

```
product "consul" {
  command {
    run = "consul members"
    format = "string"
  }

  command {
    run = "consul operator raft list-peers"
    format = "string"
  }

  // You could copy files with -includes as well
  command {
    run = "cat /etc/consul.d/example.hcl"
    format = "string"
  }

  // if you want specific journalctl flags, you can use a command:
  command {
    run = "journalctl -u consul --since=yesterday --output=json"
    format = "json"
  }

}
```

Then, a user can run hcdiag with the following command:

```
hcdiag -consul -config hcdiag.hcl
```

This will point hcdiag at your custom config file and execute your custom commands.


### Collect additional Nomad information

**Example:** we want to run a custom `nomad operator debug` command, instead of the built-in. Our command should run with log-level `DEBUG` and return unlimited nodes.

Create a custom config file named `hcdiag.hcl` with the following content:

```
product "nomad" {
  command {
    run = "nomad operator debug -log-level=DEBUG -duration=20s -interval=10s -max-nodes=0"
    format = "string"
  }
}
```

Run hcdiag with this config file:

```
hcdiag -nomad -debug-duration=0 -config hcdiag.hcl
```

This will point hcdiag at your custom config file and execute your custom command.

**NOTE** The -debug-duration flag is here to suppress the built-in nomad debug command, preventing the built-in version of the `nomad operator debug` command from being run, and a second nomad-debug archive being created in your hcdiag bundle.
