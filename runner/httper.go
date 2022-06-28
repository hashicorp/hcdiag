package runner

import (
	"github.com/hashicorp/hcdiag/client"
	"github.com/hashicorp/hcdiag/op"
)

// HTTPer hits APIs.
type HTTPer struct {
	Path   string            `json:"path"`
	Client *client.APIClient `json:"client"`
}

func NewHTTPer(client *client.APIClient, path string) *HTTPer {
	return &HTTPer{
		Client: client,
		Path:   path,
	}
}

func (h HTTPer) ID() string {
	return "GET" + " " + h.Path
}

// Run executes a GET request to the Path using the Client
func (h HTTPer) Run() op.Op {
	result, err := h.Client.Get(h.Path)
	if err != nil {
		return op.New(h.ID(), result, op.Unknown, err, Params(h))
	}
	return op.New(h.ID(), result, op.Success, nil, Params(h))
}
