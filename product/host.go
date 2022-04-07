package product

import (
	"runtime"

	"github.com/hashicorp/hcdiag/seeker"
	s "github.com/hashicorp/hcdiag/seeker"
)

// NewHost takes a product config and creates a Product containing all of the host's seekers.
func NewHost(cfg Config) *Product {
	return &Product{
		Seekers: HostSeekers(cfg.OS),
	}
}

// HostSeekers checks the operating system and passes it into the seekers.
func HostSeekers(os string) []*s.Seeker {
	if os == "auto" {
		os = runtime.GOOS
	}
	return []*s.Seeker{
		{
			Identifier: "stats",
			Runner: &seeker.Host{
				OS: os,
			},
		},
	}
}
