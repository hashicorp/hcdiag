package runner

import (
	"context"
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
	// Parameters that are not shared/common
	Command string `json:"command"`
	Format  string `json:"format"`

	// Parameters that are common across runner types
	ctx context.Context

	Timeout    Timeout          `json:"timeout"`
	Redactions []*redact.Redact `json:"redactions"`
}

// CommandConfig is the configuration object passed into NewCommand or NewCommandWithContext. It includes
// the fields that those constructors will use to configure the Command object that they return.
type CommandConfig struct {
	// Command is the command line string that this runner should execute.
	Command string

	// Format is the desired output format. Valid options are "string" or "json"; the default will be "string" when
	// creating an object from the constructor functions NewCommand and NewCommandWithContext.
	Format string

	// Timeout specifies the amount of time that the runner should be allowed to execute before cancellation.
	Timeout time.Duration

	// Redactions includes any redactions to apply to the output of the runner.
	Redactions []*redact.Redact
}

// NewCommand provides a runner for bin commands.
func NewCommand(cfg CommandConfig) (*Command, error) {
	return NewCommandWithContext(context.Background(), cfg)
}

// NewCommandWithContext provides a runner for bin commands that includes a context.
func NewCommandWithContext(ctx context.Context, cfg CommandConfig) (*Command, error) {
	cmd := cfg.Command
	if cmd == "" {
		return nil, CommandConfigError{
			config: cfg,
			err:    fmt.Errorf("command must not be empty when creating a Command runner"),
		}
	}

	var format string
	switch f := strings.ToLower(cfg.Format); f {
	// default to "string" if not set
	case "":
		format = "string"
	// check for valid format types
	case "string", "json":
		format = f
	// error on invalid format
	default:
		return nil, CommandConfigError{
			config: cfg,
			err:    fmt.Errorf("format must be either 'string' or 'json', but got '%s'", f),
		}
	}

	timeout := cfg.Timeout
	if timeout < 0 {
		return nil, CommandConfigError{
			config: cfg,
			err:    fmt.Errorf("timeout must be a nonnegative value, but got '%s'", timeout.String()),
		}
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return &Command{
		ctx:        ctx,
		Command:    cmd,
		Format:     format,
		Timeout:    Timeout(timeout),
		Redactions: cfg.Redactions,
	}, nil
}

func (c Command) ID() string {
	return c.Command
}

// Run executes the Command
func (c Command) Run() op.Op {
	// protect from accidental nil reference panics
	if c.ctx == nil {
		c.ctx = context.Background()
	}

	runCtx := c.ctx
	var runCancelFunc context.CancelFunc
	if c.Timeout > 0 {
		runCtx, runCancelFunc = context.WithTimeout(c.ctx, time.Duration(c.Timeout))
		defer runCancelFunc()
	}

	startTime := time.Now()

	resultsChannel := make(chan op.Op, 1)
	go func(results chan<- op.Op) {
		p, err := parseCommand(c.Command)
		if err != nil {
			results <- op.New(c.ID(), nil, op.Fail, err, Params(c), startTime, time.Now())
		}

		// Exit early with a wrapped error if the command isn't found on this system
		_, err = util.HostCommandExists(p.cmd)
		if err != nil {
			results <- op.New(c.ID(), nil, op.Skip, err, Params(c), startTime, time.Now())
		}

		// Execute command
		bts, err := exec.Command(p.cmd, p.args...).CombinedOutput()
		if err != nil {
			err1 := CommandExecError{command: c.Command, format: c.Format, err: err}
			result := map[string]any{"text": string(bts)}
			results <- op.New(c.ID(), result, op.Unknown, err1, Params(c), startTime, time.Now())
		}

		// Parse result format
		// TODO(mkcp): This can be detected rather than branching on user input
		switch {
		case c.Format == "string":
			redBts, err := redact.Bytes(bts, c.Redactions)
			if err != nil {
				results <- op.New(c.ID(), nil, op.Fail, err, Params(c), startTime, time.Now())
			}
			redResult := strings.TrimSuffix(string(redBts), "\n")

			result := map[string]any{"text": redResult}
			results <- op.New(c.ID(), result, op.Success, nil, Params(c), startTime, time.Now())

		case c.Format == "json":
			var obj any
			marshErr := json.Unmarshal(bts, &obj)
			if marshErr != nil {
				// Redact the string to return the failed-to-parse JSON
				redBts, redErr := redact.Bytes(bts, c.Redactions)
				if redErr != nil {
					results <- op.New(c.ID(), nil, op.Fail, redErr, Params(c), startTime, time.Now())
				}
				result := map[string]any{"json": string(redBts)}
				results <- op.New(c.ID(), result, op.Unknown,
					UnmarshalError{
						command: c.Command,
						err:     marshErr,
					}, Params(c), startTime, time.Now())
			}
			redResult, redErr := redact.JSON(obj, c.Redactions)
			if redErr != nil {
				results <- op.New(c.ID(), nil, op.Fail, redErr, Params(c), startTime, time.Now())
			}
			result := map[string]any{"json": redResult}
			results <- op.New(c.ID(), result, op.Success, nil, Params(c), startTime, time.Now())
		default:
			redBts, redErr := redact.Bytes(bts, c.Redactions)
			if redErr != nil {
				results <- op.New(c.ID(), nil, op.Fail, redErr, Params(c), startTime, time.Now())
			}
			result := map[string]any{"out": string(redBts)}
			results <- op.New(c.ID(), result, op.Fail,
				FormatUnknownError{
					command: c.Command,
					format:  c.Format,
				}, Params(c), startTime, time.Now())
		}
	}(resultsChannel)

	select {
	case <-runCtx.Done():
		switch runCtx.Err() {
		case context.Canceled:
			return op.New(c.ID(), nil, op.Canceled, c.ctx.Err(), Params(c), startTime, time.Now())
		case context.DeadlineExceeded:
			return op.New(c.ID(), nil, op.Timeout, c.ctx.Err(), Params(c), startTime, time.Now())
		default:
			return op.New(c.ID(), nil, op.Unknown, c.ctx.Err(), Params(c), startTime, time.Now())
		}
	case result := <-resultsChannel:
		return result
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

	// We only support a single command, without piping from one to the next, in Command
	if len(p) > 1 {
		e := CommandParseError{
			command: command,
			err:     fmt.Errorf("piped commands are unsupported, please use a Shell runner or multiple Command runners, command=%s", command),
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
	return fmt.Sprintf("error parsing command in Command runner, command=%s, error=%s", e.command, e.err.Error())
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

var _ error = CommandConfigError{}

type CommandConfigError struct {
	config CommandConfig
	err    error
}

func (e CommandConfigError) Error() string {
	message := "invalid Command Config"
	if e.err != nil {
		return fmt.Sprintf("%s: %s", message, e.err.Error())
	}
	return message
}

func (e CommandConfigError) Unwrap() error {
	return e.err
}
