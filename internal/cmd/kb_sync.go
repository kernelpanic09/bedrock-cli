package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/client"
)

var flagKBSyncDS string

var kbSyncCmd = &cobra.Command{
	Use:   "sync <kb-id>",
	Short: "Trigger an ingestion job for a Knowledge Base",
	Long: `Start a new ingestion job to sync the S3 data source into the vector store.

If the KB has multiple data sources, use --data-source to target a specific one.
Omitting it picks the first data source.

Examples:
  bedrock-cli kb sync KBID1234
  bedrock-cli kb sync KBID1234 --data-source DS_ID`,
	Args: cobra.ExactArgs(1),
	RunE: runKBSync,
}

func init() {
	kbSyncCmd.Flags().StringVar(&flagKBSyncDS, "data-source", "", "data source ID to sync (default: first one)")
}

func runKBSync(cmd *cobra.Command, args []string) error {
	kbID := args[0]
	ctx := context.Background()
	cl, err := client.NewAgentClient(ctx, resolveRegion())
	if err != nil {
		return fmt.Errorf("creating agent client: %w", err)
	}

	job, err := cl.StartIngestionJob(ctx, kbID, flagKBSyncDS)
	if err != nil {
		return err
	}

	out.Success(fmt.Sprintf("Ingestion job started"))
	out.Printf("  Job ID      : %s\n", job.JobID)
	out.Printf("  Data source : %s\n", job.DataSourceID)
	out.Printf("  Status      : %s\n", job.Status)
	if len(job.FailureReasons) > 0 {
		for _, r := range job.FailureReasons {
			out.Warn(r)
		}
	}
	out.Dim("Track progress with: bedrock-cli kb jobs " + kbID)
	return nil
}
