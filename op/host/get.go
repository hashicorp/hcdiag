package host

import (
	"strings"

	"github.com/hashicorp/hcdiag/op"
)

var _ op.Runner = Get{}

type Get struct {
	path string
}

func NewGetter(path string) *Get {
	return &Get{path}
}

func (g Get) ID() string {
	return "GET" + " " + g.path
}

func (g Get) Run() op.Op {
	cmd := strings.Join([]string{"curl -s", g.path}, " ")
	// NOTE(mkcp): We will get JSON back from a lot of requests, so this can be improved
	format := "string"
	o := op.NewCommander(cmd, format).Run()
	return op.New(g, o.Result, o.Status, o.Error)

}
