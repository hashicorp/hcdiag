package host

import (
	"strings"
	"time"

	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/hcdiag/op"

	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = Get{}

type Get struct {
	Path       string           `json:"path"`
	Redactions []*redact.Redact `json:"redactions"`
}

func NewGet(path string, redactions []*redact.Redact) *Get {
	return &Get{
		Path:       path,
		Redactions: redactions,
	}
}

func (g Get) ID() string {
	return "GET" + " " + g.Path
}

func (g Get) Run() op.Op {
	startTime := time.Now()

	cmd := strings.Join([]string{"curl -s", g.Path}, " ")
	// NOTE(mkcp): We will get JSON back from a lot of requests, so this can be improved
	format := "string"
	o := runner.NewCommand(cmd, format, g.Redactions).Run()
	return op.New(g.ID(), o.Result, o.Status, o.Error, runner.Params(g), startTime, time.Now())
}
