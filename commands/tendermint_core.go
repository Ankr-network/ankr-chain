package commands

import (
	"github.com/spf13/cobra"
	tmcorecommands "github.com/tendermint/tendermint/cmd/tendermint/commands"
)

func AddTendermintCoreCommands(cmd *cobra.Command) {
	cmd.AddCommand(
		tmcorecommands.ResetAllCmd,
		tmcorecommands.VersionCmd)
}


