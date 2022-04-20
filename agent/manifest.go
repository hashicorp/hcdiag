package agent

import "github.com/hashicorp/hcdiag/seeker"

// ManifestSeeker provides a subset of seeker state, specifically excluding results, so we can safely render metadata
// about seekers without exposing customer info in manifest.json
type ManifestSeeker struct {
	ID     string        `json:"seeker"`
	Error  string        `json:"error"`
	Status seeker.Status `json:"status"`
}
