package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/client"
)

var flagKBMaxResults int

var kbQueryCmd = &cobra.Command{
	Use:   "query <kb-id> <query>",
	Short: "Query a Knowledge Base",
	Long: `Send a natural-language query to a Bedrock Knowledge Base and return
the top matching document chunks.

Examples:
  bedrock-cli kb query KBID1234 "What's our backup policy?"
  bedrock-cli kb query KBID1234 "How do we handle incident response?" --max-results 5`,
	Args: cobra.ExactArgs(2),
	RunE: runKBQuery,
}

func init() {
	kbQueryCmd.Flags().IntVar(&flagKBMaxResults, "max-results", 3, "number of results to return")
}

func runKBQuery(cmd *cobra.Command, args []string) error {
	kbID := args[0]
	query := args[1]

	ctx := context.Background()
	cl, err := client.NewAgentClient(ctx, resolveRegion())
	if err != nil {
		return fmt.Errorf("creating agent client: %w", err)
	}

	results, err := cl.QueryKB(ctx, kbID, query, flagKBMaxResults)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		out.Dim("No results found.")
		return nil
	}

	if flagJSON {
		type row struct {
			Content string  `json:"content"`
			Score   float64 `json:"score"`
			Source  string  `json:"source,omitempty"`
		}
		var rows []row
		for _, r := range results {
			rows = append(rows, row{Content: r.Content, Score: r.Score, Source: r.Source})
		}
		enc := json.NewEncoder(out.Writer())
		enc.SetIndent("", "  ")
		return enc.Encode(rows)
	}

	for i, r := range results {
		label := fmt.Sprintf("Result %d", i+1)
		if r.Source != "" {
			label += " | " + r.Source
		}
		label += fmt.Sprintf(" | score %.2f", r.Score)

		content := strings.TrimSpace(r.Content)
		out.BoxedResponse(label, content)
		fmt.Println()
	}
	return nil
}
