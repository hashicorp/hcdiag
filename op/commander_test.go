package op

import (
	"reflect"
	"testing"

	"github.com/hashicorp/go-hclog"

	"github.com/stretchr/testify/assert"
)

func TestNewCommander(t *testing.T) {
	expect := Op{
		Identifier: "echo hello",
		Runner: Commander{
			Command: "echo hello",
			format:  "string",
		},
	}
	actual := NewCommander("echo hello", "string")
	// TODO: proper comparison, my IDE doesn't like this: "avoid using reflect.DeepEqual with errors"
	if !reflect.DeepEqual(&expect, actual) {
		t.Errorf("expected (%#v) does not match actual (%#v)", expect, actual)
	}
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
			c := NewCommander(tc.command, tc.format)
			result, status, err := c.Runner.Run()
			assert.NoError(t, err)
			assert.Equal(t, Success, status)
			assert.Equal(t, tc.expect, result)
		})
	}
}

func TestCommander_RunError(t *testing.T) {
	tt := []struct {
		desc    string
		command string
		format  string
		expect  interface{}
		status  Status
	}{
		{
			desc:    "errors and unknown when bash returns error",
			command: "cat no-file-to-see-here",
			format:  "string",
			expect:  "cat: no-file-to-see-here: No such file or directory\n",
			status:  Unknown,
		},
		{
			desc:    "errors and fails on bad json",
			command: `echo {"bad",}`,
			format:  "json",
			status:  Unknown,
		},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			c := NewCommander(tc.command, tc.format)
			result, status, err := c.Runner.Run()
			assert.Error(t, err)
			hclog.L().Trace("commander.Run() errored", "error", err, "error type", reflect.TypeOf(err))
			assert.Equal(t, tc.status, status)
			if tc.expect != nil {
				assert.Equal(t, tc.expect, result)
			}
		})
	}
}
