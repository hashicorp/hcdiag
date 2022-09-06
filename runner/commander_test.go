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
			expect:  "hello",
		},
		{
			desc:    "can run with json format",
			command: "echo {\"hi\":\"there\"}",
			format:  "json",
			expect: func() interface{} {
				expect := make(map[string]interface{})
				expect["hi"] = "there"
				return expect
			}(),
		},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			c := NewCommander(tc.command, tc.format, nil)
			o := c.Run()
			assert.NoError(t, o[0].Error)
			assert.Equal(t, op.Success, o[0].Status)
			assert.Equal(t, tc.expect, o[0].Result)
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
			expect:  "cat: no-file-to-see-here: No such file or directory\n",
			status:  op.Unknown,
		},
		{
			desc:    "errors and fails on bad json",
			command: `echo {"bad",}`,
			format:  "json",
			expect:  string("{\"bad\",}\n"),
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
			assert.Error(t, o[0].Error)
			hclog.L().Trace("commander.Run() errored", "error", o[0].Error, "error type", reflect.TypeOf(o[0].Error))
			assert.Equal(t, tc.status, o[0].Status)
			if tc.expect != nil {
				assert.Equal(t, tc.expect, o[0].Result)
			}
		})
	}
}
