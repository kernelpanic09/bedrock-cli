package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/client"
)

var kbJobsCmd = &cobra.Command{
	Use:   "jobs <kb-id>",
	Short: "List recent ingestion jobs for a Knowledge Base",
	Long: `Show the status of ingestion jobs across all data sources of the KB.

Examples:
  bedrock-cli kb jobs KBID1234`,
	Args: cobra.ExactArgs(1),
	RunE: runKBJobs,
}

func runKBJobs(cmd *cobra.Command, args []string) error {
	kbID := args[0]
	ctx := context.Background()
	cl, err := client.NewAgentClient(ctx, resolveRegion())
	if err != nil {
		return fmt.Errorf("creating agent client: %w", err)
	}

	jobs, err := cl.ListIngestionJobs(ctx, kbID)
	if err != nil {
		return err
	}

	if len(jobs) == 0 {
		out.Dim("No ingestion jobs found for knowledge base " + kbID)
		return nil
	}

	if flagJSON {
		enc := json.NewEncoder(out.Writer())
		enc.SetIndent("", "  ")
		return enc.Encode(jobs)
	}

	headers := []string{"Job ID", "Data Source", "Status", "Started", "Updated"}
	var rows [][]string
	for _, j := range jobs {
		rows = append(rows, []string{j.JobID, j.DataSourceID, j.Status, j.StartedAt, j.UpdatedAt})
	}
	out.Table(headers, rows)
	return nil
}
