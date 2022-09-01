// Package returns includes a set of semantic return codes that commands within hcdiag may use to indicate their
// success or failure. Particularly in the failure case, it intends to help avoid the proliferation of magic numbers
// while also allowing different subcommands to behave consistently.
//
// Groups of errors are numbered using an offset of iota, so that new errors can be added more easily, with lower
// risk of accidental duplication of return codes.
package returns

// Success indicates a successful command execution.
const Success int = 0

// The following error group is intended for issues within the initial setup process of a command's execution.
const (
	// FlagParseError indicates that a command was unable to successfully parse the flags/arguments provided to it.
	FlagParseError int = iota + 16
)

// The following error group is intended for issues with the Agent.
const (
	// AgentSetupError is returned when the agent cannot be setup properly prior to command execution.
	AgentSetupError int = iota + 32

	// AgentExecutionError is returned when the agent returns an error to the calling command.
	AgentExecutionError
)
