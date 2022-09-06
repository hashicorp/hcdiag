package host

import (
	"strings"

	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/hcdiag/op"

	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = Get{}

type Get struct {
	Path       string           `json:"path"`
	Redactions []*redact.Redact `json:"redactions"`
}

func NewGetter(path string, redactions []*redact.Redact) *Get {
	return &Get{
		Path:       path,
		Redactions: redactions,
	}
}

func (g Get) ID() string {
	return "GET" + " " + g.Path
}

func (g Get) Run() []op.Op {
	opList := make([]op.Op, 0)
	cmd := strings.Join([]string{"curl -s", g.Path}, " ")
	// NOTE(mkcp): We will get JSON back from a lot of requests, so this can be improved
	format := "string"
	o := runner.NewCommander(cmd, format, g.Redactions).Run()
	return append(opList, op.New(g.ID(), o[0].Result, o[0].Status, o[0].Error, runner.Params(g)))

}
