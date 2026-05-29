package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/client"
)

var kbListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Knowledge Bases in the account",
	Long: `List Bedrock Knowledge Bases, showing ID, name, status, and vector store type.

Examples:
  bedrock-cli kb list
  bedrock-cli kb list --json`,
	RunE: runKBList,
}

func runKBList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	cl, err := client.NewAgentClient(ctx, resolveRegion())
	if err != nil {
		return fmt.Errorf("creating agent client: %w", err)
	}

	kbs, err := cl.ListKnowledgeBases(ctx)
	if err != nil {
		return err
	}

	if len(kbs) == 0 {
		out.Dim("No knowledge bases found in region " + resolveRegion())
		return nil
	}

	if flagJSON {
		type row struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Status      string `json:"status"`
			VectorStore string `json:"vector_store,omitempty"`
			Description string `json:"description,omitempty"`
		}
		var rows []row
		for _, kb := range kbs {
			rows = append(rows, row{
				ID:          kb.ID,
				Name:        kb.Name,
				Status:      kb.Status,
				VectorStore: kb.VectorStore,
				Description: kb.Description,
			})
		}
		enc := json.NewEncoder(out.Writer())
		enc.SetIndent("", "  ")
		return enc.Encode(rows)
	}

	headers := []string{"ID", "Name", "Status", "Vector Store"}
	var rows [][]string
	for _, kb := range kbs {
		rows = append(rows, []string{kb.ID, kb.Name, kb.Status, kb.VectorStore})
	}
	out.Table(headers, rows)
	return nil
}
