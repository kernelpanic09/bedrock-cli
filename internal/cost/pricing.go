package cost

import "github.com/kernelpanic09/bedrock-cli/internal/models"

// Calculate returns the USD cost for a given number of input and output tokens
// for the specified model. It resolves both aliases and full model IDs.
// Returns 0.0 if the model isn't in the catalog (better than crashing).
func Calculate(modelAliasOrID string, inputTokens, outputTokens int) float64 {
	m, err := models.Resolve(modelAliasOrID)
	if err != nil {
		return 0.0
	}
	// Pricing is per 1000 tokens.
	input := float64(inputTokens) / 1000.0 * m.InputPrice
	output := float64(outputTokens) / 1000.0 * m.OutputPrice
	return input + output
}
