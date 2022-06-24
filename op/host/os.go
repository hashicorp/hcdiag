package host

import (
	"github.com/hashicorp/hcdiag/op"
)

var _ op.Runner = OS{}

func NewOS(os string) *op.Op {
	osCmd := "uname -v"
	if os == "windows" {
		osCmd = "systeminfo"
	}

	return &op.Op{
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
func (o OS) Run() (interface{}, op.Status, error) {
	// NOTE(mkcp): This op can be made consistent between multiple operating systems if we parse the output of
	//   systeminfo to match uname's scope of concerns.
	return op.NewCommander(o.command, "string").Runner.Run()
}
