package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/client"
)

var kbListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available knowledge bases",
	Long: `List Bedrock Knowledge Bases in your account and region.

Note: listing knowledge bases requires the bedrock-agent control plane.
If this command fails, use the AWS CLI directly:
  aws bedrock-agent list-knowledge-bases --region <region>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cl, err := client.New(ctx, resolveRegion())
		if err != nil {
			return fmt.Errorf("creating Bedrock client: %w", err)
		}

		kbs, err := cl.ListKnowledgeBases(ctx)
		if err != nil {
			return err
		}

		if len(kbs) == 0 {
			out.Dim("No knowledge bases found in region " + resolveRegion())
			return nil
		}

		headers := []string{"ID", "Name", "Status", "Description"}
		var rows [][]string
		for _, kb := range kbs {
			rows = append(rows, []string{kb.ID, kb.Name, kb.Status, kb.Description})
		}
		out.Table(headers, rows)
		return nil
	},
}
