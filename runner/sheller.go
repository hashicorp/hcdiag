package runner

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"

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
func (s Sheller) Run() op.Op {
	// Read the shell from the environment
	shell, err := util.GetShell()
	if err != nil {
		return op.New(s.ID(), nil, op.Fail, err, Params(s))
	}
	s.Shell = shell

	// Run the command
	args := []string{"-c", s.Command}
	bts, err := exec.Command(s.Shell, args...).CombinedOutput()
	if err != nil {
		// TODO(dcohen) redact error
		// Return the stdout result even on failure
		// TODO(mkcp): This is a good place to switch on exec.Command errors and provide better guidance.
		err1 := ShellExecError{
			command: s.Command,
			err:     err,
		}

		// Let's Redact (maybe wrap this in a function? We're going to be repeating it a *lot*)
		r := bytes.NewReader(bts)
		w := bytes.NewBuffer(make([]byte, 0))
		err = redact.ApplyMany(s.Redactions, w, r)
		// TODO(dcohen) make this error useful
		if err != nil {
			err1 = ShellExecError{
				command: s.Command,
				err:     err,
			}
		}
		return op.New(s.ID(), string(bts), op.Unknown, err1, Params(s))
	}

	return op.New(s.ID(), string(bts), op.Success, nil, Params(s))
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
