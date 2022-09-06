package runner

import (
	"sync"

	"github.com/hashicorp/hcdiag/op"
)

var _ Runner = Do{}

// Do runs shell commands.
type Do struct {
	Runners []Runner `json:"runners"`
}

// NewDo returns a pointer to a new Do runner
func NewDo(runners []Runner) *Do {
	return &Do{
		Runners: runners,
	}
}

func (d Do) ID() string {
	return "Do"
}

// Run asynchronously executes all Runners and returns the resulting ops when the last one has finished
func (d Do) Run() []op.Op {
	// sortMap maps each runner-index to its resulting ops, so that we avoid a non-deterministically-ordered
	// slice that includes ops mixed together from *all* runners
	sortMap := make(map[int][]op.Op)
	// mu protects sortMap, which is going to be concurrently accessed by our runner goroutines
	var mu sync.Mutex

	// Start WaitGroup that will wait for all of our Runners
	wg := new(sync.WaitGroup)
	wg.Add(len(d.Runners))

	for i, r := range d.Runners {
		go func(m *sync.Mutex, sm *map[int][]op.Op, r Runner, ridx int) {
			// Dereference pointer TODO(is there a better way?)
			sortMap := *sm
			o := r.Run()

			mu.Lock()
			sortMap[ridx] = o
			mu.Unlock()
		}(&mu, &sortMap, r, i)
	}

	wg.Wait()
	// TODO(dcohen) add success op for the Do block when we know what that should look like
	return buildResults(sortMap)
}

// buildResults takes a map of runner indexes -> run Ops and returns a flat slice of Ops
func buildResults(sortMap map[int][]op.Op) []op.Op {
	results := make([]op.Op, 0)

	for i := 0; i < len(sortMap); i++ {
		// Add all ops that runner i produced to the results Op slice
		results = append(results, sortMap[i]...)
	}
	return results
}
