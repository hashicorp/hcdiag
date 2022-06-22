package agent

import "github.com/hashicorp/hcdiag/op"

// ManifestSeeker provides a subset of op state, specifically excluding results, so we can safely render metadata
// about seekers without exposing customer info in manifest.json
type ManifestSeeker struct {
	ID     string    `json:"op"`
	Error  string    `json:"error"`
	Status op.Status `json:"status"`
}
