package host

import (
	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/util"
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
	return op.Op{
		Identifier: o.ID(),
		Result:     c.Result,
		Error:      c.Error,
		ErrString:  c.ErrString,
		Status:     c.Status,
		Params:     util.RunnerParams(o),
	}
}
