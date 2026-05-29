package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/client"
)

var agentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Bedrock Agents in the account",
	RunE:  runAgentsList,
}

func runAgentsList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	cl, err := client.NewAgentClient(ctx, resolveRegion())
	if err != nil {
		return fmt.Errorf("creating agent client: %w", err)
	}

	agents, err := cl.ListAgents(ctx)
	if err != nil {
		return err
	}

	if len(agents) == 0 {
		out.Dim("No agents found in region " + resolveRegion())
		return nil
	}

	if flagJSON {
		type row struct {
			AgentID   string `json:"agent_id"`
			AgentName string `json:"agent_name"`
			Status    string `json:"status"`
		}
		var rows []row
		for _, a := range agents {
			rows = append(rows, row{AgentID: a.AgentID, AgentName: a.AgentName, Status: a.Status})
		}
		enc := json.NewEncoder(out.Writer())
		enc.SetIndent("", "  ")
		return enc.Encode(rows)
	}

	headers := []string{"Agent ID", "Name", "Status"}
	var rows [][]string
	for _, a := range agents {
		rows = append(rows, []string{a.AgentID, a.AgentName, a.Status})
	}
	out.Table(headers, rows)
	return nil
}
