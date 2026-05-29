package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var agentsSessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "Show locally stored agent sessions",
	Long: `List recent agent sessions saved by 'bedrock-cli agents invoke'.
Use the session ID with --session to continue a conversation.

Examples:
  bedrock-cli agents sessions
  bedrock-cli agents invoke AGENT_ID "follow-up question" --session <id>`,
	RunE: runAgentsSessions,
}

func runAgentsSessions(cmd *cobra.Command, args []string) error {
	sessions, err := loadAgentSessions()
	if err != nil {
		return fmt.Errorf("reading sessions: %w", err)
	}

	if len(sessions) == 0 {
		out.Dim("No saved sessions. Run 'bedrock-cli agents invoke' to start one.")
		return nil
	}

	if flagJSON {
		enc := json.NewEncoder(out.Writer())
		enc.SetIndent("", "  ")
		return enc.Encode(sessions)
	}

	headers := []string{"Agent ID", "Session ID", "Last Active"}
	var rows [][]string
	for _, s := range sessions {
		rows = append(rows, []string{
			s.AgentID,
			s.SessionID,
			s.LastActive.Local().Format("2006-01-02 15:04:05"),
		})
	}
	out.Table(headers, rows)
	return nil
}
