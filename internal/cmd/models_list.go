package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kernelpanic09/bedrock-cli/internal/models"
)

var flagModelsLiveList bool

var modelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available models",
	Long: `List models from the built-in catalog. Pass --live to query Bedrock directly
for the full list of foundation models available in your region (slower, requires AWS creds).`,
	RunE: runModelsList,
}

func init() {
	modelsListCmd.Flags().BoolVar(&flagModelsLiveList, "live", false, "query Bedrock API instead of local catalog")
}

func runModelsList(cmd *cobra.Command, args []string) error {
	if flagModelsLiveList {
		return runModelsListLive()
	}
	return runModelsListLocal()
}

func runModelsListLocal() error {
	all := models.All()

	if flagJSON {
		type row struct {
			ID          string  `json:"id"`
			Alias       string  `json:"alias"`
			Provider    string  `json:"provider"`
			Description string  `json:"description"`
			InputPrice  float64 `json:"input_price_per_1k"`
			OutputPrice float64 `json:"output_price_per_1k"`
			MaxTokens   int     `json:"max_tokens"`
		}
		var rows []row
		for _, m := range all {
			rows = append(rows, row{
				ID:          m.ID,
				Alias:       m.Alias,
				Provider:    m.Provider,
				Description: m.Description,
				InputPrice:  m.InputPrice,
				OutputPrice: m.OutputPrice,
				MaxTokens:   m.MaxTokens,
			})
		}
		enc := json.NewEncoder(out.Writer())
		enc.SetIndent("", "  ")
		return enc.Encode(rows)
	}

	headers := []string{"Alias", "Provider", "Description", "In $/1k", "Out $/1k"}
	var rows [][]string
	for _, m := range all {
		rows = append(rows, []string{
			m.Alias,
			m.Provider,
			m.Description,
			fmt.Sprintf("$%.5f", m.InputPrice),
			fmt.Sprintf("$%.5f", m.OutputPrice),
		})
	}
	out.Table(headers, rows)
	out.Dim("\nUse full model IDs with --model or 'bedrock-cli models info <alias>'.")
	return nil
}

func runModelsListLive() error {
	// TODO: implement live listing via the Bedrock control plane.
	// Placeholder that tells the user what to do in the meantime.
	return fmt.Errorf("live listing not yet implemented; run 'aws bedrock list-foundation-models --region %s'", resolveRegion())
}
