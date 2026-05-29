package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/client"
)

var flagGuardrailVersion string

var guardrailsTestCmd = &cobra.Command{
	Use:   "test <id> <prompt>",
	Short: "Apply a guardrail to a prompt and show what gets blocked",
	Long: `Run a prompt through a guardrail and print the action taken (NONE or
GUARDRAIL_INTERVENED) along with which policy triggered.

Examples:
  bedrock-cli guardrails test gr-1234 "Tell me how to hack into a server"
  bedrock-cli guardrails test gr-1234 "My SSN is 123-45-6789" --version 1`,
	Args: cobra.ExactArgs(2),
	RunE: runGuardrailsTest,
}

func init() {
	guardrailsTestCmd.Flags().StringVar(&flagGuardrailVersion, "version", "", "guardrail version (default: DRAFT)")
}

func runGuardrailsTest(cmd *cobra.Command, args []string) error {
	guardrailID := args[0]
	promptText := args[1]

	ctx := context.Background()
	cl, err := client.NewAgentClient(ctx, resolveRegion())
	if err != nil {
		return fmt.Errorf("creating agent client: %w", err)
	}

	result, err := cl.TestGuardrail(ctx, guardrailID, flagGuardrailVersion, promptText)
	if err != nil {
		return err
	}

	if flagJSON {
		enc := json.NewEncoder(out.Writer())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	out.Header("Guardrail result")
	out.Printf("\n  Action : %s\n", result.Action)

	if len(result.Outputs) > 0 {
		out.Println("")
		out.Bold("Output after filtering:")
		for _, o := range result.Outputs {
			out.Printf("  %s\n", o)
		}
	}

	if len(result.Assessments) > 0 {
		out.Println("")
		out.Bold("Assessments:")
		for i, a := range result.Assessments {
			out.Printf("  [%d]", i+1)
			if a.TopicPolicy != "" {
				out.Printf("  Topic   : %s\n", a.TopicPolicy)
			}
			if a.ContentPolicy != "" {
				out.Printf("  Content : %s\n", a.ContentPolicy)
			}
			if a.PIIPolicy != "" {
				out.Printf("  PII     : %s\n", a.PIIPolicy)
			}
			if a.WordPolicy != "" {
				out.Printf("  Words   : %s\n", a.WordPolicy)
			}
		}
	}
	return nil
}
