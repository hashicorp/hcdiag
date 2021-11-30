package command

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/agent"
	"github.com/mitchellh/cli"
	"time"
)

func Run(log hclog.Logger) {
	// Migrate this to a run command?
	cfg := agent.Config{
		Host:        nil,
		Products:    nil,
		OS:          "",
		Serial:      false,
		Dryrun:      false,
		Consul:      false,
		Nomad:       false,
		TFE:         false,
		Vault:       false,
		Includes:    nil,
		IncludeFrom: time.Time{},
		IncludeTo:   time.Time{},
		Destination: "",
	}
	a := agent.NewAgent(cfg, log)
	a.Run()
}

func Command(x cli.Command) (cli.Command, error) {
	return x, nil
}
