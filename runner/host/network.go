// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package host

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/runner"
	"github.com/shirou/gopsutil/v3/net"
)

var _ runner.Runner = &Network{}

// NetworkInterface represents details about a network interface. This serves as the basis for the results produced
// by the Network runner.
type NetworkInterface struct {
	Index int      `json:"index"`
	MTU   int      `json:"mtu"`
	Name  string   `json:"name"`
	Flags []string `json:"flags"`
	Addrs []string `json:"addrs"`
}

type NetworkConfig struct {
	// Redactions includes any redactions to apply to the output of the runner.
	Redactions []*redact.Redact
	// Timeout specifies the amount of time that the runner should be allowed to execute before cancellation.
	Timeout time.Duration
}

type Network struct {
	ctx context.Context

	// Redactions includes any redactions to apply to the output of the runner.
	Redactions []*redact.Redact `json:"redactions"`
	// Timeout specifies the amount of time that the runner should be allowed to execute before cancellation.
	Timeout runner.Timeout `json:"timeout"`
}

func NewNetwork(cfg NetworkConfig) *Network {
	return NewNetworkWithContext(context.Background(), cfg)
}

func NewNetworkWithContext(ctx context.Context, cfg NetworkConfig) *Network {
	return &Network{
		ctx:        ctx,
		Redactions: cfg.Redactions,
		Timeout:    runner.Timeout(cfg.Timeout),
	}
}

func (n Network) ID() string {
	return "network"
}

func (n Network) Run() op.Op {
	startTime := time.Now()

	if n.ctx == nil {
		n.ctx = context.Background()
	}

	resChan := make(chan op.Op, 1)
	runCtx := n.ctx
	var cancel context.CancelFunc
	if 0 < n.Timeout {
		runCtx, cancel = context.WithTimeout(n.ctx, time.Duration(n.Timeout))
		defer cancel()
	}

	go func(resChan chan op.Op) {
		o := n.run()
		o.Start = startTime
		resChan <- o
	}(resChan)

	select {
	case <-runCtx.Done():
		switch runCtx.Err() {
		case context.Canceled:
			return runner.CancelOp(n, runCtx.Err(), startTime)
		case context.DeadlineExceeded:
			return runner.TimeoutOp(n, runCtx.Err(), startTime)
		default:
			return op.New(n.ID(), nil, op.Unknown, runCtx.Err(), runner.Params(n), startTime, time.Now())
		}
	case o := <-resChan:
		return o
	}
}

func (n Network) run() op.Op {
	result := make(map[string]any)

	netIfs, err := net.Interfaces()
	if err != nil {
		hclog.L().Trace("runner/host.Network.Run()", "error", err)
		return op.New(n.ID(), result, op.Fail, err, runner.Params(n), time.Time{}, time.Now())
	}

	for _, netIf := range netIfs {
		ifce, err := n.networkInterface(netIf)
		if err != nil {
			hclog.L().Trace("runner/host.Network.Run()", "error", err)
			err1 := fmt.Errorf("error converting network information err=%w", err)
			return op.New(n.ID(), result, op.Fail, err1, runner.Params(n), time.Time{}, time.Now())
		}
		result[ifce.Name] = ifce
	}
	return op.New(n.ID(), result, op.Success, nil, runner.Params(n), time.Time{}, time.Now())
}

func (n Network) networkInterface(nis net.InterfaceStat) (NetworkInterface, error) {
	netIf := NetworkInterface{
		Index: nis.Index,
		MTU:   nis.MTU,
	}

	name, err := redact.String(nis.Name, n.Redactions)
	if err != nil {
		return NetworkInterface{}, err
	}
	netIf.Name = name

	// Make a slice rather than an empty var declaration to avoid later marshalling null instead of empty array
	flags := make([]string, 0)
	for _, f := range nis.Flags {
		flag, err := redact.String(f, n.Redactions)
		if err != nil {
			return NetworkInterface{}, err
		}
		flags = append(flags, flag)
	}
	netIf.Flags = flags

	// Make a slice rather than an empty var declaration to avoid later marshalling null instead of empty array
	addrs := make([]string, 0)
	for _, a := range nis.Addrs {
		addr, err := redact.String(a.Addr, n.Redactions)
		if err != nil {
			return NetworkInterface{}, err
		}
		addrs = append(addrs, addr)
	}
	netIf.Addrs = addrs

	return netIf, nil
}
