package cmd

import "github.com/spf13/cobra"

var kbCmd = &cobra.Command{
	Use:   "kb",
	Short: "Interact with Bedrock Knowledge Bases",
	Long: `List and query Bedrock Knowledge Bases.

Knowledge Bases let you give models access to your own documents via
retrieval-augmented generation. You need to create them via the Bedrock console
or API first.`,
}

func init() {
	kbCmd.AddCommand(kbListCmd)
	kbCmd.AddCommand(kbQueryCmd)
}
