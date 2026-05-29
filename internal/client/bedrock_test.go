package client

import (
	"encoding/json"
	"testing"
)

func TestParseAnthropicResponse(t *testing.T) {
	body := []byte(`{
		"content": [{"type": "text", "text": "Hello, world!"}],
		"usage": {"input_tokens": 10, "output_tokens": 3}
	}`)

	result, err := parseResponse("anthropic.claude-haiku-4-5-20251001-v1:0", body)
	if err != nil {
		t.Fatalf("parseResponse() error: %v", err)
	}
	if result.Response != "Hello, world!" {
		t.Errorf("Response = %q, want %q", result.Response, "Hello, world!")
	}
	if result.InputTokens != 10 {
		t.Errorf("InputTokens = %d, want 10", result.InputTokens)
	}
	if result.OutputTokens != 3 {
		t.Errorf("OutputTokens = %d, want 3", result.OutputTokens)
	}
}

func TestParseTitanResponse(t *testing.T) {
	body := []byte(`{
		"results": [{"outputText": "Titan says hi"}],
		"inputTextTokenCount": 5
	}`)

	result, err := parseResponse("amazon.titan-text-express-v1", body)
	if err != nil {
		t.Fatalf("parseResponse() error: %v", err)
	}
	if result.Response != "Titan says hi" {
		t.Errorf("Response = %q, want %q", result.Response, "Titan says hi")
	}
	if result.InputTokens != 5 {
		t.Errorf("InputTokens = %d, want 5", result.InputTokens)
	}
}

func TestParseLlamaResponse(t *testing.T) {
	body := []byte(`{
		"generation": "Llama answer",
		"prompt_token_count": 20,
		"generation_token_count": 5
	}`)

	result, err := parseResponse("meta.llama3-70b-instruct-v1:0", body)
	if err != nil {
		t.Fatalf("parseResponse() error: %v", err)
	}
	if result.Response != "Llama answer" {
		t.Errorf("Response = %q, want %q", result.Response, "Llama answer")
	}
	if result.InputTokens != 20 {
		t.Errorf("InputTokens = %d, want 20", result.InputTokens)
	}
	if result.OutputTokens != 5 {
		t.Errorf("OutputTokens = %d, want 5", result.OutputTokens)
	}
}

func TestParseMistralResponse(t *testing.T) {
	body := []byte(`{"outputs": [{"text": "Mistral speaks"}]}`)

	result, err := parseResponse("mistral.mistral-7b-instruct-v0:2", body)
	if err != nil {
		t.Fatalf("parseResponse() error: %v", err)
	}
	if result.Response != "Mistral speaks" {
		t.Errorf("Response = %q, want %q", result.Response, "Mistral speaks")
	}
}

func TestParseCohereResponse(t *testing.T) {
	body := []byte(`{"generations": [{"text": "Cohere output"}]}`)

	result, err := parseResponse("cohere.command-r-v1:0", body)
	if err != nil {
		t.Fatalf("parseResponse() error: %v", err)
	}
	if result.Response != "Cohere output" {
		t.Errorf("Response = %q, want %q", result.Response, "Cohere output")
	}
}

func TestParseAnthropicMultipleContentBlocks(t *testing.T) {
	body := []byte(`{
		"content": [
			{"type": "text", "text": "Part 1. "},
			{"type": "text", "text": "Part 2."}
		],
		"usage": {"input_tokens": 5, "output_tokens": 6}
	}`)

	result, err := parseResponse("anthropic.claude-sonnet-4-6-20250514-v1:0", body)
	if err != nil {
		t.Fatalf("parseResponse() error: %v", err)
	}
	if result.Response != "Part 1. Part 2." {
		t.Errorf("Response = %q, want %q", result.Response, "Part 1. Part 2.")
	}
}

func TestBuildRequestBodyAnthropic(t *testing.T) {
	body, err := buildRequestBody("anthropic.claude-haiku-4-5-20251001-v1:0", "test prompt", 0.7, 1024)
	if err != nil {
		t.Fatalf("buildRequestBody() error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshaling request body: %v", err)
	}

	if _, ok := payload["messages"]; !ok {
		t.Error("Anthropic request body missing 'messages' field")
	}
	if _, ok := payload["anthropic_version"]; !ok {
		t.Error("Anthropic request body missing 'anthropic_version' field")
	}
}

func TestBuildRequestBodyTitan(t *testing.T) {
	body, err := buildRequestBody("amazon.titan-text-express-v1", "hello", 0.5, 512)
	if err != nil {
		t.Fatalf("buildRequestBody() error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshaling request body: %v", err)
	}

	if _, ok := payload["inputText"]; !ok {
		t.Error("Titan request body missing 'inputText' field")
	}
}

func TestBuildRequestBodyLlama(t *testing.T) {
	body, err := buildRequestBody("meta.llama3-70b-instruct-v1:0", "hi", 0.5, 256)
	if err != nil {
		t.Fatalf("buildRequestBody() error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshaling request body: %v", err)
	}

	promptVal, ok := payload["prompt"].(string)
	if !ok {
		t.Error("Llama request body missing 'prompt' field")
	}
	// Llama wraps the prompt in its special tokens.
	if len(promptVal) <= len("hi") {
		t.Error("Llama prompt should be wrapped in special tokens")
	}
}

func TestIsAnthropicModel(t *testing.T) {
	tests := []struct {
		modelID string
		want    bool
	}{
		{"anthropic.claude-haiku-4-5-20251001-v1:0", true},
		{"anthropic.claude-sonnet-4-6-20250514-v1:0", true},
		{"amazon.titan-text-express-v1", false},
		{"meta.llama3-70b-instruct-v1:0", false},
		{"mistral.mistral-7b-instruct-v0:2", false},
	}

	for _, tt := range tests {
		got := isAnthropicModel(tt.modelID)
		if got != tt.want {
			t.Errorf("isAnthropicModel(%q) = %v, want %v", tt.modelID, got, tt.want)
		}
	}
}
