package host

import (
	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = OS{}

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
func (o OS) Run() runner.Op {
	// NOTE(mkcp): This runner can be made consistent between multiple operating systems if we parse the output of
	//   systeminfo to match uname's scope of concerns.
	c := runner.NewCommander(o.command, "string").Run()
	return runner.New(o, c.Result, c.Status, c.Error)
}
