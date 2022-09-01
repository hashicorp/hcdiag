package version

import (
	"github.com/hashicorp/hcdiag/version"
	"github.com/mitchellh/cli"
)

const helpText = `Usage: hcdiag version`
const synopsisText = `Print the current version of hcdiag`

var _ cli.Command = &cmd{}

type cmd struct {
	ui cli.Ui
}

func New(ui cli.Ui) *cmd {
	return &cmd{ui: ui}
}

// CommandFactory provides a cli.CommandFactory that will produce an appropriately-initiated *cmd.
func CommandFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		return New(ui), nil
	}
}

func (c cmd) Help() string {
	return helpText
}

func (c cmd) Run([]string) int {
	v := version.GetVersion()
	c.ui.Output(v.FullVersionNumber(true))

	return 0
}

func (c cmd) Synopsis() string {
	return synopsisText
}
