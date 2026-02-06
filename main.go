// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/hashicorp/hcdiag/command"
	"github.com/mitchellh/cli"
)

const appName = "hcdiag"

func main() {
	os.Exit(realMain())
}

func realMain() (returnCode int) {
	ui := &cli.BasicUi{
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	c := &cli.CLI{
		Name: appName,

		// Args should not include the command name, so we pass along [1:] instead of all args.
		Args: os.Args[1:],

		// Something more robust will be necessary if we add sub-subcommands, however the approach of having factories
		// within the various command packages seemed like a clean starting point.
		Commands: map[string]cli.CommandFactory{
			// The empty string key is what will happen when no subcommands are provided to hcdiag.
			"":        command.RunCommandFactory(ui),
			"run":     command.RunCommandFactory(ui),
			"version": command.VersionCommandFactory(ui),
		},

		HiddenCommands: []string{
			// Keep the output of available commands a bit cleaner by not including the default command in the help.
			"",
		},

		HelpFunc: mainHelp,
	}

	// This overrides the output format for the --version flag so that it matches the version subcommand
	if c.IsVersion() {
		return command.NewVersionCommand(ui).Run(c.Args)
	}

	rc, err := c.Run()
	if err != nil {
		ui.Warn(err.Error())

	}
	return rc
}

func mainHelp(commands map[string]cli.CommandFactory) string {
	// NOTE: This is a slightly modified copy of mitchellh/cli.BasicHelp so that we can include some general
	// application usage information in it.
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("Usage: %s [--version] [--help] <command> [<args>]\n\n", appName))

	appUsage := strings.TrimSpace(`
hcdiag simplifies debugging HashiCorp products by automating shared and product-specific diagnostics data collection
on individual nodes. Running the binary issues a set of operations that read the current state of the system then write
the results to a tar.gz bundle.

DEPRECATION NOTICE:
To maintain backward compatibility for a short time, you may execute a local diagnostic run using 'hcdiag [<args>]'.
However, this will be deprecated in an upcoming release. You should begin using the 'run' subcommand for such purposes.
For guidance on available options for running local execution, please refer to the help output from 'run' by using
'hcdiag run --help'.
`)
	buf.WriteString(fmt.Sprintf("%s\n\n", appUsage))

	buf.WriteString("Available commands are:\n")

	// Get the list of keys so we can sort them, and also get the maximum
	// key length so they can be aligned properly.
	keys := make([]string, 0, len(commands))
	maxKeyLen := 0
	for key := range commands {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}

		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		commandFunc, ok := commands[key]
		if !ok {
			// This should never happen since we JUST built the list of
			// keys.
			panic("command not found: " + key)
		}

		c, err := commandFunc()
		if err != nil {
			log.Printf("[ERR] cli: Command '%s' failed to load: %s",
				key, err)
			continue
		}

		key = fmt.Sprintf("%s%s", key, strings.Repeat(" ", maxKeyLen-len(key)))
		buf.WriteString(fmt.Sprintf("    %s    %s\n", key, c.Synopsis()))
	}

	return buf.String()
}
