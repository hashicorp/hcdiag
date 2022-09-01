package command

// Success indicates a successful command execution.
const Success int = 0

// The following error group is intended for issues within the initial setup process of a command's execution.
const (
	// FlagParseError indicates that a command was unable to successfully parse the flags/arguments provided to it.
	FlagParseError int = iota + 16

	// ConfigError indicates that there was an error in the hcdiag configuration.
	ConfigError
)

// The following error group is intended for issues with the Agent.
const (
	// AgentSetupError is returned when the agent cannot be setup properly prior to command execution.
	AgentSetupError int = iota + 32

	// AgentExecutionError is returned when the agent returns an error to the calling command.
	AgentExecutionError
)
