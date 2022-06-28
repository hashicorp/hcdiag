package runner

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/hcdiag/op"
)

var _ Runner = Commander{}

// Commander runs shell commands.
type Commander struct {
	Command string `json:"command"`
	Format  string `json:"format"`
}

// NewCommander provides a runner for bin commands
func NewCommander(command string, format string) *Commander {
	return &Commander{
		Command: command,
		Format:  format,
	}
}

func (c Commander) ID() string {
	return c.Command
}

// Run executes the Command
func (c Commander) Run() op.Op {
	var result interface{}

	bits := strings.Split(c.Command, " ")
	cmd := bits[0]
	args := bits[1:]

	// TODO(mkcp): Add cross-platform commandExists() func to ensure there's a bin we can call

	// Execute command
	bts, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		err1 := CommandExecError{command: c.Command, format: c.Format, err: err}
		return op.New(c.ID(), string(bts), op.Unknown, err1, Params(c))
	}

	// Parse result
	// TODO(mkcp): This can be detected rather than branching on user input
	switch {
	case c.Format == "string":
		result = strings.TrimSuffix(string(bts), "\n")

	case c.Format == "json":
		if err := json.Unmarshal(bts, &result); err != nil {
			// Return the command's response even if we can't parse it as json
			return op.New(c.ID(), string(bts), op.Unknown,
				UnmarshalError{
					command: c.Command,
					err:     err,
				},
				Params(c))
		}

	default:
		return op.New(c.ID(), result, op.Fail, FormatUnknownError{
			command: c.Command,
			format:  c.Format,
		},
			Params(c))
	}

	return op.New(c.ID(), result, op.Success, nil, Params(c))
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
