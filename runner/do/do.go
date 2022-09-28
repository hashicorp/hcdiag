package do

import (
	"fmt"
	"sync"

	"github.com/hashicorp/hcdiag/runner"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/op"
)

var _ runner.Runner = Do{}

// Do runs shell commands.
type Do struct {
	Runners []runner.Runner `json:"runners"`
	log     hclog.Logger
}

// New returns a pointer to a new Do runner
func New(l hclog.Logger, runners []runner.Runner) *Do {
	return &Do{
		Runners: runners,
		log:     l,
	}
}

func (d Do) ID() string {
	return "do"
}

// Run asynchronously executes all Runners and returns the resulting ops when the last one has finished
func (d Do) Run() op.Op {
	var wg sync.WaitGroup
	m := sync.Map{}

	wg.Add(len(d.Runners))
	for _, r := range d.Runners {
		d.log.Info("running operation", "runner", r.ID())
		go func(results *sync.Map, r runner.Runner) {
			m.Store(r.ID(), r.Run())
			wg.Done()
		}(&m, r)
	}
	wg.Wait()

	return op.New(d.ID(), snapshot(&m), op.Success, nil, runner.Params(d))
}

func snapshot(s *sync.Map) map[string]any {
	m := make(map[string]any)
	s.Range(func(k, v any) bool {
		m[fmt.Sprint(k)] = v
		return true
	})
	return m
}
