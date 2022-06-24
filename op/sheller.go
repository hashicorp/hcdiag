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

func (s Sheller) params() map[string]string {
	return map[string]string{
		"command": s.command,
		"shell":   s.shell,
	}
}

// Run ensures a shell exists and optimistically executes the given Command string
func (s Sheller) Run() Op {
	// Read the shell from the environment
	shell, err := util.GetShell()
	if err != nil {
		return Op{
			Identifier: s.ID(),
			Error:      err,
			ErrString:  err.Error(),
			Status:     Fail,
			Params:     util.RunnerParams(s),
		}
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
		return Op{
			Identifier: s.id,
			Result:     string(bts),
			Error:      err1,
			ErrString:  err1.Error(),
			Status:     Unknown,
			Params:     util.RunnerParams(s),
		}
	}

	return Op{
		Identifier: s.id,
		Result:     string(bts),
		Status:     Success,
		Params:     util.RunnerParams(s),
	}
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
