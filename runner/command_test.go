// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package runner

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/hcdiag/op"

	"github.com/hashicorp/go-hclog"

	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	tt := []struct {
		desc      string
		cfg       CommandConfig
		expect    *Command
		expectErr bool
	}{
		{
			desc:      "empty config causes an error",
			cfg:       CommandConfig{},
			expectErr: true,
		},
		{
			desc: "empty format defaults to string",
			cfg: CommandConfig{
				Command: "bogus-command",
			},
			expect: &Command{
				Command: "bogus-command",
				Format:  "string",
				ctx:     context.Background(),
			},
		},
		{
			desc: "invalid format causes an error",
			cfg: CommandConfig{
				Command: "bogus-command",
				Format:  "invalid",
			},
			expectErr: true,
		},
		{
			desc: "negative timeout duration causes an error",
			cfg: CommandConfig{
				Command: "bogus-command",
				Timeout: -10 * time.Second,
			},
			expectErr: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := NewCommand(tc.cfg)
			if tc.expectErr {
				assert.ErrorAs(t, err, &CommandConfigError{})
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expect, c)
			}
		})
	}
}

func TestCommand_Run(t *testing.T) {
	tt := []struct {
		desc    string
		command string
		format  string
		expect  interface{}
	}{
		{
			desc:    "can run with string format",
			command: "echo hello",
			format:  "string",
			expect:  map[string]any{"text": "hello"},
		},
		{
			desc:    "can run with json format",
			command: "echo '{\"hi\":\"there\"}'",
			format:  "json",
			expect: func() interface{} {
				expect := make(map[string]any)
				expect["json"] = map[string]any{"hi": "there"}
				return expect
			}(),
		},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			cfg := CommandConfig{
				Command: tc.command,
				Format:  tc.format,
			}
			c, err := NewCommand(cfg)
			assert.NoError(t, err)
			o := c.Run()
			assert.NoError(t, o.Error)
			assert.Equal(t, op.Success, o.Status)
			assert.Equal(t, tc.expect, o.Result)
		})
	}
}

func TestCommand_RunError(t *testing.T) {
	tt := []struct {
		desc    string
		command string
		format  string
		expect  interface{}
		status  op.Status
	}{
		{
			desc:    "errors and unknown when bash returns error",
			command: "cat no-file-to-see-here",
			format:  "string",
			expect:  map[string]any{"text": "cat: no-file-to-see-here: No such file or directory\n"},
			status:  op.Unknown,
		},
		{
			desc:    "errors and fails on bad json",
			command: `echo '{"bad",}'`,
			format:  "json",
			expect:  map[string]any{"json": "{\"bad\",}\n"},
			status:  op.Unknown,
		},
		{
			desc:    "returns a Skip status when a nonexistent command is called",
			command: "fooblarbalurg this is not a real command",
			format:  "string",
			expect:  nil,
			status:  op.Skip,
		},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			cfg := CommandConfig{
				Command: tc.command,
				Format:  tc.format,
			}
			c, err := NewCommand(cfg)
			assert.NoError(t, err)
			o := c.Run()
			assert.Error(t, o.Error)
			hclog.L().Trace("command.Run() errored", "error", o.Error, "error type", reflect.TypeOf(o.Error))
			assert.Equal(t, tc.status, o.Status)
			if tc.expect != nil {
				assert.Equal(t, tc.expect, o.Result)
			}
		})
	}
}

func TestCommand_RunCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()

	cmd := Command{
		Command: "bogus-command",
		ctx:     ctx,
	}

	result := cmd.Run()
	assert.Equal(t, op.Canceled, result.Status)
	assert.ErrorIs(t, result.Error, context.Canceled)
}

func TestCommand_RunTimeout(t *testing.T) {
	t.Parallel()

	// Set to a short timeout, and sleep briefly to ensure it passes before we try to run the command
	ctx, cancelFunc := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancelFunc()
	time.Sleep(1 * time.Nanosecond)

	cmd := Command{
		Command: "bogus-command",
		ctx:     ctx,
	}

	result := cmd.Run()
	assert.Equal(t, op.Timeout, result.Status)
	assert.ErrorIs(t, result.Error, context.DeadlineExceeded)
}

func Test_parseCommand(t *testing.T) {
	tt := []struct {
		desc    string
		command string
		expect  parsedCommand
	}{
		{
			desc:    "Test regular string arguments",
			command: "cmd --arg1 value1 --arg2 value2",
			expect: parsedCommand{
				cmd: "cmd",
				args: []string{
					"--arg1",
					"value1",
					"--arg2",
					"value2",
				},
				err: nil,
			},
		},
		{
			desc:    "Test a command with no args",
			command: "cmd",
			expect: parsedCommand{
				cmd:  "cmd",
				args: []string{},
			},
		},
		{
			desc:    "Test JSON input with spaces",
			command: "jq -n '$in.\"foo bar\"' --argjson in '{\"foo bar\": 22}'",
			expect: parsedCommand{
				cmd: "jq",
				args: []string{
					"-n",
					"$in.\"foo bar\"",
					"--argjson",
					"in",
					"{\"foo bar\": 22}",
				},
			},
		},
		{
			desc:    "Test JSON style args and invalid JSON",
			command: "echo '{\"bad\",}'",
			expect: parsedCommand{
				cmd: "echo",
				args: []string{
					"{\"bad\",}",
				},
			},
		},
		{
			desc:    "Test Backticks Produce Error",
			command: "ls `pwd`",
			expect: parsedCommand{
				err: &CommandParseError{},
			},
		},
		{
			desc:    "Test Pipes Produce Error",
			command: "cmd1 arg1 | cmd2",
			expect: parsedCommand{
				err: &CommandParseError{},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			p, err := parseCommand(tc.command)
			if tc.expect.err != nil {
				assert.Error(t, p.err)
				assert.Error(t, err)
				// We verify that we can decode the error as the one we expect
				assert.ErrorAs(t, p.err, tc.expect.err)
				assert.ErrorAs(t, err, tc.expect.err)
			} else {
				assert.NoError(t, p.err)
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expect.cmd, p.cmd)
			assert.Equal(t, tc.expect.args, p.args)
		})
	}
}
