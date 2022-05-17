package product

import (
	"runtime"

	"github.com/hashicorp/hcdiag/seeker"
	"github.com/hashicorp/hcdiag/seeker/host"
)

// NewHost takes a product config and creates a Product containing all of the host's seekers.
func NewHost(cfg Config) *Product {
	return &Product{
		Seekers: HostSeekers(cfg.OS),
	}
}

// HostSeekers checks the operating system and passes it into the seekers.
func HostSeekers(os string) []*seeker.Seeker {
	if os == "auto" {
		os = runtime.GOOS
	}
	return []*seeker.Seeker{
		host.NewOS(os),
		host.NewDisk(),
		host.NewInfo(),
		host.NewMemory(),
		host.NewProcess(),
		host.NewNetwork(),
		host.NewEtcHosts(),
	}
}
