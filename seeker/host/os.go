package host

import (
	"github.com/hashicorp/hcdiag/seeker"
)

var _ seeker.Runner = OS{}

func NewOS(os string) *seeker.Seeker {
	osCmd := "uname -v"
	if os == "windows" {
		osCmd = "systeminfo"
	}

	return &seeker.Seeker{
		Identifier: osCmd,
		Runner: OS{
			OS:      os,
			command: osCmd,
		},
	}
}

type OS struct {
	OS      string `json:"os"`
	command string
}

// Run calls the given OS utility to get information on the operating system
func (o OS) Run() (interface{}, seeker.Status, error) {
	// NOTE(mkcp): This seeker can be made consistent between multiple operating systems if we parse the output of
	//   systeminfo to match uname's scope of concerns.
	return seeker.NewCommander(o.command, "string").Runner.Run()
}
