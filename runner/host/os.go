package host

import (
	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/runner"
)

var _ runner.Runner = OS{}

type OS struct {
	OS      string `json:"os"`
	Command string `json:"command"`
}

func NewOS(os string) *OS {
	osCmd := "uname -v"
	if os == "windows" {
		osCmd = "systeminfo"
	}

	return &OS{
		OS:      os,
		Command: osCmd,
	}
}

func (o OS) ID() string {
	return o.Command
}

// Run calls the given OS utility to get information on the operating system
func (o OS) Run() op.Op {
	// NOTE(mkcp): This runner can be made consistent between multiple operating systems if we parse the output of
	//   systeminfo to match uname's scope of concerns.
	c := runner.NewCommander(o.Command, "string").Run()
	return op.New(o.ID(), c.Result, c.Status, c.Error, runner.Params(o))
}
