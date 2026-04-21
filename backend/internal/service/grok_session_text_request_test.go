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
	require.Contains(t, message, "[user]: Hello")
	require.Contains(t, message, "[assistant]: Hi.")
	require.Contains(t, message, "[assistant]:\n<tool_calls>")
	require.Contains(t, message, "<tool_name>get_weather</tool_name>")
	require.Contains(t, message, `<parameters>{"city":"Shanghai"}</parameters>`)
	require.Contains(t, message, "[tool result for call_1]:\n{\"temp_c\":26}")
	require.Contains(t, message, "[user]: What should I wear?")
}

func TestBuildGrokSessionTextPayloadFromResponsesRequest_StripsAssistantGeneratedArtifacts(t *testing.T) {
	input, err := json.Marshal([]apicompat.ResponsesInputItem{
		{
			Role:    "user",
			Content: json.RawMessage(`"Summarize the result."`),
		},
		{
			Role: "assistant",
			Content: json.RawMessage(`[
				{"type":"output_text","text":"<thinking>private chain</thinking>\nFinal answer.\n![image](data:image/png;base64,AAAA)"},
				{"type":"output_text","text":"<tool_calls><tool_call><tool_name>search</tool_name><parameters>{\"q\":\"demo\"}</parameters></tool_call></tool_calls>"},
				{"type":"output_text","text":"<function_call><name>search</name><arguments>{\"q\":\"demo\"}</arguments></function_call>"},
				{"type":"output_text","text":"<invoke name=\"search\">{\"q\":\"demo\"}</invoke>"},
				{"type":"output_text","text":"## Sources\n[grok2api-sources]: #\n- [Doc](https://example.com/doc)"},
				{"type":"output_text","text":"Second paragraph."}
			]`),
		},
		{
			Role:    "user",
			Content: json.RawMessage(`"Continue."`),
		},
	})
	require.NoError(t, err)

	payload, err := buildGrokSessionTextPayloadFromResponsesRequest(&apicompat.ResponsesRequest{
		Model: "grok-3",
		Input: input,
	})
	require.NoError(t, err)

	message, ok := payload["message"].(string)
	require.True(t, ok)
	require.Contains(t, message, "[assistant]:\nFinal answer.\n[图片]\nSecond paragraph.")
	require.NotContains(t, message, "<thinking>")
	require.NotContains(t, message, "<tool_calls>")
	require.NotContains(t, message, "<function_call>")
	require.NotContains(t, message, "<invoke ")
	require.NotContains(t, message, "## Sources")
	require.NotContains(t, message, "data:image/png;base64")
}

func TestGrokSessionTextContentFromResponsesContent_PreservesUserSourcesText(t *testing.T) {
	content, err := grokSessionTextContentFromResponsesContent(
		json.RawMessage(`"## Sources\n[grok2api-sources]: #\n- [User supplied](https://example.com)"`),
		"user",
	)
	require.NoError(t, err)
	require.Equal(t, "## Sources\n[grok2api-sources]: #\n- [User supplied](https://example.com)", content.Text)
}

func TestBuildGrokSessionTextPayloadFromResponsesRequest_InjectsToolPrompt(t *testing.T) {
	input, err := json.Marshal([]apicompat.ResponsesInputItem{{
		Role:    "user",
		Content: json.RawMessage(`"What's the weather?"`),
	}})
	require.NoError(t, err)

	payload, err := buildGrokSessionTextPayloadFromResponsesRequest(&apicompat.ResponsesRequest{
		Model: "grok-3",
		Input: input,
		Tools: []apicompat.ResponsesTool{{
			Type:        "function",
			Name:        "get_weather",
			Description: "Look up weather",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}}}`),
		}},
		ToolChoice: json.RawMessage(`"required"`),
	})
	require.NoError(t, err)

	personality, ok := payload["customPersonality"].(string)
	require.True(t, ok)
	require.Contains(t, personality, "AVAILABLE TOOLS:")
	require.Contains(t, personality, "Tool: get_weather")
	require.Contains(t, personality, "WHEN TO CALL: You MUST output a <tool_calls> XML block.")
}

func TestBuildGrokSessionTextPayloadFromResponsesRequest_DeepsearchPreset(t *testing.T) {
	input, err := json.Marshal([]apicompat.ResponsesInputItem{{
		Role:    "user",
		Content: json.RawMessage(`"Research the latest launch."`),
	}})
	require.NoError(t, err)

	payload, err := buildGrokSessionTextPayloadFromResponsesRequest(&apicompat.ResponsesRequest{
		Model:      "grok-3",
		Input:      input,
		Deepsearch: "deeper",
	})
	require.NoError(t, err)
	require.Equal(t, "deeper", payload["deepsearchPreset"])
}

func TestBuildGrokSessionTextPayloadFromResponsesRequest_InvalidDeepsearchPreset(t *testing.T) {
	input, err := json.Marshal([]apicompat.ResponsesInputItem{{
		Role:    "user",
		Content: json.RawMessage(`"Research the latest launch."`),
	}})
	require.NoError(t, err)

	_, err = buildGrokSessionTextPayloadFromResponsesRequest(&apicompat.ResponsesRequest{
		Model:      "grok-3",
		Input:      input,
		Deepsearch: "max",
	})
	require.EqualError(t, err, `unsupported deepsearch value "max"`)
}

func TestGrokSessionTextRequestFromResponsesRequest_CollectsAttachments(t *testing.T) {
	responsesReq, err := apicompat.ChatCompletionsToResponses(&apicompat.ChatCompletionsRequest{
		Model: "grok-3-fast",
		Messages: []apicompat.ChatMessage{
			{
				Role:    "system",
				Content: json.RawMessage(`"Describe the image precisely."`),
			},
			{
				Role: "user",
				Content: json.RawMessage(`[
					{"type":"text","text":"What is in this image and audio clip?"},
					{"type":"image_url","image_url":{"url":"https://example.com/cat.png"}},
					{"type":"input_audio","input_audio":{"data":"data:audio/wav;base64,abc123","filename":"clip.wav","mime_type":"audio/wav"}},
					{"type":"file","file":{"file_data":"data:text/plain;base64,Zm9v","filename":"notes.txt","mime_type":"text/plain"}}
				]`),
			},
		},
	})
	require.NoError(t, err)

	request, err := grokSessionTextRequestFromResponsesRequest(responsesReq)
	require.NoError(t, err)
	require.Equal(t, "fast", request.ModeID)
	require.Equal(t, "Describe the image precisely.", request.SystemPrompt)
	require.Equal(t, "What is in this image and audio clip?", request.Message)
	require.Len(t, request.AttachmentInputs, 3)
	require.Equal(t, "https://example.com/cat.png", request.AttachmentInputs[0].Source)
	require.Equal(t, "data:audio/wav;base64,abc123", request.AttachmentInputs[1].Source)
	require.Equal(t, "clip.wav", request.AttachmentInputs[1].FileName)
	require.Equal(t, "data:text/plain;base64,Zm9v", request.AttachmentInputs[2].Source)
	require.Equal(t, "notes.txt", request.AttachmentInputs[2].FileName)
}

func TestGrokSessionTextRequestFromResponsesRequest_CollectsRawBase64Audio(t *testing.T) {
	input, err := json.Marshal([]apicompat.ResponsesInputItem{{
		Role: "user",
		Content: json.RawMessage(`[
			{"type":"input_text","text":"Listen carefully."},
			{"type":"input_audio","input_audio":{"data":"UklGRngAAABXQVZF","format":"wav"}}
		]`),
	}})
	require.NoError(t, err)

	request, err := grokSessionTextRequestFromResponsesRequest(&apicompat.ResponsesRequest{
		Model: "grok-3",
		Input: input,
	})
	require.NoError(t, err)
	require.Equal(t, "Listen carefully.", request.Message)
	require.Len(t, request.AttachmentInputs, 1)
	require.Equal(t, "UklGRngAAABXQVZF", request.AttachmentInputs[0].Base64)
	require.Equal(t, "audio/wav", request.AttachmentInputs[0].MimeType)
	require.Equal(t, "upload.wav", request.AttachmentInputs[0].FileName)
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
