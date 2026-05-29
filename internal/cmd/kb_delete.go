package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/client"
)

var flagKBDeleteForce bool

var kbDeleteCmd = &cobra.Command{
	Use:   "delete <kb-id>",
	Short: "Delete a Knowledge Base and its data sources",
	Long: `Delete a Bedrock Knowledge Base. This also removes attached data sources.
The vector store collection itself is not deleted (managed separately).

You'll be prompted to confirm unless --force is used.

Examples:
  bedrock-cli kb delete KBID1234
  bedrock-cli kb delete KBID1234 --force`,
	Args: cobra.ExactArgs(1),
	RunE: runKBDelete,
}

func init() {
	kbDeleteCmd.Flags().BoolVar(&flagKBDeleteForce, "force", false, "skip confirmation prompt")
}

func runKBDelete(cmd *cobra.Command, args []string) error {
	kbID := args[0]

	if !flagKBDeleteForce {
		fmt.Printf("Delete knowledge base %s? This cannot be undone. [y/N]: ", kbID)
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(strings.ToLower(line))
		if line != "y" && line != "yes" {
			out.Dim("Aborted.")
			return nil
		}
	}

	ctx := context.Background()
	cl, err := client.NewAgentClient(ctx, resolveRegion())
	if err != nil {
		return fmt.Errorf("creating agent client: %w", err)
	}

	if err := cl.DeleteKnowledgeBase(ctx, kbID); err != nil {
		return err
	}

	out.Success("Deleted knowledge base " + kbID)
	return nil
}
