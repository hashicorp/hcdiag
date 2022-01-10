package seeker

import (
	"github.com/hashicorp/hcdiag/util"
)

// Sheller runs shell commands in a real unix shell.
type Sheller struct {
	Command string `json:"command"`
}

// NewSheller provides a Seeker for running shell commands.
func NewSheller(command string) *Seeker {
	return &Seeker{
		Identifier: command,
		Runner:     Sheller{Command: command},
	}
}

func (s Sheller) Run() (result interface{}, err error) {
	return util.ShellExec(s.Command)
}
