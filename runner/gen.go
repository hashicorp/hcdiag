package runner

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/hashicorp/hcdiag/op"
)

var _ Runner = Gen{}

type Gen struct {
	Name        string
	Description string
	Pattern     Runner
	Collection  []string
}

func NewGen(name string, coll []string, pattern Runner, description string) Gen {
	return Gen{
		Name:        name,
		Description: description,
		Pattern:     pattern,
		Collection:  coll,
	}
}

func (g Gen) ID() string {
	return "gen " + g.Name
}

func (g Gen) Run() op.Op {
	result := map[string]any{}
	l := sync.Mutex{}

	runners, err := build(g.Pattern, g.Collection)
	if err != nil {
		panic("well shoot!")
	}

	wg := sync.WaitGroup{}
	wg.Add(len(runners))

	for _, r := range runners {
		go func(r Runner, l *sync.Mutex, group *sync.WaitGroup) {
			o := r.Run()
			l.Lock()
			result[r.ID()] = o
			l.Unlock()
			wg.Done()
		}(r, &l, &wg)
	}
	wg.Wait()

	return op.Op{
		Result: result,
		Status: op.Success,
	}
}

func build(pattern Runner, input []string) ([]Runner, error) {
	result := make([]Runner, len(input))
	for i, in := range input {
		switch any(pattern).(type) {
		case Commander:
			p := pattern.(Commander)
			cmd := NewCommander(in, p.Format, p.Redactions)
			result[i] = cmd
		case Sheller:
			p := pattern.(Sheller)
			shell := NewSheller(in, p.Redactions)
			result[i] = shell
		default:
			return nil, fmt.Errorf("unknown runner kind, runner=%s", reflect.TypeOf(pattern))
		}

	}
	return result, nil
}
