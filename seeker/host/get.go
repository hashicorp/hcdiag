package host

import (
	"strings"

	"github.com/hashicorp/hcdiag/seeker"
)

var _ seeker.Runner = Get{}

type Get struct {
	path string
}

func NewGetter(path string) *seeker.Seeker {
	return &seeker.Seeker{
		Identifier: "GET" + " " + path,
		Runner: Get{
			path: path,
		},
	}
}

func (g Get) Run() (interface{}, seeker.Status, error) {
	cmd := strings.Join([]string{"curl -s", g.path}, " ")
	// NOTE(mkcp): We will get JSON back from a lot of requests, so this can be improved
	format := "string"
	return seeker.NewCommander(cmd, format).Runner.Run()
}
