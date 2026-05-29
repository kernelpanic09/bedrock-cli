package cmd

import "github.com/spf13/cobra"

var costCmd = &cobra.Command{
	Use:   "cost",
	Short: "View cost and usage history",
	Long:  `Show what you've spent across Bedrock invocations tracked by bedrock-cli.`,
}

func init() {
	costCmd.AddCommand(costSummaryCmd)
	costCmd.AddCommand(costByModelCmd)
	costCmd.AddCommand(costByProjectCmd)
	costCmd.AddCommand(costByAccountCmd)
}
