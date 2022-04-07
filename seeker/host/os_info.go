package host

import (
	"github.com/hashicorp/hcdiag/seeker"
)

var _ seeker.Runner = OSInfo{}

type OSInfo struct {
	OS      string `json:"os"`
	command string
}

// NewOSInfo calls the utility to get information on the operating system
func NewOSInfo(os string) *seeker.Seeker {
	osInfoCmd := "uname -v"
	if os == "windows" {
		osInfoCmd = "systeminfo"
	}

	return &seeker.Seeker{
		Identifier: osInfoCmd,
		Runner: OSInfo{
			OS:      os,
			command: osInfoCmd,
		},
	}
}

func (o OSInfo) Run() (interface{}, seeker.Status, error) {
	// NOTE(mkcp): This seeker can be made consistent between multiple operating systems if we parse the output of
	//   systeminfo to match uname's scope of concerns.
	return seeker.NewCommander(o.command, "string").Runner.Run()
}
