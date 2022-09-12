package runner

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/hcdiag/op"
)

var _ Runner = Commander{}

// Commander runs shell commands.
type Commander struct {
	Command    string           `json:"command"`
	Format     string           `json:"format"`
	Redactions []*redact.Redact `json:"redactions"`
}

// NewCommander provides a runner for bin commands
func NewCommander(command string, format string, redactions []*redact.Redact) *Commander {
	return &Commander{
		Command:    command,
		Format:     format,
		Redactions: redactions,
	}
}

func (c Commander) ID() string {
	return c.Command
}

// CommandExists is a cross-platform function that returns true if a command exists on the host
func CommandExists(command string) bool {
	var cmd string
	var args []string

	// Set appropriate lookup command based on OS
	if runtime.GOOS == "windows" {
		cmd = "where"
		args = []string{command}
	} else {
		// Should work on all POSIX-compliant systems
		cmd = "command"
		args = []string{"-v", command}
	}

	// No redactions because we never store or inspect the output of this command
	err := exec.Command(cmd, args...).Run()
	return err == nil
}

// Run executes the Command
func (c Commander) Run() op.Op {
	bits := strings.Split(c.Command, " ")
	cmd := bits[0]
	args := bits[1:]

	// Exit early if the command isn't found on this system
	if !CommandExists(cmd) {
		cmdErr := fmt.Sprintf("%s: command not found", cmd)
		return op.New(c.ID(), cmdErr, op.Skip, nil, Params(c))
	}

	// Execute command
	bts, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		err1 := CommandExecError{command: c.Command, format: c.Format, err: err}
		return op.New(c.ID(), string(bts), op.Unknown, err1, Params(c))
	}

	// Parse result format
	// TODO(mkcp): This can be detected rather than branching on user input
	switch {
	case c.Format == "string":
		redBts, err := redact.Bytes(bts, c.Redactions)
		if err != nil {
			return op.New(c.ID(), nil, op.Fail, err, Params(c))
		}
		redResult := strings.TrimSuffix(string(redBts), "\n")
		return op.New(c.ID(), redResult, op.Success, nil, Params(c))

	case c.Format == "json":
		var obj any
		marshErr := json.Unmarshal(bts, &obj)
		if marshErr != nil {
			// Redact the string to return the failed-to-parse JSON
			redBts, redErr := redact.Bytes(bts, c.Redactions)
			if redErr != nil {
				return op.New(c.ID(), nil, op.Fail, redErr, Params(c))
			}
			return op.New(c.ID(), string(redBts), op.Unknown,
				UnmarshalError{
					command: c.Command,
					err:     marshErr,
				}, Params(c))
		}
		redResult, redErr := redact.JSON(obj, c.Redactions)
		if redErr != nil {
			return op.New(c.ID(), nil, op.Fail, redErr, Params(c))
		}
		return op.New(c.ID(), redResult, op.Success, nil, Params(c))
	default:
		redBts, redErr := redact.Bytes(bts, c.Redactions)
		if redErr != nil {
			return op.New(c.ID(), nil, op.Fail, redErr, Params(c))
		}
		return op.New(c.ID(), string(redBts), op.Fail,
			FormatUnknownError{
				command: c.Command,
				format:  c.Format,
			}, Params(c))
	}
}

type CommandExecError struct {
	command string
	format  string
	err     error
}

func (e CommandExecError) Error() string {
	return fmt.Sprintf("exec error, command=%s, format=%s, error=%s", e.command, e.format, e.err.Error())
}

func (e CommandExecError) Unwrap() error {
	return e.err
}

type UnmarshalError struct {
	command string
	err     error
}

func (e UnmarshalError) Error() string {
	return fmt.Sprintf("json.Unmarshal(...) error, command=%s, error=%s", e.command, e.err)
}

func (e UnmarshalError) Unwrap() error {
	return e.err
}

type FormatUnknownError struct {
	command string
	format  string
}

func (e FormatUnknownError) Error() string {
	return fmt.Sprintf("unknown format: must be either 'string' or 'json', format=%s, command=%s", e.format, e.command)
}
