package cmd

import "github.com/spf13/cobra"

var kbCmd = &cobra.Command{
	Use:   "kb",
	Short: "Manage Bedrock Knowledge Bases",
	Long: `Create, inspect, sync, and query Bedrock Knowledge Bases.

Knowledge Bases give models access to your own documents via retrieval-augmented
generation. You can manage the full lifecycle here: create a KB, upload docs,
trigger ingestion, and query the results.`,
}

func init() {
	kbCmd.AddCommand(kbListCmd)
	kbCmd.AddCommand(kbDescribeCmd)
	kbCmd.AddCommand(kbCreateCmd)
	kbCmd.AddCommand(kbAddDocsCmd)
	kbCmd.AddCommand(kbSyncCmd)
	kbCmd.AddCommand(kbJobsCmd)
	kbCmd.AddCommand(kbQueryCmd)
	kbCmd.AddCommand(kbDeleteCmd)
}
