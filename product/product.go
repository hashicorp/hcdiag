// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package product

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/hcdiag/hcl"
	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/hcdiag/op"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/runner"
)

type Name string

const (
	Consul Name = "consul"
	Host   Name = "host"
	Nomad  Name = "nomad"
	TFE    Name = "terraform-ent"
	Vault  Name = "vault"
)

const (
	DefaultDuration = 10 * time.Second
	DefaultInterval = 5 * time.Second
)

type Config struct {
	Name          Name
	TmpDir        string
	Since         time.Time
	Until         time.Time
	OS            string
	DebugDuration time.Duration
	DebugInterval time.Duration
	HCL           *hcl.Product
	Redactions    []*redact.Redact
}

type Product struct {
	l        hclog.Logger
	Name     Name
	Runners  []runner.Runner
	Excludes []string
	Selects  []string
	Config   Config
}

// Run iterates over the list of runners in a product and returns a map of runner IDs to Ops.
func (p *Product) Run() map[string]op.Op {
	p.l.Info("Running operations for", "product", p.Name)
	results := make(map[string]op.Op)
	for _, r := range p.Runners {
		p.l.Info("running operation", "product", p.Name, "runner", r.ID())
		o := r.Run()
		results[r.ID()] = o
		// Note runner errors to users and keep going.
		if o.Error != nil {
			switch o.Status {
			case op.Fail:
				// TODO(dcohen) This should be p.l.Error, but that outputs an "Error:" line which breaks the test-functional github workflow
				p.l.Warn("result",
					"runner", o.Identifier,
					"status", o.Status,
					"result", fmt.Sprintf("%s", o.Result),
					"error", o.Error,
				)
			case op.Unknown:
				p.l.Warn("result",
					"runner", o.Identifier,
					"status", o.Status,
					"result", fmt.Sprintf("%s", o.Result),
					"error", o.Error,
				)
			case op.Skip:
				p.l.Info("result",
					"runner", o.Identifier,
					"status", o.Status,
					"result", fmt.Sprintf("%s", o.Result),
					"error", o.Error,
				)
			}
		}
	}
	return results
}

// Filter applies our slices of exclude and select runner.ID() matchers to the set of the product's runners.
func (p *Product) Filter() error {
	if p.Runners == nil {
		p.Runners = []runner.Runner{}
	}
	var err error
	// The presence of Selects takes precedence over Excludes
	if p.Selects != nil && 0 < len(p.Selects) {
		p.Runners, err = runner.Select(p.Selects, p.Runners)
		// Skip any Excludes
		return err
	}
	// No Selects, we can apply Excludes
	if p.Excludes != nil {
		p.Runners, err = runner.Exclude(p.Excludes, p.Runners)
	}
	return err
}

// CommandHealthCheck employs the CLI to check if the client and then the agent are available.
func CommandHealthCheck(client, agent string) error {
	return CommandHealthCheckWithContext(context.Background(), client, agent)
}

// CommandHealthCheckWithContext employs the CLI to check if the client and then the agent are available.
func CommandHealthCheckWithContext(ctx context.Context, client, agent string) error {
	clientCmd, err := runner.NewCommandWithContext(ctx, runner.CommandConfig{Command: client})
	if err != nil {
		return err
	}
	checkClient := clientCmd.Run()
	if checkClient.Error != nil {
		return fmt.Errorf("client not available, healthcheck=%v, result=%v, error=%v", client, checkClient.Result, checkClient.Error)
	}

	agentCmd, err := runner.NewCommandWithContext(ctx, runner.CommandConfig{Command: agent})
	if err != nil {
		return err
	}
	checkAgent := agentCmd.Run()
	if checkAgent.Error != nil {
		return fmt.Errorf("agent not available, healthcheck=%v, result=%v, error=%v", agent, checkAgent.Result, checkAgent.Error)
	}

	return nil
}

// CountRunners takes a map of product references and returns a count of all the runners.
func CountRunners(products map[Name]*Product) int {
	var count int
	for _, product := range products {
		count += len(product.Runners)
	}
	return count
}
