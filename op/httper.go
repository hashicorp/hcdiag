package op

import (
	"github.com/hashicorp/hcdiag/client"
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
func (h HTTPer) Run() Op {
	result, err := h.Client.Get(h.Path)
	if err != nil {
		return h.op(result, Unknown, err)
	}
	return h.op(result, Success, nil)
}

func (h HTTPer) op(result interface{}, status Status, err error) Op {
	return Op{
		Identifier: h.ID(),
		Result:     result,
		Error:      err,
		ErrString:  err.Error(),
		Status:     status,
		Params:     map[string]interface{}{"path": h.Path},
	}
}
