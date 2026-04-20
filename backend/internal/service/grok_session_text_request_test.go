package service

import (
	"encoding/json"
	"testing"

	"github.com/senran-N/sub2api/internal/pkg/apicompat"
	"github.com/stretchr/testify/require"
)

func TestBuildGrokSessionTextPayloadFromResponsesRequest_SingleUserMessage(t *testing.T) {
	input, err := json.Marshal([]apicompat.ResponsesInputItem{
		{
			Role:    "user",
			Content: json.RawMessage(`"hello grok"`),
		},
	})
	require.NoError(t, err)

	payload, err := buildGrokSessionTextPayloadFromResponsesRequest(&apicompat.ResponsesRequest{
		Model: "grok-4.20-fast",
		Input: input,
	})
	require.NoError(t, err)

	require.Equal(t, "fast", payload["modeId"])
	require.Equal(t, "hello grok", payload["message"])
	disableSearch, ok := payload["disableSearch"].(bool)
	require.True(t, ok)
	require.False(t, disableSearch)
	sendFinalMetadata, ok := payload["sendFinalMetadata"].(bool)
	require.True(t, ok)
	require.True(t, sendFinalMetadata)
	_, hasCustomPersonality := payload["customPersonality"]
	require.False(t, hasCustomPersonality)
}

func TestBuildGrokSessionTextPayloadFromResponsesRequest_FlattensSystemAndHistory(t *testing.T) {
	input, err := json.Marshal([]apicompat.ResponsesInputItem{
		{
			Role:    "system",
			Content: json.RawMessage(`"Be terse."`),
		},
		{
			Role:    "user",
			Content: json.RawMessage(`"Hello"`),
		},
		{
			Role:    "assistant",
			Content: json.RawMessage(`[{"type":"output_text","text":"Hi."}]`),
		},
		{
			Type:      "function_call",
			Name:      "get_weather",
			Arguments: `{"city":"Shanghai"}`,
		},
		{
			Type:   "function_call_output",
			CallID: "call_1",
			Output: `{"temp_c":26}`,
		},
		{
			Role:    "user",
			Content: json.RawMessage(`"What should I wear?"`),
		},
	})
	require.NoError(t, err)

	payload, err := buildGrokSessionTextPayloadFromResponsesRequest(&apicompat.ResponsesRequest{
		Model: "grok-3",
		Input: input,
	})
	require.NoError(t, err)

	require.Equal(t, "auto", payload["modeId"])
	require.Equal(t, "Be terse.", payload["customPersonality"])
	message, ok := payload["message"].(string)
	require.True(t, ok)
	require.Contains(t, message, "User: Hello")
	require.Contains(t, message, "Assistant: Hi.")
	require.Contains(t, message, `Assistant tool call get_weather: {"city":"Shanghai"}`)
	require.Contains(t, message, `Tool result (call_1): {"temp_c":26}`)
	require.Contains(t, message, "User: What should I wear?")
}

func TestGrokSessionTextRequestFromResponsesRequest_CollectsImageInputs(t *testing.T) {
	responsesReq, err := apicompat.ChatCompletionsToResponses(&apicompat.ChatCompletionsRequest{
		Model: "grok-3-fast",
		Messages: []apicompat.ChatMessage{
			{
				Role:    "system",
				Content: json.RawMessage(`"Describe the image precisely."`),
			},
			{
				Role:    "user",
				Content: json.RawMessage(`[{"type":"text","text":"What is in this image?"},{"type":"image_url","image_url":{"url":"https://example.com/cat.png"}}]`),
			},
		},
	})
	require.NoError(t, err)

	request, err := grokSessionTextRequestFromResponsesRequest(responsesReq)
	require.NoError(t, err)
	require.Equal(t, "fast", request.ModeID)
	require.Equal(t, "Describe the image precisely.", request.SystemPrompt)
	require.Equal(t, "What is in this image?", request.Message)
	require.Len(t, request.ImageInputs, 1)
	require.Equal(t, "https://example.com/cat.png", request.ImageInputs[0].Source)
}

func TestBuildGrokSessionTextPayloadFromAnthropicRequest_UsesSystemAsCustomPersonality(t *testing.T) {
	payload, err := buildGrokSessionTextPayloadFromAnthropicRequest(&apicompat.AnthropicRequest{
		Model:  "grok-4.20-expert",
		System: json.RawMessage(`"Answer like an infra lead."`),
		Messages: []apicompat.AnthropicMessage{
			{
				Role:    "user",
				Content: json.RawMessage(`"Summarize the incident."`),
			},
		},
	})
	require.NoError(t, err)

	require.Equal(t, "expert", payload["modeId"])
	require.Equal(t, "Answer like an infra lead.", payload["customPersonality"])
	require.Equal(t, "Summarize the incident.", payload["message"])
}

func TestCreateGrokSessionTestPayload_UsesSharedTextBuilder(t *testing.T) {
	payload, err := createGrokSessionTestPayload("grok-3", "")
	require.NoError(t, err)

	require.Equal(t, defaultTestPrompt, payload["message"])
	require.Equal(t, "auto", payload["modeId"])
	disableSearch, ok := payload["disableSearch"].(bool)
	require.True(t, ok)
	require.True(t, disableSearch)
	sendFinalMetadata, ok := payload["sendFinalMetadata"].(bool)
	require.True(t, ok)
	require.False(t, sendFinalMetadata)
	_, hasModelName := payload["modelName"]
	require.False(t, hasModelName)
}
