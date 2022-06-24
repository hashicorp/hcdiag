package op

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/hcdiag/util"
)

var _ Runner = Commander{}

// Commander runs shell commands.
type Commander struct {
	command string
	format  string
}

// NewCommander provides a Op for running shell commands.
func NewCommander(command string, format string) *Commander {
	return &Commander{
		command: command,
		format:  format,
	}
}

func (c Commander) ID() string {
	return c.command
}

func (c Commander) params() map[string]string {
	return map[string]string{
		"command": c.command,
		"format":  c.format,
	}
}

func (c Commander) op(result interface{}, status Status, err error) Op {
	return Op{
		Identifier: c.ID(),
		Result:     result,
		Error:      err,
		Status:     status,
		Params:     util.RunnerParams(c),
	}
}

// Run executes the Command
func (c Commander) Run() Op {
	var result interface{}

	bits := strings.Split(c.command, " ")
	cmd := bits[0]
	args := bits[1:]

	// TODO(mkcp): Add cross-platform commandExists() func to ensure there's a bin we can call

	// Execute command
	bts, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		err1 := CommandExecError{command: c.command, format: c.format, err: err}
		return c.op(string(bts), Unknown, err1)
	}

	// Parse result
	// TODO(mkcp): This can be detected rather than branching on user input
	switch {
	case c.format == "string":
		result = strings.TrimSuffix(string(bts), "\n")

	case c.format == "json":
		if err := json.Unmarshal(bts, &result); err != nil {
			// Return the command's response even if we can't parse it as json
			return c.op(string(bts), Unknown, UnmarshalError{command: c.command, err: err})
		}

	default:
		return c.op(result, Fail, FormatUnknownError{command: c.command, format: c.format})
	}

	return c.op(result, Success, nil)
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
