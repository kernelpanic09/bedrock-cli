package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/client"
)

var flagCheckVersion string

var guardrailsCheckCmd = &cobra.Command{
	Use:   "check <id> <file>",
	Short: "Run prompts from a file through a guardrail and report results",
	Long: `Read prompts from a file (one per line or a JSON array of strings) and
run each through the guardrail. Prints a pass/fail report.

Examples:
  bedrock-cli guardrails check gr-1234 prompts.txt
  bedrock-cli guardrails check gr-1234 prompts.json --json`,
	Args: cobra.ExactArgs(2),
	RunE: runGuardrailsCheck,
}

func init() {
	guardrailsCheckCmd.Flags().StringVar(&flagCheckVersion, "version", "", "guardrail version (default: DRAFT)")
}

func runGuardrailsCheck(cmd *cobra.Command, args []string) error {
	guardrailID := args[0]
	filePath := args[1]

	prompts, err := loadPromptsFromFile(filePath)
	if err != nil {
		return err
	}
	if len(prompts) == 0 {
		out.Dim("No prompts found in " + filePath)
		return nil
	}

	ctx := context.Background()
	cl, err := client.NewAgentClient(ctx, resolveRegion())
	if err != nil {
		return fmt.Errorf("creating agent client: %w", err)
	}

	type checkRow struct {
		Prompt  string `json:"prompt"`
		Action  string `json:"action"`
		Details string `json:"details,omitempty"`
	}
	var rows []checkRow

	passed, blocked := 0, 0
	for i, p := range prompts {
		result, err := cl.TestGuardrail(ctx, guardrailID, flagCheckVersion, p)
		if err != nil {
			rows = append(rows, checkRow{Prompt: truncate(p, 60), Action: "ERROR", Details: err.Error()})
			continue
		}
		details := buildAssessmentSummary(result.Assessments)
		rows = append(rows, checkRow{Prompt: truncate(p, 60), Action: result.Action, Details: details})
		if result.Action == "NONE" {
			passed++
		} else {
			blocked++
		}
		if !flagJSON {
			status := "PASS"
			if result.Action != "NONE" {
				status = "BLOCK"
			}
			out.Printf("[%d/%d] %-6s %s\n", i+1, len(prompts), status, truncate(p, 50))
		}
	}

	if flagJSON {
		type report struct {
			Total   int        `json:"total"`
			Passed  int        `json:"passed"`
			Blocked int        `json:"blocked"`
			Results []checkRow `json:"results"`
		}
		enc := json.NewEncoder(out.Writer())
		enc.SetIndent("", "  ")
		return enc.Encode(report{Total: len(prompts), Passed: passed, Blocked: blocked, Results: rows})
	}

	out.Println("")
	out.Printf("Total: %d  Passed: %d  Blocked: %d\n", len(prompts), passed, blocked)
	return nil
}

func loadPromptsFromFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	// Try JSON array first.
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		return arr, nil
	}

	// Fall back to one prompt per non-empty line.
	var prompts []string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			prompts = append(prompts, line)
		}
	}
	return prompts, scanner.Err()
}

func buildAssessmentSummary(assessments []client.GuardrailAssessment) string {
	var parts []string
	for _, a := range assessments {
		if a.TopicPolicy != "" {
			parts = append(parts, "topic:"+a.TopicPolicy)
		}
		if a.ContentPolicy != "" {
			parts = append(parts, "content:"+a.ContentPolicy)
		}
		if a.PIIPolicy != "" {
			parts = append(parts, "pii:"+a.PIIPolicy)
		}
		if a.WordPolicy != "" {
			parts = append(parts, "words:"+a.WordPolicy)
		}
	}
	return strings.Join(parts, "; ")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}
