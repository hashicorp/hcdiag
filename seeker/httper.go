package seeker

import (
	"github.com/hashicorp/host-diagnostics/apiclients"
)

func NewHTTPer(client *apiclients.APIClient, path string) *Seeker {
	return &Seeker{
		Identifier: "GET" + " " + path,
		Runner: HTTPer{
			Client: client,
			Path:   path,
		},
	}
}

// HTTPer hits APIs.
type HTTPer struct {
	Path   string                `json:"path"`
	Client *apiclients.APIClient `json:"client"`
}

func (h HTTPer) Run() (interface{}, error) {
	return h.Client.Get(h.Path)
}
