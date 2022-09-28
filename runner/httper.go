package runner

import (
	"github.com/hashicorp/hcdiag/client"
	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"
)

var _ Runner = HTTPer{}

// HTTPer hits APIs.
type HTTPer struct {
	Path       string            `json:"path"`
	Client     *client.APIClient `json:"client"`
	Redactions []*redact.Redact  `json:"redactions"`
}

func NewHTTPer(client *client.APIClient, path string, redactions []*redact.Redact) *HTTPer {
	return &HTTPer{
		Client:     client,
		Path:       path,
		Redactions: redactions,
	}
}

func (h HTTPer) ID() string {
	return "GET" + " " + h.Path
}

// Run executes a GET request to the Path using the Client
func (h HTTPer) Run() op.Op {
	redactedResponse, err := h.Client.RedactGet(h.Path, h.Redactions)
	result := map[string]any{"response": redactedResponse}
	if err != nil {
		op.New(h.ID(), result, op.Fail, err, Params(h))
	}
	return op.New(h.ID(), result, op.Success, nil, Params(h))
}
