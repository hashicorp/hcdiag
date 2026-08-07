// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cli

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubCommand is a minimal Command for use in tests.
type stubCommand struct {
	runFunc func(args []string) int
}

func (s *stubCommand) Run(args []string) int {
	if s.runFunc != nil {
		return s.runFunc(args)
	}
	return 0
}
func (s *stubCommand) Help() string     { return "stub help" }
func (s *stubCommand) Synopsis() string { return "stub synopsis" }

func factory(cmd Command) CommandFactory {
	return func() (Command, error) { return cmd, nil }
}

func TestCLI_Run_Dispatch(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		commands map[string]CommandFactory
		wantCode int
	}{
		{
			name: "known command returns its exit code",
			args: []string{"run"},
			commands: map[string]CommandFactory{
				"run": factory(&stubCommand{runFunc: func([]string) int { return 0 }}),
			},
			wantCode: 0,
		},
		{
			name: "known command passes remaining args",
			args: []string{"run", "--flag"},
			commands: map[string]CommandFactory{
				"run": factory(&stubCommand{runFunc: func(args []string) int {
					if len(args) == 1 && args[0] == "--flag" {
						return 42
					}
					return 1
				}}),
			},
			wantCode: 42,
		},
		{
			name: "default command used when no args",
			args: []string{},
			commands: map[string]CommandFactory{
				"": factory(&stubCommand{runFunc: func([]string) int { return 5 }}),
			},
			wantCode: 5,
		},
		{
			name: "unknown command returns 127",
			args: []string{"notacommand"},
			commands: map[string]CommandFactory{
				"run": factory(&stubCommand{}),
			},
			wantCode: 127,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := &CLI{
				Args:     tc.args,
				Commands: tc.commands,
			}
			code, err := c.Run()
			require.NoError(t, err)
			assert.Equal(t, tc.wantCode, code)
		})
	}
}

func TestCLI_Run_Help(t *testing.T) {
	helpCalled := false
	c := &CLI{
		Args: []string{"--help"},
		Commands: map[string]CommandFactory{
			"run": factory(&stubCommand{}),
		},
		HelpFunc: func(cmds map[string]CommandFactory) string {
			helpCalled = true
			return "help text"
		},
	}
	code, err := c.Run()
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.True(t, helpCalled)
}

func TestCLI_Run_UnknownCommandPrintsHelp(t *testing.T) {
	helpCalled := false
	c := &CLI{
		Args: []string{"notacommand"},
		Commands: map[string]CommandFactory{
			"run": factory(&stubCommand{}),
		},
		HelpFunc: func(cmds map[string]CommandFactory) string {
			helpCalled = true
			return "help text"
		},
	}
	code, err := c.Run()
	require.NoError(t, err)
	assert.Equal(t, 127, code)
	assert.True(t, helpCalled, "HelpFunc should be called for unknown commands")
}

func TestCLI_Run_HiddenCommandsExcludedFromHelp(t *testing.T) {
	var helpCmds map[string]CommandFactory
	c := &CLI{
		Args: []string{"--help"},
		Commands: map[string]CommandFactory{
			"run": factory(&stubCommand{}),
			"":    factory(&stubCommand{}),
		},
		HiddenCommands: []string{""},
		HelpFunc: func(cmds map[string]CommandFactory) string {
			helpCmds = cmds
			return "help text"
		},
	}
	_, err := c.Run()
	require.NoError(t, err)
	_, hasEmpty := helpCmds[""]
	assert.False(t, hasEmpty, "hidden command '' should not appear in help")
	_, hasRun := helpCmds["run"]
	assert.True(t, hasRun, "visible command 'run' should appear in help")
}

func TestCLI_Run_FlagAsFirstArg(t *testing.T) {
	// A flag (starts with '-') as the first arg should fall through to the "" default command.
	defaultCalled := false
	c := &CLI{
		Args: []string{"--some-flag", "value"},
		Commands: map[string]CommandFactory{
			"": factory(&stubCommand{runFunc: func(args []string) int {
				defaultCalled = true
				return 0
			}}),
		},
	}
	code, err := c.Run()
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.True(t, defaultCalled, "default command should be invoked when first arg is a flag")
}

func TestCLI_Run_HelpFuncNil(t *testing.T) {
	// --help with no HelpFunc set should return 0 without panicking.
	c := &CLI{
		Args:     []string{"--help"},
		HelpFunc: nil,
		Commands: map[string]CommandFactory{
			"run": factory(&stubCommand{}),
		},
	}
	code, err := c.Run()
	require.NoError(t, err)
	assert.Equal(t, 0, code)
}

func TestCLI_Run_FactoryError(t *testing.T) {
	// When a CommandFactory returns an error, Run should propagate it and return exit code 1.
	factoryErr := fmt.Errorf("factory exploded")
	c := &CLI{
		Args: []string{"run"},
		Commands: map[string]CommandFactory{
			"run": func() (Command, error) { return nil, factoryErr },
		},
	}
	code, err := c.Run()
	assert.Equal(t, 1, code)
	assert.ErrorIs(t, err, factoryErr)
}

func TestCLI_IsVersion(t *testing.T) {
	testCases := []struct {
		name   string
		args   []string
		expect bool
	}{
		{
			name:   "--version flag",
			args:   []string{"--version"},
			expect: true,
		},
		{
			name:   "-version flag",
			args:   []string{"-version"},
			expect: true,
		},
		{
			name:   "no version flag",
			args:   []string{"run", "--flag"},
			expect: false,
		},
		{
			name:   "empty args",
			args:   []string{},
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := &CLI{Args: tc.args}
			assert.Equal(t, tc.expect, c.IsVersion())
		})
	}
}
