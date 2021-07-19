package seeker

import (
	"github.com/hashicorp/hcdiag/client"
)

// HTTPer hits APIs.
type HTTPer struct {
	Path   string            `json:"path"`
	Client *client.APIClient `json:"client"`
}

func NewHTTPer(client *client.APIClient, path string) *Seeker {
	return &Seeker{
		Identifier: "GET" + " " + path,
		Runner: HTTPer{
			Client: client,
			Path:   path,
		},
	}
}

func (h HTTPer) Run() (interface{}, error) {
	return h.Client.Get(h.Path)
}
