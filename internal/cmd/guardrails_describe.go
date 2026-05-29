package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/client"
)

var guardrailsDescribeCmd = &cobra.Command{
	Use:   "describe <id>",
	Short: "Show full configuration for a guardrail",
	Long: `Print the guardrail's content filters, PII policies, denied topics, and word lists.

Examples:
  bedrock-cli guardrails describe gr-1234abcd`,
	Args: cobra.ExactArgs(1),
	RunE: runGuardrailsDescribe,
}

func runGuardrailsDescribe(cmd *cobra.Command, args []string) error {
	guardrailID := args[0]
	ctx := context.Background()
	cl, err := client.NewAgentClient(ctx, resolveRegion())
	if err != nil {
		return fmt.Errorf("creating agent client: %w", err)
	}

	detail, err := cl.DescribeGuardrail(ctx, guardrailID)
	if err != nil {
		return err
	}

	if flagJSON {
		enc := json.NewEncoder(out.Writer())
		enc.SetIndent("", "  ")
		return enc.Encode(detail)
	}

	out.Header("Guardrail: " + detail.Name)
	out.Printf("\n  ID      : %s\n", detail.ID)
	out.Printf("  Version : %s\n", detail.Version)
	out.Printf("  Status  : %s\n", detail.Status)
	if detail.Description != "" {
		out.Printf("  Desc    : %s\n", detail.Description)
	}

	if len(detail.TopicPolicies) > 0 {
		out.Println("")
		out.Bold("Denied topics:")
		for _, t := range detail.TopicPolicies {
			out.Printf("  - %s\n", t)
		}
	}

	if len(detail.ContentFilters) > 0 {
		out.Println("")
		out.Bold("Content filters:")
		headers := []string{"Type", "Input Strength", "Output Strength"}
		var rows [][]string
		for _, f := range detail.ContentFilters {
			rows = append(rows, []string{f.Type, f.InputStrength, f.OutputStrength})
		}
		out.Table(headers, rows)
	}

	if len(detail.PIIRedactions) > 0 {
		out.Println("")
		out.Bold("PII redactions:")
		out.Println("  " + strings.Join(detail.PIIRedactions, ", "))
	}

	if len(detail.WordFilters) > 0 {
		out.Println("")
		out.Bold("Word filters:")
		out.Println("  " + strings.Join(detail.WordFilters, ", "))
	}
	return nil
}
