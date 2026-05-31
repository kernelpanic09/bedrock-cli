package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	brtypes "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// InvokeResult holds the response from a model invocation.
type InvokeResult struct {
	Response     string
	InputTokens  int
	OutputTokens int
}

// Client wraps the AWS Bedrock SDK clients.
type Client struct {
	runtime      *bedrockruntime.Client
	control      *bedrock.Client
	agentRuntime *bedrockagentruntime.Client
	region       string
}

// New creates a new Bedrock client for the given region.
// It uses the standard AWS credential chain (env vars, ~/.aws/credentials, IAM role, etc.).
func New(ctx context.Context, region string) (*Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}
	return &Client{
		runtime:      bedrockruntime.NewFromConfig(cfg),
		control:      bedrock.NewFromConfig(cfg),
		agentRuntime: bedrockagentruntime.NewFromConfig(cfg),
		region:       region,
	}, nil
}

// Invoke calls the model and returns the full response text and token counts.
// It uses InvokeModel (non-streaming). For streaming use InvokeStream.
func (c *Client) Invoke(ctx context.Context, modelID, prompt string, temperature float64, maxTokens int) (*InvokeResult, error) {
	body, err := buildRequestBody(modelID, prompt, temperature, maxTokens)
	if err != nil {
		return nil, err
	}

	resp, err := c.runtime.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelID),
		Body:        body,
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	})
	if err != nil {
		return nil, fmt.Errorf("invoking model %s: %w", modelID, err)
	}

	return parseResponse(modelID, resp.Body)
}

// InvokeStream calls the model with streaming and writes tokens to the callback
// as they arrive. Returns token usage after the stream is exhausted.
func (c *Client) InvokeStream(ctx context.Context, modelID, prompt string, temperature float64, maxTokens int, onToken func(string)) (*InvokeResult, error) {
	// Use the Converse streaming API for Anthropic models - it gives us a
	// cleaner interface and works with cross-region inference profiles.
	if isAnthropicModel(modelID) {
		return c.invokeConverseStream(ctx, modelID, prompt, temperature, maxTokens, onToken)
	}
	return c.invokeModelStream(ctx, modelID, prompt, temperature, maxTokens, onToken)
}

// invokeConverseStream uses the ConverseStream API (supported by Anthropic + several others).
func (c *Client) invokeConverseStream(ctx context.Context, modelID, prompt string, temperature float64, maxTokens int, onToken func(string)) (*InvokeResult, error) {
	maxTok := int32(maxTokens)
	temp := float32(temperature)

	input := &bedrockruntime.ConverseStreamInput{
		ModelId: aws.String(modelID),
		Messages: []brtypes.Message{
			{
				Role: brtypes.ConversationRoleUser,
				Content: []brtypes.ContentBlock{
					&brtypes.ContentBlockMemberText{Value: prompt},
				},
			},
		},
		InferenceConfig: &brtypes.InferenceConfiguration{
			MaxTokens:   &maxTok,
			Temperature: &temp,
		},
	}

	stream, err := c.runtime.ConverseStream(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("starting converse stream for %s: %w", modelID, err)
	}

	return consumeConverseStream(stream, onToken)
}

// invokeModelStream uses InvokeModelWithResponseStream for models that don't
// support the Converse API (e.g. older Titan, Cohere).
func (c *Client) invokeModelStream(ctx context.Context, modelID, prompt string, temperature float64, maxTokens int, onToken func(string)) (*InvokeResult, error) {
	body, err := buildRequestBody(modelID, prompt, temperature, maxTokens)
	if err != nil {
		return nil, err
	}

	resp, err := c.runtime.InvokeModelWithResponseStream(ctx, &bedrockruntime.InvokeModelWithResponseStreamInput{
		ModelId:     aws.String(modelID),
		Body:        body,
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	})
	if err != nil {
		return nil, fmt.Errorf("starting stream for %s: %w", modelID, err)
	}

	return consumeModelStream(modelID, resp.GetStream(), onToken)
}

// ListFoundationModels returns all foundation models available in the region.
func (c *Client) ListFoundationModels(ctx context.Context) ([]FoundationModel, error) {
	resp, err := c.control.ListFoundationModels(ctx, &bedrock.ListFoundationModelsInput{})
	if err != nil {
		return nil, fmt.Errorf("listing foundation models: %w", err)
	}

	var models []FoundationModel
	for _, m := range resp.ModelSummaries {
		fm := FoundationModel{
			ModelID:      aws.ToString(m.ModelId),
			ModelName:    aws.ToString(m.ModelName),
			ProviderName: aws.ToString(m.ProviderName),
		}
		models = append(models, fm)
	}
	return models, nil
}

// GetFoundationModel returns details for a single model.
func (c *Client) GetFoundationModel(ctx context.Context, modelID string) (*FoundationModelDetail, error) {
	resp, err := c.control.GetFoundationModel(ctx, &bedrock.GetFoundationModelInput{
		ModelIdentifier: aws.String(modelID),
	})
	if err != nil {
		return nil, fmt.Errorf("getting model %s: %w", modelID, err)
	}

	m := resp.ModelDetails
	detail := &FoundationModelDetail{
		ModelID:      aws.ToString(m.ModelId),
		ModelName:    aws.ToString(m.ModelName),
		ProviderName: aws.ToString(m.ProviderName),
	}
	for _, mod := range m.InputModalities {
		detail.InputModalities = append(detail.InputModalities, string(mod))
	}
	for _, mod := range m.OutputModalities {
		detail.OutputModalities = append(detail.OutputModalities, string(mod))
	}
	return detail, nil
}

// ListKnowledgeBases returns all knowledge bases in the account/region via the agent runtime.
func (c *Client) ListKnowledgeBases(ctx context.Context) ([]KnowledgeBase, error) {
	// The agent client uses a different service endpoint than the runtime.
	// We delegate to a wrapper to keep this file manageable.
	return listKnowledgeBases(ctx, c.agentRuntime)
}

// QueryKnowledgeBase sends a query to a knowledge base and returns the results.
func (c *Client) QueryKnowledgeBase(ctx context.Context, kbID, query string, maxResults int) ([]KBResult, error) {
	return queryKnowledgeBase(ctx, c.agentRuntime, kbID, query, maxResults)
}

// FoundationModel is a simplified view of a Bedrock foundation model for list output.
type FoundationModel struct {
	ModelID      string
	ModelName    string
	ProviderName string
}

// FoundationModelDetail holds richer information about a single model.
type FoundationModelDetail struct {
	ModelID          string
	ModelName        string
	ProviderName     string
	InputModalities  []string
	OutputModalities []string
}

// KnowledgeBase holds basic info about a Bedrock Knowledge Base.
type KnowledgeBase struct {
	ID          string
	Name        string
	Description string
	Status      string
	VectorStore string
}

// KBResult is a single result chunk from a knowledge base query.
type KBResult struct {
	Content string
	Score   float64
	Source  string
}

// buildRequestBody constructs the JSON request payload for a model.
// Different providers use different request schemas.
func buildRequestBody(modelID, prompt string, temperature float64, maxTokens int) ([]byte, error) {
	var payload any

	switch {
	case isAnthropicModel(modelID):
		payload = map[string]any{
			"anthropic_version": "bedrock-2023-05-31",
			"max_tokens":        maxTokens,
			"temperature":       temperature,
			"messages": []map[string]any{
				{"role": "user", "content": prompt},
			},
		}
	case strings.HasPrefix(modelID, "amazon.titan"):
		payload = map[string]any{
			"inputText": prompt,
			"textGenerationConfig": map[string]any{
				"maxTokenCount": maxTokens,
				"temperature":   temperature,
			},
		}
	case strings.HasPrefix(modelID, "meta.llama"):
		payload = map[string]any{
			"prompt":      fmt.Sprintf("<|begin_of_text|><|start_header_id|>user<|end_header_id|>\n\n%s<|eot_id|><|start_header_id|>assistant<|end_header_id|>\n\n", prompt),
			"max_gen_len": maxTokens,
			"temperature": temperature,
		}
	case strings.HasPrefix(modelID, "mistral"):
		payload = map[string]any{
			"prompt":      fmt.Sprintf("<s>[INST] %s [/INST]", prompt),
			"max_tokens":  maxTokens,
			"temperature": temperature,
		}
	case strings.HasPrefix(modelID, "cohere"):
		payload = map[string]any{
			"prompt":      prompt,
			"max_tokens":  maxTokens,
			"temperature": temperature,
		}
	default:
		// Best-effort fallback for unknown providers.
		payload = map[string]any{
			"prompt":      prompt,
			"max_tokens":  maxTokens,
			"temperature": temperature,
		}
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body for %s: %w", modelID, err)
	}
	return data, nil
}

// parseResponse extracts the text and token counts from an InvokeModel response.
func parseResponse(modelID string, body []byte) (*InvokeResult, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing response from %s: %w", modelID, err)
	}

	result := &InvokeResult{}

	switch {
	case isAnthropicModel(modelID):
		var resp struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
			Usage struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("parsing Anthropic response: %w", err)
		}
		for _, c := range resp.Content {
			result.Response += c.Text
		}
		result.InputTokens = resp.Usage.InputTokens
		result.OutputTokens = resp.Usage.OutputTokens

	case strings.HasPrefix(modelID, "amazon.titan"):
		var resp struct {
			Results []struct {
				OutputText string `json:"outputText"`
			} `json:"results"`
			InputTextTokenCount int `json:"inputTextTokenCount"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("parsing Titan response: %w", err)
		}
		for _, r := range resp.Results {
			result.Response += r.OutputText
		}
		result.InputTokens = resp.InputTextTokenCount

	case strings.HasPrefix(modelID, "meta.llama"):
		var resp struct {
			Generation           string `json:"generation"`
			PromptTokenCount     int    `json:"prompt_token_count"`
			GenerationTokenCount int    `json:"generation_token_count"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("parsing Llama response: %w", err)
		}
		result.Response = resp.Generation
		result.InputTokens = resp.PromptTokenCount
		result.OutputTokens = resp.GenerationTokenCount

	case strings.HasPrefix(modelID, "mistral"):
		var resp struct {
			Outputs []struct {
				Text string `json:"text"`
			} `json:"outputs"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("parsing Mistral response: %w", err)
		}
		for _, o := range resp.Outputs {
			result.Response += o.Text
		}

	case strings.HasPrefix(modelID, "cohere"):
		var resp struct {
			Generations []struct {
				Text string `json:"text"`
			} `json:"generations"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("parsing Cohere response: %w", err)
		}
		for _, g := range resp.Generations {
			result.Response += g.Text
		}

	default:
		// Try a generic "text" or "completion" field.
		if v, ok := raw["text"]; ok {
			json.Unmarshal(v, &result.Response)
		} else if v, ok := raw["completion"]; ok {
			json.Unmarshal(v, &result.Response)
		} else {
			result.Response = string(body)
		}
	}

	return result, nil
}

func isAnthropicModel(modelID string) bool {
	return strings.HasPrefix(modelID, "anthropic.")
}
