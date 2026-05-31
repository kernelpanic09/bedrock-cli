package models

import "fmt"

// Model holds metadata about a Bedrock foundation model.
type Model struct {
	ID          string
	Alias       string
	Provider    string
	Description string
	// Pricing per 1000 tokens in USD.
	InputPrice  float64
	OutputPrice float64
	// SupportsStreaming indicates the model can use InvokeModelWithResponseStream.
	SupportsStreaming bool
	// SupportsConverse indicates the model can use the Converse / ConverseStream API.
	SupportsConverse bool
	MaxTokens        int
}

// catalog is the known set of Bedrock models with their aliases and pricing.
// Pricing as of 2024-Q4; always verify against the AWS Bedrock pricing page.
var catalog = []Model{
	{
		ID:                "anthropic.claude-haiku-4-5-20251001-v1:0",
		Alias:             "haiku",
		Provider:          "Anthropic",
		Description:       "Claude Haiku 4.5 - fast and affordable",
		InputPrice:        0.00025,
		OutputPrice:       0.00125,
		SupportsStreaming: true,
		SupportsConverse:  true,
		MaxTokens:         8192,
	},
	{
		ID:                "anthropic.claude-sonnet-4-6-20250514-v1:0",
		Alias:             "sonnet",
		Provider:          "Anthropic",
		Description:       "Claude Sonnet 4.6 - balanced performance",
		InputPrice:        0.003,
		OutputPrice:       0.015,
		SupportsStreaming: true,
		SupportsConverse:  true,
		MaxTokens:         8192,
	},
	{
		ID:                "anthropic.claude-opus-4-7-20250219-v1:0",
		Alias:             "opus",
		Provider:          "Anthropic",
		Description:       "Claude Opus 4.7 - most capable",
		InputPrice:        0.015,
		OutputPrice:       0.075,
		SupportsStreaming: true,
		SupportsConverse:  true,
		MaxTokens:         8192,
	},
	{
		ID:                "anthropic.claude-3-5-haiku-20241022-v1:0",
		Alias:             "haiku-3-5",
		Provider:          "Anthropic",
		Description:       "Claude 3.5 Haiku - previous generation fast model",
		InputPrice:        0.00080,
		OutputPrice:       0.00400,
		SupportsStreaming: true,
		SupportsConverse:  true,
		MaxTokens:         8192,
	},
	{
		ID:                "amazon.titan-text-express-v1",
		Alias:             "titan-text",
		Provider:          "Amazon",
		Description:       "Titan Text Express - general purpose",
		InputPrice:        0.0002,
		OutputPrice:       0.0006,
		SupportsStreaming: true,
		SupportsConverse:  false,
		MaxTokens:         8192,
	},
	{
		ID:                "amazon.titan-text-lite-v1",
		Alias:             "titan-lite",
		Provider:          "Amazon",
		Description:       "Titan Text Lite - fastest Titan model",
		InputPrice:        0.00003,
		OutputPrice:       0.00004,
		SupportsStreaming: true,
		SupportsConverse:  false,
		MaxTokens:         4096,
	},
	{
		ID:                "meta.llama3-70b-instruct-v1:0",
		Alias:             "llama-3-70b",
		Provider:          "Meta",
		Description:       "Llama 3 70B Instruct",
		InputPrice:        0.00265,
		OutputPrice:       0.0035,
		SupportsStreaming: true,
		SupportsConverse:  true,
		MaxTokens:         8192,
	},
	{
		ID:                "meta.llama3-8b-instruct-v1:0",
		Alias:             "llama-3-8b",
		Provider:          "Meta",
		Description:       "Llama 3 8B Instruct - smaller and faster",
		InputPrice:        0.0003,
		OutputPrice:       0.0006,
		SupportsStreaming: true,
		SupportsConverse:  true,
		MaxTokens:         8192,
	},
	{
		ID:                "mistral.mistral-7b-instruct-v0:2",
		Alias:             "mistral-7b",
		Provider:          "Mistral",
		Description:       "Mistral 7B Instruct",
		InputPrice:        0.00015,
		OutputPrice:       0.0002,
		SupportsStreaming: true,
		SupportsConverse:  true,
		MaxTokens:         8192,
	},
	{
		ID:                "mistral.mixtral-8x7b-instruct-v0:1",
		Alias:             "mixtral-8x7b",
		Provider:          "Mistral",
		Description:       "Mixtral 8x7B - mixture of experts",
		InputPrice:        0.00045,
		OutputPrice:       0.0007,
		SupportsStreaming: true,
		SupportsConverse:  true,
		MaxTokens:         8192,
	},
	{
		ID:                "cohere.command-r-v1:0",
		Alias:             "command-r",
		Provider:          "Cohere",
		Description:       "Command R - retrieval-augmented generation focused",
		InputPrice:        0.0005,
		OutputPrice:       0.0015,
		SupportsStreaming: true,
		SupportsConverse:  false,
		MaxTokens:         4096,
	},
	{
		ID:                "cohere.command-r-plus-v1:0",
		Alias:             "command-r-plus",
		Provider:          "Cohere",
		Description:       "Command R+ - stronger reasoning",
		InputPrice:        0.003,
		OutputPrice:       0.015,
		SupportsStreaming: true,
		SupportsConverse:  false,
		MaxTokens:         4096,
	},
}

// Resolve takes either a short alias (e.g. "haiku") or a full model ID
// and returns the matching Model. Returns an error if nothing matches.
func Resolve(aliasOrID string) (*Model, error) {
	for i := range catalog {
		m := &catalog[i]
		if m.Alias == aliasOrID || m.ID == aliasOrID {
			return m, nil
		}
	}
	return nil, fmt.Errorf("unknown model %q - run 'bedrock-cli models list' to see available models", aliasOrID)
}

// All returns every model in the catalog.
func All() []Model {
	result := make([]Model, len(catalog))
	copy(result, catalog)
	return result
}

// ByProvider groups the catalog by provider name.
func ByProvider() map[string][]Model {
	out := make(map[string][]Model)
	for _, m := range catalog {
		out[m.Provider] = append(out[m.Provider], m)
	}
	return out
}
