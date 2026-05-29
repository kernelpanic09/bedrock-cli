package client

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	brtypes "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// consumeConverseStream drains a ConverseStream, calling onToken for each text
// delta and returning the final token usage.
func consumeConverseStream(stream *bedrockruntime.ConverseStreamOutput, onToken func(string)) (*InvokeResult, error) {
	result := &InvokeResult{}
	evtStream := stream.GetStream()
	defer evtStream.Close()

	for event := range evtStream.Events() {
		switch v := event.(type) {
		case *brtypes.ConverseStreamOutputMemberContentBlockDelta:
			if delta, ok := v.Value.Delta.(*brtypes.ContentBlockDeltaMemberText); ok {
				onToken(delta.Value)
				result.Response += delta.Value
			}
		case *brtypes.ConverseStreamOutputMemberMetadata:
			if v.Value.Usage != nil {
				result.InputTokens = int(aws_int32(v.Value.Usage.InputTokens))
				result.OutputTokens = int(aws_int32(v.Value.Usage.OutputTokens))
			}
		}
	}

	if err := evtStream.Err(); err != nil {
		return nil, fmt.Errorf("converse stream error: %w", err)
	}
	return result, nil
}

// consumeModelStream drains an InvokeModelWithResponseStream event stream,
// calling onToken for each text chunk. It handles the different chunk formats
// across providers.
func consumeModelStream(modelID string, stream *bedrockruntime.InvokeModelWithResponseStreamEventStream, onToken func(string)) (*InvokeResult, error) {
	result := &InvokeResult{}
	defer stream.Close()

	for event := range stream.Events() {
		switch v := event.(type) {
		case *brtypes.ResponseStreamMemberChunk:
			chunk, err := parseStreamChunk(modelID, v.Value.Bytes)
			if err != nil {
				// Non-fatal: skip malformed chunks but keep going.
				continue
			}
			if chunk.text != "" {
				onToken(chunk.text)
				result.Response += chunk.text
			}
			if chunk.inputTokens > 0 {
				result.InputTokens = chunk.inputTokens
			}
			if chunk.outputTokens > 0 {
				result.OutputTokens = chunk.outputTokens
			}
		}
	}

	if err := stream.Err(); err != nil {
		return nil, fmt.Errorf("model stream error: %w", err)
	}
	return result, nil
}

type parsedChunk struct {
	text         string
	inputTokens  int
	outputTokens int
}

// parseStreamChunk handles the per-provider chunk format variations.
func parseStreamChunk(modelID string, data []byte) (*parsedChunk, error) {
	chunk := &parsedChunk{}

	switch {
	case isAnthropicModel(modelID):
		var evt struct {
			Type  string `json:"type"`
			Delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"delta"`
			Message struct {
				Usage struct {
					InputTokens  int `json:"input_tokens"`
					OutputTokens int `json:"output_tokens"`
				} `json:"usage"`
			} `json:"message"`
			Usage struct {
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
		}
		if err := json.Unmarshal(data, &evt); err != nil {
			return nil, err
		}
		if evt.Type == "content_block_delta" && evt.Delta.Type == "text_delta" {
			chunk.text = evt.Delta.Text
		}
		if evt.Type == "message_start" {
			chunk.inputTokens = evt.Message.Usage.InputTokens
		}
		if evt.Type == "message_delta" {
			chunk.outputTokens = evt.Usage.OutputTokens
		}

	case strings.HasPrefix(modelID, "amazon.titan"):
		var evt struct {
			OutputText string `json:"outputText"`
		}
		if err := json.Unmarshal(data, &evt); err != nil {
			return nil, err
		}
		chunk.text = evt.OutputText

	case strings.HasPrefix(modelID, "meta.llama"):
		var evt struct {
			Generation string `json:"generation"`
		}
		if err := json.Unmarshal(data, &evt); err != nil {
			return nil, err
		}
		chunk.text = evt.Generation

	case strings.HasPrefix(modelID, "mistral"):
		var evt struct {
			Outputs []struct {
				Text string `json:"text"`
			} `json:"outputs"`
		}
		if err := json.Unmarshal(data, &evt); err != nil {
			return nil, err
		}
		for _, o := range evt.Outputs {
			chunk.text += o.Text
		}

	default:
		// Attempt generic text extraction.
		var generic map[string]json.RawMessage
		if err := json.Unmarshal(data, &generic); err != nil {
			return nil, err
		}
		if v, ok := generic["text"]; ok {
			json.Unmarshal(v, &chunk.text)
		}
	}

	return chunk, nil
}

// aws_int32 safely dereferences an *int32 pointer.
func aws_int32(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}
