package runner

import (
	"time"

	"github.com/hashicorp/hcdiag/client"
	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"
)

var _ Runner = HTTP{}

// HTTP hits APIs.
type HTTP struct {
	Path       string            `json:"path"`
	Client     *client.APIClient `json:"client"`
	Redactions []*redact.Redact  `json:"redactions"`
}

func NewHTTP(client *client.APIClient, path string, redactions []*redact.Redact) *HTTP {
	return &HTTP{
		Client:     client,
		Path:       path,
		Redactions: redactions,
	}
}

func (h HTTP) ID() string {
	return "GET" + " " + h.Path
}

// Run executes a GET request to the Path using the Client
func (h HTTP) Run() op.Op {
	startTime := time.Now()

	redactedResponse, err := h.Client.RedactGet(h.Path, h.Redactions)
	result := map[string]any{"response": redactedResponse}
	if err != nil {
		op.New(h.ID(), result, op.Fail, err, Params(h), startTime, time.Now())
	}

	return op.New(h.ID(), result, op.Success, nil, Params(h), startTime, time.Now())
}
