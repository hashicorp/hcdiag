package host

import (
	"strings"

	"github.com/hashicorp/hcdiag/op"

	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = Get{}

type Get struct {
	Path string `json:"path"`
}

func NewGetter(path string) *Get {
	return &Get{path}
}

func (g Get) ID() string {
	return "GET" + " " + g.Path
}

func (g Get) Run() op.Op {
	cmd := strings.Join([]string{"curl -s", g.Path}, " ")
	// NOTE(mkcp): We will get JSON back from a lot of requests, so this can be improved
	format := "string"
	o := runner.NewCommander(cmd, format).Run()
	return op.New(g.ID(), o.Result, o.Status, o.Error, runner.Params(g))

}
