package cmd

import (
	"github.com/spf13/cobra"
	"trojan/trojan"
)

// upgradeCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "更新（勿用）",
	Long:  "请勿更新，慎用",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		version := ""
		if len(args) == 1 {
			version = args[0]
		}
		trojan.InstallTrojan(version)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
