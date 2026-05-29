package cmd

import "github.com/spf13/cobra"

var guardrailsCmd = &cobra.Command{
	Use:   "guardrails",
	Short: "Manage and test Bedrock Guardrails",
	Long: `List, inspect, and test Bedrock Guardrails.

Guardrails filter prompts and responses for content policy, PII, denied topics,
and custom word lists. Use 'guardrails test' to see exactly what gets blocked.`,
}

func init() {
	guardrailsCmd.AddCommand(guardrailsListCmd)
	guardrailsCmd.AddCommand(guardrailsDescribeCmd)
	guardrailsCmd.AddCommand(guardrailsTestCmd)
	guardrailsCmd.AddCommand(guardrailsCheckCmd)
}
