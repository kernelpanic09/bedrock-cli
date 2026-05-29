package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/client"
)

var kbAddDocsCmd = &cobra.Command{
	Use:   "add-docs <kb-id> <path>...",
	Short: "Upload local files to a Knowledge Base and trigger ingestion",
	Long: `Upload one or more local files or directories to the S3 bucket backing
the Knowledge Base, then start an ingestion job.

Examples:
  bedrock-cli kb add-docs KBID1234 ./docs
  bedrock-cli kb add-docs KBID1234 report.pdf meeting-notes.txt`,
	Args: cobra.MinimumNArgs(2),
	RunE: runKBAddDocs,
}

func runKBAddDocs(cmd *cobra.Command, args []string) error {
	kbID := args[0]
	paths := args[1:]

	ctx := context.Background()
	cl, err := client.NewAgentClient(ctx, resolveRegion())
	if err != nil {
		return fmt.Errorf("creating agent client: %w", err)
	}

	count, err := cl.UploadDocsToKB(ctx, kbID, paths, func(name string) {
		out.Dim("  uploaded: " + name)
	})
	if err != nil {
		return err
	}
	out.Success(fmt.Sprintf("Uploaded %d file(s).", count))

	// Start ingestion.
	out.Dim("Starting ingestion job...")
	job, err := cl.StartIngestionJob(ctx, kbID, "")
	if err != nil {
		return fmt.Errorf("starting ingestion: %w", err)
	}
	out.Success(fmt.Sprintf("Ingestion job started: %s (status: %s)", job.JobID, job.Status))
	out.Dim("Track progress with: bedrock-cli kb jobs " + kbID)
	return nil
}
