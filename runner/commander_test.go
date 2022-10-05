package runner

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcdiag/op"

	"github.com/hashicorp/go-hclog"

	"github.com/stretchr/testify/assert"
)

func TestNewCommander(t *testing.T) {
	testCmd := "echo hello"
	testFmt := "string"
	expect := &Commander{
		Command: testCmd,
		Format:  testFmt,
	}
	actual := NewCommander(testCmd, testFmt, nil)
	assert.Equal(t, expect, actual)
}

func TestCommander_Run(t *testing.T) {
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
			c := NewCommander(tc.command, tc.format, nil)
			o := c.Run()
			assert.NoError(t, o.Error)
			assert.Equal(t, op.Success, o.Status)
			assert.Equal(t, tc.expect, o.Result)
		})
	}
}

func TestCommander_RunError(t *testing.T) {
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
			c := NewCommander(tc.command, tc.format, nil)
			o := c.Run()
			assert.Error(t, o.Error)
			hclog.L().Trace("commander.Run() errored", "error", o.Error, "error type", reflect.TypeOf(o.Error))
			assert.Equal(t, tc.status, o.Status)
			if tc.expect != nil {
				assert.Equal(t, tc.expect, o.Result)
			}
		})
	}
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
