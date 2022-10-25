package runner

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/cosiner/argv"
	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"
	"github.com/hashicorp/hcdiag/util"
)

var _ Runner = Command{}

// Command runs shell commands.
type Command struct {
	Command    string           `json:"command"`
	Format     string           `json:"format"`
	Redactions []*redact.Redact `json:"redactions"`
}

// NewCommand provides a runner for bin commands
func NewCommand(command string, format string, redactions []*redact.Redact) *Command {
	return &Command{
		Command:    command,
		Format:     format,
		Redactions: redactions,
	}
}

func (c Command) ID() string {
	return c.Command
}

// Run executes the Command
func (c Command) Run() op.Op {
	startTime := time.Now()

	p, err := parseCommand(c.Command)
	if err != nil {
		return op.New(c.ID(), nil, op.Fail, err, Params(c), startTime, time.Now())
	}

	// Exit early with a wrapped error if the command isn't found on this system
	_, err = util.HostCommandExists(p.cmd)
	if err != nil {
		return op.New(c.ID(), nil, op.Skip, err, Params(c), startTime, time.Now())
	}

	// Execute command
	bts, err := exec.Command(p.cmd, p.args...).CombinedOutput()
	if err != nil {
		err1 := CommandExecError{command: c.Command, format: c.Format, err: err}
		result := map[string]any{"text": string(bts)}
		return op.New(c.ID(), result, op.Unknown, err1, Params(c), startTime, time.Now())
	}

	// Parse result format
	// TODO(mkcp): This can be detected rather than branching on user input
	switch {
	case c.Format == "string":
		redBts, err := redact.Bytes(bts, c.Redactions)
		if err != nil {
			return op.New(c.ID(), nil, op.Fail, err, Params(c), startTime, time.Now())
		}
		redResult := strings.TrimSuffix(string(redBts), "\n")

		result := map[string]any{"text": redResult}
		return op.New(c.ID(), result, op.Success, nil, Params(c), startTime, time.Now())

	case c.Format == "json":
		var obj any
		marshErr := json.Unmarshal(bts, &obj)
		if marshErr != nil {
			// Redact the string to return the failed-to-parse JSON
			redBts, redErr := redact.Bytes(bts, c.Redactions)
			if redErr != nil {
				return op.New(c.ID(), nil, op.Fail, redErr, Params(c), startTime, time.Now())
			}
			result := map[string]any{"json": string(redBts)}
			return op.New(c.ID(), result, op.Unknown,
				UnmarshalError{
					command: c.Command,
					err:     marshErr,
				}, Params(c), startTime, time.Now())
		}
		redResult, redErr := redact.JSON(obj, c.Redactions)
		if redErr != nil {
			return op.New(c.ID(), nil, op.Fail, redErr, Params(c), startTime, time.Now())
		}
		result := map[string]any{"json": redResult}
		return op.New(c.ID(), result, op.Success, nil, Params(c), startTime, time.Now())
	default:
		redBts, redErr := redact.Bytes(bts, c.Redactions)
		if redErr != nil {
			return op.New(c.ID(), nil, op.Fail, redErr, Params(c), startTime, time.Now())
		}
		result := map[string]any{"out": string(redBts)}
		return op.New(c.ID(), result, op.Fail,
			FormatUnknownError{
				command: c.Command,
				format:  c.Format,
			}, Params(c), startTime, time.Now())
	}
}

type parsedCommand struct {
	cmd  string
	args []string
	err  error
}

func parseCommand(command string) (parsedCommand, error) {
	parsed := parsedCommand{}

	// Under the hood, the arguments provided to Windows are re-joined into a single string, allowing the Windows
	// OS to handle splitting however it needs to (see syscall/exec_windows.go, along with doc comments for exec.Command).
	// Allowing Windows paths through to argv can cause issues with character escaping; rather than adding complexity
	// to command parsing, just to have the args rejoined later on Windows, we short-circuit here, pulling out the cmd
	// name and arguments using a simple split on spaces.
	if runtime.GOOS == "windows" {
		split := strings.Split(command, " ")
		parsed.cmd = split[0]
		parsed.args = split[1:]
		return parsed, nil
	}

	// Argv returns a [][]string, where each outer slice represents commands split by '|' and the inner slices
	// have the command at element 0 and any arguments to the command in the remaining elements.
	p, err := argv.Argv(command, nil, nil)
	if err != nil {
		e := CommandParseError{
			command: command,
			err:     err,
		}
		parsed.err = e
		return parsed, e
	}

	// We only support a single command, without piping from one to the next, in Commander
	if len(p) > 1 {
		e := CommandParseError{
			command: command,
			err:     fmt.Errorf("piped commands are unsupported, please use a Sheller runner or multiple Commander runners, command=%s", command),
		}
		parsed.err = e
		return parsed, e
	}

	parsed.cmd = p[0][0]
	parsed.args = p[0][1:]

	return parsed, nil
}

var _ error = CommandParseError{}

type CommandParseError struct {
	command string
	err     error
}

func (e CommandParseError) Error() string {
	return fmt.Sprintf("error parsing command in Commander runner, command=%s, error=%s", e.command, e.err.Error())
}

func (e CommandParseError) Unwrap() error {
	return e.err
}

var _ error = CommandExecError{}

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

var _ error = UnmarshalError{}

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

var _ error = FormatUnknownError{}

type FormatUnknownError struct {
	command string
	format  string
}

func (e FormatUnknownError) Error() string {
	return fmt.Sprintf("unknown format: must be either 'string' or 'json', format=%s, command=%s", e.format, e.command)
}
