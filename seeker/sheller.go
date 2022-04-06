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

// Run ensures a shell exists and optimistically executes the given Command string
func (s *Sheller) Run() (interface{}, Status, error) {
	// Read the shell from the environment
	shell, err := util.GetShell()
	if err != nil {
		return nil, Fail, err
	}
	s.Shell = shell

	// Run the command
	args := []string{"-c", s.Command}
	bts, err := exec.Command(s.Shell, args...).CombinedOutput()
	if err != nil {
		// Return the stdout result even on failure
		// TODO(mkcp): This is a good place to check for more kinds of errors
		return string(bts), Unknown, fmt.Errorf("exec.Command error: %s", err)
	}

	return string(bts), Success, nil
}
