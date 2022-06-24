package host

import (
	"strings"

	"github.com/hashicorp/hcdiag/op"
)

var _ op.Runner = Get{}

type Get struct {
	path string
}

func NewGetter(path string) *op.Op {
	return &op.Op{
		Identifier: "GET" + " " + path,
		Runner: Get{
			path: path,
		},
	}
}

func (g Get) Run() (interface{}, op.Status, error) {
	cmd := strings.Join([]string{"curl -s", g.path}, " ")
	// NOTE(mkcp): We will get JSON back from a lot of requests, so this can be improved
	format := "string"
	return op.NewCommander(cmd, format).Runner.Run()
}
