package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/cost"
)

var flagCostByModelSince string

var costByModelCmd = &cobra.Command{
	Use:   "by-model",
	Short: "Show cost broken down by model",
	Long: `Aggregate spending grouped by model. Useful for understanding which models are
costing the most over a period.

Examples:
  bedrock-cli cost by-model
  bedrock-cli cost by-model --since 7d`,
	RunE: runCostByModel,
}

func init() {
	costByModelCmd.Flags().StringVar(&flagCostByModelSince, "since", "", "filter to recent period (e.g. 24h, 7d)")
}

func runCostByModel(cmd *cobra.Command, args []string) error {
	since, err := parseSince(flagCostByModelSince)
	if err != nil {
		return err
	}

	tracker, err := cost.Open()
	if err != nil {
		return fmt.Errorf("opening cost database: %w", err)
	}
	defer tracker.Close()

	summaries, err := tracker.ByModel(since)
	if err != nil {
		return err
	}

	if len(summaries) == 0 {
		out.Dim("No invocations recorded yet.")
		return nil
	}

	if flagJSON {
		type row struct {
			Model        string  `json:"model"`
			Calls        int     `json:"calls"`
			InputTokens  int     `json:"input_tokens"`
			OutputTokens int     `json:"output_tokens"`
			TotalCost    float64 `json:"total_cost_usd"`
			CachedCalls  int     `json:"cached_calls"`
		}
		var rows []row
		for _, s := range summaries {
			rows = append(rows, row{
				Model:        s.Model,
				Calls:        s.Calls,
				InputTokens:  s.InputTokens,
				OutputTokens: s.OutputTokens,
				TotalCost:    s.TotalCost,
				CachedCalls:  s.CachedCalls,
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

	out.Header(fmt.Sprintf("Cost by model (%s)", sinceLabel))
	out.Println("")

	var totalCost float64
	for _, s := range summaries {
		totalCost += s.TotalCost
	}

	headers := []string{"Model", "Calls", "In Tokens", "Out Tokens", "Cached", "Total Cost", "% of Total"}
	var rows [][]string
	for _, s := range summaries {
		pct := 0.0
		if totalCost > 0 {
			pct = s.TotalCost / totalCost * 100
		}
		rows = append(rows, []string{
			s.Model,
			strconv.Itoa(s.Calls),
			strconv.Itoa(s.InputTokens),
			strconv.Itoa(s.OutputTokens),
			strconv.Itoa(s.CachedCalls),
			fmt.Sprintf("$%.4f", s.TotalCost),
			fmt.Sprintf("%.1f%%", pct),
		})
	}
	out.Table(headers, rows)
	out.Println("")
	out.Bold(fmt.Sprintf("Total: $%.4f", totalCost))

	return nil
}
