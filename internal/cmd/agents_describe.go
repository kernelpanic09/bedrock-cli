package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/client"
)

var agentsDescribeCmd = &cobra.Command{
	Use:   "describe <agent-id>",
	Short: "Show full details for a Bedrock Agent",
	Long: `Print the configuration of a Bedrock Agent: model, instruction, action groups,
and attached knowledge bases.

Examples:
  bedrock-cli agents describe ABCDEF1234`,
	Args: cobra.ExactArgs(1),
	RunE: runAgentsDescribe,
}

func runAgentsDescribe(cmd *cobra.Command, args []string) error {
	agentID := args[0]
	ctx := context.Background()
	cl, err := client.NewAgentClient(ctx, resolveRegion())
	if err != nil {
		return fmt.Errorf("creating agent client: %w", err)
	}

	detail, err := cl.DescribeAgent(ctx, agentID)
	if err != nil {
		return err
	}

	if flagJSON {
		enc := json.NewEncoder(out.Writer())
		enc.SetIndent("", "  ")
		return enc.Encode(detail)
	}

	out.Header("Agent: " + detail.AgentName)
	out.Printf("\n  ID      : %s\n", detail.AgentID)
	out.Printf("  Status  : %s\n", detail.Status)
	out.Printf("  Model   : %s\n", detail.Model)
	if detail.Description != "" {
		out.Printf("  Desc    : %s\n", detail.Description)
	}
	if len(detail.ActionGroups) > 0 {
		out.Printf("  Actions : %s\n", strings.Join(detail.ActionGroups, ", "))
	}
	if len(detail.KnowledgeBases) > 0 {
		out.Printf("  KBs     : %s\n", strings.Join(detail.KnowledgeBases, ", "))
	}
	if detail.Instruction != "" {
		out.Println("")
		out.Bold("Instruction:")
		out.Println(detail.Instruction)
	}
	return nil
}
