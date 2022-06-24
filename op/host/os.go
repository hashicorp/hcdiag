package host

import (
	"github.com/hashicorp/hcdiag/op"
)

var _ op.Runner = OS{}

func NewOS(os string) *OS {
	osCmd := "uname -v"
	if os == "windows" {
		osCmd = "systeminfo"
	}

	return &OS{
		os:      os,
		command: osCmd,
	}
}

func (o OS) ID() string {
	return o.command
}

type OS struct {
	os      string
	command string
}

// Run calls the given OS utility to get information on the operating system
func (o OS) Run() op.Op {
	// NOTE(mkcp): This op can be made consistent between multiple operating systems if we parse the output of
	//   systeminfo to match uname's scope of concerns.
	c := op.NewCommander(o.command, "string").Run()
	return op.New(o, c.Result, c.Status, c.Error)
}
