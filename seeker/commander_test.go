package seeker

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCommander(t *testing.T) {
	expect := Seeker{
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

func TestCommanderRunString(t *testing.T) {
	c := NewCommander("echo hello", "string")
	out, err := c.Run()

	if err != nil {
		t.Errorf("err should be nil, got: %s", err)
	}

	if out != "hello" {
		t.Errorf("out should be 'hello', got: '%s'", out)
	}
}

func TestCommanderRunJSON(t *testing.T) {
	expect := make(map[string]interface{})
	expect["hi"] = "there"

	c := NewCommander("echo {\"hi\":\"there\"}", "json")
	actual, err := c.Run()

	if err != nil {
		t.Errorf("err should be nil, got: %s", err)
	}
	if !reflect.DeepEqual(expect, actual) {
		t.Errorf("expected (%#v) does not match actual (%#v)", expect, actual)
	}
}

func TestCommanderRunError(t *testing.T) {
	c := NewCommander("cat no-file-to-see-here", "string")
	out, err := c.Run()

	// we should get the command's error output
	assert.Contains(t, out, "No such file")

	// and an error
	assert.Error(t, err)
}

func TestCommanderBadJSON(t *testing.T) {
	c := NewCommander(`echo {"bad",}`, "json")
	_, err := c.Run()
	assert.Errorf(t, err, "commander.Run() should error on bad json")
}
