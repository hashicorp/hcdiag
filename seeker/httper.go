package seeker

import (
	"github.com/hashicorp/host-diagnostics/apiclients"
)

func NewHTTPer(client *apiclients.APIClient, path string, mustSucceed bool) *Seeker {
	return &Seeker{
		Identifier: client.Product + " " + path,
		Runner: HTTPer{
			Client: client,
			Path:   path,
		},
		MustSucceed: mustSucceed,
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
