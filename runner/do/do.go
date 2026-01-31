// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package do

import (
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/hcdiag/runner"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
)

var _ runner.Runner = Do{}

// Do is a runner that wraps a collection of runners and executes each of them in a goroutine. It returns a map associating
// each runner's Op to its ID.
type Do struct {
	Runners     []runner.Runner `json:"runners"`
	Label       string          `json:"label"`
	Description string          `json:"description"`
	log         hclog.Logger
}

// New initializes a Do runner.
func New(l hclog.Logger, label, description string, runners []runner.Runner) *Do {
	return &Do{
		Label:       label,
		Description: description,
		Runners:     runners,
		log:         l,
	}
}

func (d Do) ID() string {
	return "do " + d.Label
}

// Run asynchronously executes all Runners and returns the resulting ops when the last one has finished
func (d Do) Run() op.Op {
	startTime := time.Now()
	var wg sync.WaitGroup
	m := sync.Map{}

	wg.Add(len(d.Runners))
	for _, r := range d.Runners {
		d.log.Info("running operation", "runner", r.ID())
		go func(results *sync.Map, r runner.Runner) {
			results.Store(r.ID(), r.Run())
			wg.Done()
		}(&m, r)
	}
	wg.Wait()

	return op.New(d.ID(), snapshot(&m), op.Success, nil, runner.Params(d), startTime, time.Now())
}

func snapshot(s *sync.Map) map[string]any {
	m := make(map[string]any)
	s.Range(func(k, v any) bool {
		m[fmt.Sprint(k)] = v
		return true
	})
	return m
}
