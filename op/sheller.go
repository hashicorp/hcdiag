package op

import (
	"fmt"
	"os/exec"

	"github.com/hashicorp/hcdiag/util"
)

// Sheller runs shell commands in a real unix shell.
type Sheller struct {
	id      string
	command string
	shell   string
}

// NewSheller provides a Op for running shell commands.
func NewSheller(command string) *Sheller {
	return &Sheller{
		command: command,
	}
}

func (s Sheller) ID() string {
	return s.command
}

// Run ensures a shell exists and optimistically executes the given Command string
func (s Sheller) Run() Op {
	// Read the shell from the environment
	shell, err := util.GetShell()
	if err != nil {
		return New(s, nil, Fail, err)
	}
	s.shell = shell

	// Run the command
	args := []string{"-c", s.command}
	bts, err := exec.Command(s.shell, args...).CombinedOutput()
	if err != nil {
		// Return the stdout result even on failure
		// TODO(mkcp): This is a good place to switch on exec.Command errors and provide better guidance.
		err1 := ShellExecError{
			command: s.command,
			err:     err,
		}
		return New(s, string(bts), Unknown, err1)
	}

	return New(s, string(bts), Success, nil)
}

type ShellExecError struct {
	command string
	err     error
}

func (e ShellExecError) Error() string {
	return fmt.Sprintf("exec error, command=%s, error=%s", e.command, e.err.Error())
}

func (e ShellExecError) Unwrap() error {
	return e.err
}
