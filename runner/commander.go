package runner

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/util"
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

// Run executes the Command
func (c Commander) Run() []op.Op {
	opList := make([]op.Op, 0)
	bits := strings.Split(c.Command, " ")
	cmd := bits[0]
	args := bits[1:]

	// Exit early with a wrapped error if the command isn't found on this system
	_, err := util.HostCommandExists(cmd)
	if err != nil {
		return op.New(c.ID(), nil, op.Skip, err, Params(c))
	}

	// Execute command
	bts, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		err1 := CommandExecError{command: c.Command, format: c.Format, err: err}
		return append(opList, op.New(c.ID(), string(bts), op.Unknown, err1, Params(c)))
	}

	// Parse result format
	// TODO(mkcp): This can be detected rather than branching on user input
	switch {
	case c.Format == "string":
		redBts, err := redact.Bytes(bts, c.Redactions)
		if err != nil {
			return append(opList, op.New(c.ID(), nil, op.Fail, err, Params(c)))
		}
		redResult := strings.TrimSuffix(string(redBts), "\n")
		return append(opList, op.New(c.ID(), redResult, op.Success, nil, Params(c)))

	case c.Format == "json":
		var obj any
		marshErr := json.Unmarshal(bts, &obj)
		if marshErr != nil {
			// Redact the string to return the failed-to-parse JSON
			redBts, redErr := redact.Bytes(bts, c.Redactions)
			if redErr != nil {
				return append(opList, op.New(c.ID(), nil, op.Fail, redErr, Params(c)))
			}
			return append(opList, op.New(c.ID(), string(redBts), op.Unknown,
				UnmarshalError{
					command: c.Command,
					err:     marshErr,
				}, Params(c)))
		}
		redResult, redErr := redact.JSON(obj, c.Redactions)
		if redErr != nil {
			return append(opList, op.New(c.ID(), nil, op.Fail, redErr, Params(c)))
		}
		return append(opList, op.New(c.ID(), redResult, op.Success, nil, Params(c)))
	default:
		redBts, redErr := redact.Bytes(bts, c.Redactions)
		if redErr != nil {
			return append(opList, op.New(c.ID(), nil, op.Fail, redErr, Params(c)))
		}
		return append(opList, op.New(c.ID(), string(redBts), op.Fail,
			FormatUnknownError{
				command: c.Command,
				format:  c.Format,
			}, Params(c)))
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
