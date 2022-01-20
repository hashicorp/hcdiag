package seeker

import (
	"fmt"
	"os/exec"

	"github.com/hashicorp/hcdiag/util"
)

// Sheller runs shell commands in a real unix shell.
type Sheller struct {
	Command string `json:"command"`
	Shell   string `json:"shell"`
}

// NewSheller provides a Seeker for running shell commands.
func NewSheller(command string) *Seeker {
	return &Seeker{
		Identifier: command,
		Runner:     &Sheller{Command: command},
	}
}

func (s *Sheller) Run() (result interface{}, err error) {
	if s.Shell, err = util.GetShell(); err != nil {
		return nil, err
	}
	args := []string{"-c", s.Command}
	bts, err := exec.Command(s.Shell, args...).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("exec.Command error: %s", err)
	}
	return string(bts), err
}
