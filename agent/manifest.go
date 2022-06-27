package agent

import (
	"github.com/hashicorp/hcdiag/op"
)

// ManifestOp provides a subset of op state, specifically excluding results, so we can safely render metadata
// about ops without exposing customer info in manifest.json
type ManifestOp struct {
	ID     string    `json:"op"`
	Error  string    `json:"error"`
	Status op.Status `json:"status"`
}
