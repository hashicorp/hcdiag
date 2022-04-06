package seeker

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Commander runs shell commands.
type Commander struct {
	Command string `json:"command"`
	format  string
}

// NewCommander provides a Seeker for running shell commands.
func NewCommander(command string, format string) *Seeker {
	return &Seeker{
		Identifier: command,
		Runner:     Commander{Command: command, format: format},
	}
}

// Run executes the Command
func (c Commander) Run() (interface{}, Status, error) {
	var result interface{}
	var err error

	bits := strings.Split(c.Command, " ")
	cmd := bits[0]
	args := bits[1:]

	// TODO(mkcp): Add cross-platform commandExists() func to ensure there's a bin we can call

	// Execute command
	bts, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		return string(bts), Unknown, fmt.Errorf("exec.Command error: %s", err)
	}

	// Parse result
	// TODO(mkcp): This can be detected rather than branching on user input
	switch {
	case c.format == "string":
		result = strings.TrimSuffix(string(bts), "\n")

	case c.format == "json":
		if e := json.Unmarshal(bts, &result); e != nil {
			return nil, Fail, fmt.Errorf("json.Unmarshal error: %s", e)
		}

	default:
		err = errors.New("command output format must be either 'string' or 'json'")
		return result, Fail, err
	}

	return result, Success, nil
}
