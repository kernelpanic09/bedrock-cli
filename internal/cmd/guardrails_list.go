package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/client"
)

var guardrailsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Bedrock Guardrails",
	RunE:  runGuardrailsList,
}

func runGuardrailsList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	cl, err := client.NewAgentClient(ctx, resolveRegion())
	if err != nil {
		return fmt.Errorf("creating agent client: %w", err)
	}

	gs, err := cl.ListGuardrails(ctx)
	if err != nil {
		return err
	}

	if len(gs) == 0 {
		out.Dim("No guardrails found in region " + resolveRegion())
		return nil
	}

	if flagJSON {
		enc := json.NewEncoder(out.Writer())
		enc.SetIndent("", "  ")
		return enc.Encode(gs)
	}

	headers := []string{"ID", "Name", "Version", "Status"}
	var rows [][]string
	for _, g := range gs {
		rows = append(rows, []string{g.ID, g.Name, g.Version, g.Status})
	}
	out.Table(headers, rows)
	return nil
}
