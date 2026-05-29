package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/cost"
)

var flagCostSince string

var costSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show total cost and recent invocations",
	Long: `Show a summary of your Bedrock spending tracked by bedrock-cli.

--since accepts a duration like 24h, 7d, 30d.

Examples:
  bedrock-cli cost summary
  bedrock-cli cost summary --since 7d
  bedrock-cli cost summary --json`,
	RunE: runCostSummary,
}

func init() {
	costSummaryCmd.Flags().StringVar(&flagCostSince, "since", "", "filter to recent period (e.g. 24h, 7d)")
}

func runCostSummary(cmd *cobra.Command, args []string) error {
	since, err := parseSince(flagCostSince)
	if err != nil {
		return err
	}

	tracker, err := cost.Open()
	if err != nil {
		return fmt.Errorf("opening cost database: %w", err)
	}
	defer tracker.Close()

	invocations, err := tracker.Summary(since)
	if err != nil {
		return err
	}

	total, err := tracker.TotalCost(since)
	if err != nil {
		return err
	}

	if flagJSON {
		type row struct {
			Timestamp    string  `json:"timestamp"`
			Model        string  `json:"model"`
			InputTokens  int     `json:"input_tokens"`
			OutputTokens int     `json:"output_tokens"`
			CostUSD      float64 `json:"cost_usd"`
			Cached       bool    `json:"cached"`
			TemplateName string  `json:"template,omitempty"`
			DurationMs   int64   `json:"duration_ms"`
		}
		type summary struct {
			TotalCostUSD float64 `json:"total_cost_usd"`
			Count        int     `json:"count"`
			Invocations  []row   `json:"invocations"`
		}
		var rows []row
		for _, inv := range invocations {
			rows = append(rows, row{
				Timestamp:    inv.Timestamp.Format(time.RFC3339),
				Model:        inv.Model,
				InputTokens:  inv.InputTokens,
				OutputTokens: inv.OutputTokens,
				CostUSD:      inv.CostUSD,
				Cached:       inv.Cached,
				TemplateName: inv.TemplateName,
				DurationMs:   inv.DurationMs,
			})
		}
		enc := json.NewEncoder(out.Writer())
		enc.SetIndent("", "  ")
		return enc.Encode(summary{TotalCostUSD: total, Count: len(invocations), Invocations: rows})
	}

	sinceLabel := "all time"
	if !since.IsZero() {
		sinceLabel = fmt.Sprintf("since %s", since.Format("2006-01-02"))
	}

	out.Header(fmt.Sprintf("Cost summary (%s)", sinceLabel))
	out.Printf("\n  Total: $%.4f across %d invocations\n\n", total, len(invocations))

	if len(invocations) == 0 {
		out.Dim("No invocations recorded yet.")
		return nil
	}

	// Show the 10 most recent.
	limit := 10
	if len(invocations) < limit {
		limit = len(invocations)
	}

	headers := []string{"Time", "Model", "In", "Out", "Cost", "Cached", "Duration"}
	var rows [][]string
	for _, inv := range invocations[:limit] {
		cached := ""
		if inv.Cached {
			cached = "yes"
		}
		rows = append(rows, []string{
			inv.Timestamp.Local().Format("01-02 15:04"),
			inv.Model,
			strconv.Itoa(inv.InputTokens),
			strconv.Itoa(inv.OutputTokens),
			fmt.Sprintf("$%.4f", inv.CostUSD),
			cached,
			fmt.Sprintf("%.1fs", float64(inv.DurationMs)/1000),
		})
	}
	out.Table(headers, rows)

	if len(invocations) > limit {
		out.Dim(fmt.Sprintf("\n  ... and %d older invocations. Use --json for the full list.", len(invocations)-limit))
	}
	return nil
}

// parseSince converts strings like "7d", "24h", "30d" into a time.Time.
func parseSince(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}

	s = strings.ToLower(strings.TrimSpace(s))
	if strings.HasSuffix(s, "d") {
		n, err := strconv.Atoi(s[:len(s)-1])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid duration %q: expected a number followed by d or h", s)
		}
		return time.Now().Add(-time.Duration(n) * 24 * time.Hour), nil
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid duration %q: %w", s, err)
	}
	return time.Now().Add(-d), nil
}
