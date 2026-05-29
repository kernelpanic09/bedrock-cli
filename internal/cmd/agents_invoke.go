package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/client"
	"github.com/kernelpanic09/bedrock-cli/internal/config"
)

var flagAgentSession string

var agentsInvokeCmd = &cobra.Command{
	Use:   "invoke <agent-id> <task>",
	Short: "Invoke a Bedrock Agent and stream the response",
	Long: `Send a task to a Bedrock Agent. The response is streamed as it arrives.
Sessions are persisted locally so you can continue a conversation.

Examples:
  bedrock-cli agents invoke ABCDEF1234 "Summarize the Q3 report"
  bedrock-cli agents invoke ABCDEF1234 "What were the key risks?" --session sess-abc123`,
	Args: cobra.ExactArgs(2),
	RunE: runAgentsInvoke,
}

func init() {
	agentsInvokeCmd.Flags().StringVar(&flagAgentSession, "session", "", "session ID to continue a previous conversation")
}

// agentSession is a locally stored session record.
type agentSession struct {
	AgentID    string    `json:"agent_id"`
	SessionID  string    `json:"session_id"`
	LastActive time.Time `json:"last_active"`
}

func runAgentsInvoke(cmd *cobra.Command, args []string) error {
	agentID := args[0]
	task := args[1]

	sessionID := flagAgentSession
	if sessionID == "" {
		// Look for an existing session for this agent.
		sessions, _ := loadAgentSessions()
		for _, s := range sessions {
			if s.AgentID == agentID {
				sessionID = s.SessionID
				out.Dim(fmt.Sprintf("Continuing session %s", sessionID))
				break
			}
		}
	}
	if sessionID == "" {
		sessionID = fmt.Sprintf("bedrock-cli-%d", time.Now().UnixNano())
	}

	ctx := context.Background()
	cl, err := client.NewAgentClient(ctx, resolveRegion())
	if err != nil {
		return fmt.Errorf("creating agent client: %w", err)
	}

	var sb strings.Builder
	result, err := cl.InvokeAgent(ctx, agentID, sessionID, task, func(token string) {
		fmt.Print(token)
		sb.WriteString(token)
	})
	if err != nil {
		return err
	}
	fmt.Println()

	// Persist the session.
	_ = saveAgentSession(agentSession{
		AgentID:    agentID,
		SessionID:  result.SessionID,
		LastActive: time.Now(),
	})

	if flagJSON {
		return printJSON(map[string]any{
			"agent_id":   agentID,
			"session_id": result.SessionID,
			"response":   sb.String(),
		})
	}

	out.Dim(fmt.Sprintf("[session: %s]", result.SessionID))
	return nil
}

func agentSessionsPath() (string, error) {
	dir, err := config.CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "agent_sessions.json"), nil
}

func loadAgentSessions() ([]agentSession, error) {
	path, err := agentSessionsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var sessions []agentSession
	if err := json.Unmarshal(data, &sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}

func saveAgentSession(s agentSession) error {
	sessions, _ := loadAgentSessions()

	// Update or append.
	updated := false
	for i, existing := range sessions {
		if existing.AgentID == s.AgentID {
			sessions[i] = s
			updated = true
			break
		}
	}
	if !updated {
		sessions = append(sessions, s)
	}

	path, err := agentSessionsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
