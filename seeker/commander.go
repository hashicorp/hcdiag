package seeker

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// NewCommander provides a Seeker for running shell commands.
func NewCommander(command string, format string, mustSucceed bool) *Seeker {
	return &Seeker{
		Identifier:  command,
		Runner:      Commander{Command: command, format: format},
		MustSucceed: mustSucceed,
	}
}

// Commander runs shell commands.
type Commander struct {
	Command string `json:"command"`
	format  string
}

func (c Commander) Run() (result interface{}, err error) {
	bits := strings.Split(c.Command, " ")
	cmd := bits[0]
	args := bits[1:]

	bts, err := exec.Command(cmd, args...).CombinedOutput()

	if err != nil {
		return nil, fmt.Errorf("exec.Command error: %s", err)
	}

	switch {
	case c.format == "string":
		result = strings.TrimSuffix(string(bts), "\n")

	case c.format == "json":
		if err := json.Unmarshal(bts, &result); err != nil {
			return nil, fmt.Errorf("json.Unmarshal error: %s", err)
		}

	default:
		return nil, errors.New("command output format must be either 'string' or 'json'")
	}

	return result, err
}
