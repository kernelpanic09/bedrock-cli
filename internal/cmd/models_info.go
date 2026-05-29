package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/models"
)

var modelsInfoCmd = &cobra.Command{
	Use:   "info <alias-or-id>",
	Short: "Show details for a specific model",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := models.Resolve(args[0])
		if err != nil {
			return err
		}

		if flagJSON {
			type detail struct {
				ID               string  `json:"id"`
				Alias            string  `json:"alias"`
				Provider         string  `json:"provider"`
				Description      string  `json:"description"`
				InputPrice       float64 `json:"input_price_per_1k"`
				OutputPrice      float64 `json:"output_price_per_1k"`
				MaxTokens        int     `json:"max_tokens"`
				SupportsStream   bool    `json:"supports_streaming"`
				SupportsConverse bool    `json:"supports_converse"`
			}
			enc := json.NewEncoder(out.Writer())
			enc.SetIndent("", "  ")
			return enc.Encode(detail{
				ID:               m.ID,
				Alias:            m.Alias,
				Provider:         m.Provider,
				Description:      m.Description,
				InputPrice:       m.InputPrice,
				OutputPrice:      m.OutputPrice,
				MaxTokens:        m.MaxTokens,
				SupportsStream:   m.SupportsStreaming,
				SupportsConverse: m.SupportsConverse,
			})
		}

		out.Bold(m.Alias + " / " + m.ID)
		out.Println("")
		out.Table([]string{"Field", "Value"}, [][]string{
			{"Provider", m.Provider},
			{"Description", m.Description},
			{"Max Tokens", fmt.Sprintf("%d", m.MaxTokens)},
			{"Input price", fmt.Sprintf("$%.5f / 1k tokens", m.InputPrice)},
			{"Output price", fmt.Sprintf("$%.5f / 1k tokens", m.OutputPrice)},
			{"Streaming", fmt.Sprintf("%v", m.SupportsStreaming)},
			{"Converse API", fmt.Sprintf("%v", m.SupportsConverse)},
		})
		return nil
	},
}
