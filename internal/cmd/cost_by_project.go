package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/cost"
)

var flagCostByProjectSince string

var costByProjectCmd = &cobra.Command{
	Use:   "by-project",
	Short: "Show cost broken down by project tag",
	Long: `Aggregate spending grouped by project. Pass --project to 'prompt' or 'compare'
to tag invocations. Untagged invocations appear under an empty project label.

Examples:
  bedrock-cli cost by-project
  bedrock-cli cost by-project --since 7d`,
	RunE: runCostByProject,
}

func init() {
	costByProjectCmd.Flags().StringVar(&flagCostByProjectSince, "since", "", "filter to recent period (e.g. 24h, 7d)")
}

func runCostByProject(cmd *cobra.Command, args []string) error {
	since, err := parseSince(flagCostByProjectSince)
	if err != nil {
		return err
	}

	tracker, err := cost.Open()
	if err != nil {
		return fmt.Errorf("opening cost database: %w", err)
	}
	defer tracker.Close()

	summaries, err := tracker.ByProject(since)
	if err != nil {
		return err
	}

	if len(summaries) == 0 {
		out.Dim("No invocations recorded yet.")
		return nil
	}

	if flagJSON {
		type row struct {
			Project      string  `json:"project"`
			Calls        int     `json:"calls"`
			InputTokens  int     `json:"input_tokens"`
			OutputTokens int     `json:"output_tokens"`
			TotalCost    float64 `json:"total_cost_usd"`
		}
		var rows []row
		for _, s := range summaries {
			rows = append(rows, row{
				Project:      s.Project,
				Calls:        s.Calls,
				InputTokens:  s.InputTokens,
				OutputTokens: s.OutputTokens,
				TotalCost:    s.TotalCost,
			})
		}
		enc := json.NewEncoder(out.Writer())
		enc.SetIndent("", "  ")
		return enc.Encode(rows)
	}

	sinceLabel := "all time"
	if !since.IsZero() {
		sinceLabel = fmt.Sprintf("since %s", since.Format("2006-01-02"))
	}

	out.Header(fmt.Sprintf("Cost by project (%s)", sinceLabel))
	out.Println("")

	var totalCost float64
	for _, s := range summaries {
		totalCost += s.TotalCost
	}

	headers := []string{"Project", "Calls", "In Tokens", "Out Tokens", "Total Cost", "% of Total"}
	var rows [][]string
	for _, s := range summaries {
		label := s.Project
		if label == "" {
			label = "(untagged)"
		}
		pct := 0.0
		if totalCost > 0 {
			pct = s.TotalCost / totalCost * 100
		}
		rows = append(rows, []string{
			label,
			strconv.Itoa(s.Calls),
			strconv.Itoa(s.InputTokens),
			strconv.Itoa(s.OutputTokens),
			fmt.Sprintf("$%.4f", s.TotalCost),
			fmt.Sprintf("%.1f%%", pct),
		})
	}
	out.Table(headers, rows)
	out.Println("")
	out.Bold(fmt.Sprintf("Total: $%.4f", totalCost))
	return nil
}
