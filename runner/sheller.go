package runner

import (
	"fmt"
	"os/exec"

	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/hcdiag/op"

	"github.com/hashicorp/hcdiag/util"
)

// Sheller runs shell commands in a real unix shell.
type Sheller struct {
	Command    string           `json:"command"`
	Shell      string           `json:"shell"`
	Redactions []*redact.Redact `json:"redactions"`
}

// NewSheller provides a runner for arbitrary shell code.
func NewSheller(command string, redactions []*redact.Redact) *Sheller {
	return &Sheller{
		Command:    command,
		Redactions: redactions,
	}
}

func (s Sheller) ID() string {
	return s.Command
}

// Run ensures a shell exists and optimistically executes the given Command string
func (s Sheller) Run() []op.Op {
	opList := make([]op.Op, 0)

	// Read the shell from the environment
	shell, err := util.GetShell()
	if err != nil {
		return append(opList, op.New(s.ID(), nil, op.Fail, err, Params(s)))
	}
	s.Shell = shell

	// Run the command
	args := []string{"-c", s.Command}
	bts, cmdErr := exec.Command(s.Shell, args...).CombinedOutput()
	// Store and redact the result before cmd error handling, so we can return it in error and success cases.
	redBts, redErr := redact.Bytes(bts, s.Redactions)
	// Fail run if unable to redact
	if redErr != nil {
		return append(opList, op.New(s.ID(), nil, op.Fail, redErr, Params(s)))
	}
	if cmdErr != nil {
		return append(opList, op.New(s.ID(), string(redBts), op.Unknown,
			ShellExecError{
				command: s.Command,
				err:     cmdErr,
			}, Params(s)))
	}
	return append(opList, op.New(s.ID(), string(redBts), op.Success, nil, Params(s)))
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
