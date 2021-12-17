package agent

import "time"

type Config struct {
	Host     *HostConfig      `hcl:"host,block" json:"-"`
	Products []*ProductConfig `hcl:"product,block" json:"-"`

	OS          string    `json:"operating_system"`
	Serial      bool      `json:"serial"`
	Dryrun      bool      `json:"dry_run"`
	Consul      bool      `json:"consul_enabled"`
	Nomad       bool      `json:"nomad_enabled"`
	TFE         bool      `json:"terraform_ent_enabled"`
	Vault       bool      `json:"vault_enabled"`
	Includes    []string  `json:"includes"`
	IncludeFrom time.Time `json:"include_from"`
	IncludeTo   time.Time `json:"include_to"`
	Destination string    `json:"destination"`
}

type HostConfig struct {
	Commands []CommandConfig `hcl:"command,block"`
	GETs     []GETConfig     `hcl:"GET,block"`
	Copies   []CopyConfig    `hcl:"copy,block"`
	Excludes []string        `hcl:"excludes,optional"`
	Selects  []string        `hcl:"selects,optional"`
}

type ProductConfig struct {
	Name     string          `hcl:"name,label"`
	Commands []CommandConfig `hcl:"command,block"`
	GETs     []GETConfig     `hcl:"GET,block"`
	Copies   []CopyConfig    `hcl:"copy,block"`
	Excludes []string        `hcl:"excludes,optional"`
	Selects  []string        `hcl:"selects,optional"`
}

type CommandConfig struct {
	Run    string `hcl:"run"`
	Format string `hcl:"format"`
}

type GETConfig struct {
	Path string `hcl:"path"`
}

type CopyConfig struct {
	Path string `hcl:"path"`
	// FIXME(mkcp): This should be a duration that we parse
	Since string `hcl:"since,optional"`
}

const ExampleConfig = `host {
  command {
    run    = "sar -q"
    format = "string"
  }
}

product "vault" {
  command {
    run    = "journalctl -u vault > {{ tmpDir }}/vault.log"
    format = "string"
  }

  GET {
    path = "/v1/sys/mounts"
  }
}
`
