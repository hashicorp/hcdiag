package runner

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/hashicorp/hcdiag/client"
	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"
)

// HTTPer hits APIs.
type HTTPer struct {
	Path       string            `json:"path"`
	Client     *client.APIClient `json:"client"`
	Redactions []*redact.Redact  `json:"redactions"`
}

func NewHTTPer(client *client.APIClient, path string, redactions []*redact.Redact) *HTTPer {
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
	redResult, redErr := h.redact(fmt.Sprint(result))
	if err != nil {
		if redErr != nil {
			return op.New(h.ID(), nil, op.Fail, redErr, Params(h))
		}
		return op.New(h.ID(), redResult, op.Unknown, err, Params(h))
	}
	return op.New(h.ID(), redResult, op.Success, nil, Params(h))
}

func (h HTTPer) redact(result string) (string, error) {
	r := strings.NewReader(result)
	buf := new(bytes.Buffer)
	err := redact.ApplyMany(h.Redactions, buf, r)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
