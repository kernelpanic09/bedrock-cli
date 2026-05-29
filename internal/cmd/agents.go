package cmd

import "github.com/spf13/cobra"

var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Manage and invoke Bedrock Agents",
	Long: `List, inspect, and invoke Bedrock Agents.

Agents combine foundation models with action groups and knowledge bases to
complete multi-step tasks. Sessions persist between invocations.`,
}

func init() {
	agentsCmd.AddCommand(agentsListCmd)
	agentsCmd.AddCommand(agentsDescribeCmd)
	agentsCmd.AddCommand(agentsInvokeCmd)
	agentsCmd.AddCommand(agentsSessionsCmd)
}
