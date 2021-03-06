package cmd

import (
	"fmt"

	"github.com/Ankr-network/ankr-chain/version"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the cli version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.CliVersion)
	},
}
