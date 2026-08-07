// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cli

import "fmt"

// Command is the interface that every CLI subcommand must implement.
type Command interface {
	Run(args []string) int
	Help() string
	Synopsis() string
}

// CommandFactory is a function that produces a Command.
type CommandFactory func() (Command, error)

// CLI routes Args to the correct Command.
type CLI struct {
	Name           string
	Args           []string
	Commands       map[string]CommandFactory
	HiddenCommands []string
	HelpFunc       func(map[string]CommandFactory) string
}

// IsVersion reports whether --version or -version appears in Args.
func (c *CLI) IsVersion() bool {
	for _, arg := range c.Args {
		if arg == "--version" || arg == "-version" {
			return true
		}
	}
	return false
}

// Run dispatches to the correct Command based on Args and returns its exit code.
// Returns 0 after printing help when --help or -help is present.
// Returns 127 for an unknown command name.
func (c *CLI) Run() (int, error) {
	// Check for --help before anything else.
	for _, arg := range c.Args {
		if arg == "--help" || arg == "-help" {
			if c.HelpFunc != nil {
				fmt.Print(c.HelpFunc(c.visibleCommands()))
			}
			return 0, nil
		}
	}

	// Determine the command name: first non-flag arg, or "" if none.
	name := ""
	remaining := c.Args
	if len(c.Args) > 0 && len(c.Args[0]) > 0 && c.Args[0][0] != '-' {
		name = c.Args[0]
		remaining = c.Args[1:]
	}

	factory, ok := c.Commands[name]
	if !ok {
		if c.HelpFunc != nil {
			fmt.Print(c.HelpFunc(c.visibleCommands()))
		}
		return 127, nil
	}

	cmd, err := factory()
	if err != nil {
		return 1, err
	}

	return cmd.Run(remaining), nil
}

// visibleCommands returns a copy of Commands with HiddenCommands keys removed.
func (c *CLI) visibleCommands() map[string]CommandFactory {
	hidden := make(map[string]struct{}, len(c.HiddenCommands))
	for _, h := range c.HiddenCommands {
		hidden[h] = struct{}{}
	}
	visible := make(map[string]CommandFactory, len(c.Commands))
	for k, v := range c.Commands {
		if _, isHidden := hidden[k]; !isHidden {
			visible[k] = v
		}
	}
	return visible
}
