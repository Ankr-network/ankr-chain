package commands

import (
	"github.com/spf13/cobra"
)

func AddHistoryStorageNodeFlags(cmd *cobra.Command, hsDBType string, hsDBHost string, hsDBName string) {
	cmd.Flags().String("historydb.type", hsDBType, "Type of history DB")
	cmd.Flags().String("historydb.host", hsDBHost, "Host of history DB")
	cmd.Flags().String("historydb.name", hsDBName, "Name of history DB")
}
