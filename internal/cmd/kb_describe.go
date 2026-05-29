package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/client"
)

var kbDescribeCmd = &cobra.Command{
	Use:   "describe <kb-id>",
	Short: "Show full details for a Knowledge Base",
	Long: `Print the Knowledge Base configuration: embedding model, vector store,
role ARN, and all attached data sources.

Examples:
  bedrock-cli kb describe KBID1234`,
	Args: cobra.ExactArgs(1),
	RunE: runKBDescribe,
}

func runKBDescribe(cmd *cobra.Command, args []string) error {
	kbID := args[0]
	ctx := context.Background()
	cl, err := client.NewAgentClient(ctx, resolveRegion())
	if err != nil {
		return fmt.Errorf("creating agent client: %w", err)
	}

	detail, err := cl.DescribeKnowledgeBase(ctx, kbID)
	if err != nil {
		return err
	}

	if flagJSON {
		enc := json.NewEncoder(out.Writer())
		enc.SetIndent("", "  ")
		return enc.Encode(detail)
	}

	out.Header("Knowledge Base: " + detail.Name)
	out.Printf("\n  ID          : %s\n", detail.ID)
	out.Printf("  Status      : %s\n", detail.Status)
	out.Printf("  Vector store: %s\n", detail.VectorStore)
	out.Printf("  Embedding   : %s\n", detail.EmbeddingModel)
	out.Printf("  Role ARN    : %s\n", detail.RoleARN)

	if len(detail.DataSources) > 0 {
		out.Println("")
		out.Bold("Data sources:")
		headers := []string{"ID", "Name", "Type", "Bucket", "Status"}
		var rows [][]string
		for _, ds := range detail.DataSources {
			rows = append(rows, []string{ds.ID, ds.Name, ds.Type, ds.S3Bucket, ds.Status})
		}
		out.Table(headers, rows)
	}
	return nil
}
