package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "synker",
	Short: "tools for sync differents harbor repository's members",
	Long:  "tools for sync differents harbor repository's members",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}
