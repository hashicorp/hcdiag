package command

import (
	"github.com/hashicorp/hcdiag/version"
	"github.com/mitchellh/cli"
)

var _ cli.Command = &VersionCommand{}

type VersionCommand struct {
	ui cli.Ui
}

func NewVersionCommand(ui cli.Ui) *VersionCommand {
	return &VersionCommand{ui: ui}
}

// VersionCommandFactory provides a cli.CommandFactory that will produce an appropriately-initiated *command.
func VersionCommandFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		return NewVersionCommand(ui), nil
	}
}

func (c VersionCommand) Help() string {
	return "Usage: hcdiag version"
}

func (c VersionCommand) Run([]string) int {
	v := version.GetVersion()
	c.ui.Output(v.FullVersionNumber(true))

	return Success
}

func (c VersionCommand) Synopsis() string {
	return "Print the current version of hcdiag"
}
