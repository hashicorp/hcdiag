package agent

import "time"

// Config stores all user-provided inputs from the CLI and HCL
type Config struct {
	// HostConfig and ProductConfig are specified by HCL
	Host     *HostConfig      `hcl:"host,block" json:"host_config"`
	Products []*ProductConfig `hcl:"product,block" json:"products_config"`

	OS          string    `json:"operating_system"`
	Serial      bool      `json:"serial"`
	Dryrun      bool      `json:"dry_run"`
	Consul      bool      `json:"consul_enabled"`
	Nomad       bool      `json:"nomad_enabled"`
	TFE         bool      `json:"terraform_ent_enabled"`
	Vault       bool      `json:"vault_enabled"`
	Since       time.Time `json:"since"`
	Until       time.Time `json:"until"`
	Includes    []string  `json:"includes"`
	Destination string    `json:"destination"`

	// DebugDuration
	DebugDuration time.Duration `json:"debug_duration"`
	// DebugInterval
	DebugInterval time.Duration `json:"debug_interval"`
}

type HostConfig struct {
	Commands []CommandConfig `hcl:"command,block"`
	Shells   []ShellConfig   `hcl:"shell,block"`
	GETs     []GETConfig     `hcl:"GET,block"`
	Copies   []CopyConfig    `hcl:"copy,block"`
	Excludes []string        `hcl:"excludes,optional"`
	Selects  []string        `hcl:"selects,optional"`
}

type ProductConfig struct {
	Name     string          `hcl:"name,label"`
	Commands []CommandConfig `hcl:"command,block"`
	Shells   []ShellConfig   `hcl:"shell,block"`
	GETs     []GETConfig     `hcl:"GET,block"`
	Copies   []CopyConfig    `hcl:"copy,block"`
	Excludes []string        `hcl:"excludes,optional"`
	Selects  []string        `hcl:"selects,optional"`
}

type CommandConfig struct {
	Run    string `hcl:"run"`
	Format string `hcl:"format"`
}

type ShellConfig struct {
	Run string `hcl:"run"`
}

type GETConfig struct {
	Path string `hcl:"path"`
}

type CopyConfig struct {
	Path  string `hcl:"path"`
	Since string `hcl:"since,optional"`
}
