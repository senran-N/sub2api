package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/apicompat"
	"github.com/stretchr/testify/require"
)

func TestRelayGrokSessionResponses_StreamEmitsResponsesEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	upstream := strings.Join([]string{
		`data: {"result":{"response":{"token":"think ","isThinking":true}}}`,
		`{"result":{"response":{"token":"done","messageTag":"final"}}}`,
		`data: {"result":{"response":{"isSoftStop":true,"finalMetadata":{"stop_reason":"end_turn"}}}}`,
	}, "\n")

	err := relayGrokSessionResponses(c, strings.NewReader(upstream), "grok-3", true)
	require.NoError(t, err)
	require.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))

	events := decodeResponsesSSEEvents(t, rec.Body.String())
	require.Len(t, events, 6)
	require.Equal(t, "response.created", events[0].Type)
	require.Equal(t, "response.output_item.added", events[1].Type)
	require.Equal(t, "response.reasoning_summary_text.delta", events[2].Type)
	require.Equal(t, "response.output_item.added", events[3].Type)
	require.Equal(t, "response.output_text.delta", events[4].Type)
	require.Equal(t, "response.completed", events[5].Type)
	require.Equal(t, "done", events[4].Delta)
	require.NotNil(t, events[5].Response)
	require.Len(t, events[5].Response.Output, 2)
	require.Equal(t, "reasoning", events[5].Response.Output[0].Type)
	require.Equal(t, "message", events[5].Response.Output[1].Type)
}

func TestRelayGrokSessionResponses_StreamAggregatesReasoningTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	upstream := strings.Join([]string{
		`data: {"result":{"response":{"token":"Analyzing","isThinking":true}}}`,
		`data: {"result":{"response":{"token":" the","isThinking":true}}}`,
		`data: {"result":{"response":{"token":" request.","isThinking":true}}}`,
		`data: {"result":{"response":{"token":"done","messageTag":"final"}}}`,
		`data: {"result":{"response":{"isSoftStop":true,"finalMetadata":{"stop_reason":"end_turn"}}}}`,
	}, "\n")

	err := relayGrokSessionResponses(c, strings.NewReader(upstream), "grok-3", true)
	require.NoError(t, err)

	events := decodeResponsesSSEEvents(t, rec.Body.String())
	require.Len(t, events, 6)
	require.Equal(t, "response.reasoning_summary_text.delta", events[2].Type)
	require.Equal(t, "Analyzing the request.", events[2].Delta)
	require.Equal(t, "response.output_text.delta", events[4].Type)
	require.Equal(t, "done", events[4].Delta)
}

func TestRelayGrokSessionChatCompletions_StreamAggregatesReasoningContent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	upstream := strings.Join([]string{
		`data: {"result":{"response":{"token":"Analyzing","isThinking":true}}}`,
		`data: {"result":{"response":{"token":" the","isThinking":true}}}`,
		`data: {"result":{"response":{"token":" request.","isThinking":true}}}`,
		`data: {"result":{"response":{"token":"done","messageTag":"final"}}}`,
		`data: {"result":{"response":{"isSoftStop":true,"finalMetadata":{"stop_reason":"end_turn"}}}}`,
	}, "\n")

	err := relayGrokSessionChatCompletions(c, strings.NewReader(upstream), "grok-3", true, false)
	require.NoError(t, err)
	require.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))

	chunks := decodeChatCompletionsSSEChunks(t, rec.Body.String())
	require.Len(t, chunks, 4)
	require.Equal(t, "assistant", chunks[0].Choices[0].Delta.Role)
	require.NotNil(t, chunks[1].Choices[0].Delta.ReasoningContent)
	require.Equal(t, "Analyzing the request.", *chunks[1].Choices[0].Delta.ReasoningContent)
	require.NotNil(t, chunks[2].Choices[0].Delta.Content)
	require.Equal(t, "done", *chunks[2].Choices[0].Delta.Content)
	require.NotNil(t, chunks[3].Choices[0].FinishReason)
	require.Equal(t, "stop", *chunks[3].Choices[0].FinishReason)
}

func TestRelayGrokSessionResponses_BufferedBuildsResponsesResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	upstream := strings.Join([]string{
		`{"result":{"response":{"token":"step ","isThinking":true}}}`,
		`{"result":{"response":{"token":"answer","messageTag":"final"}}}`,
		`{"result":{"response":{"finalMetadata":{"stop_reason":"end_turn"}}}}`,
	}, "\n")

	err := relayGrokSessionResponses(c, strings.NewReader(upstream), "grok-3", false)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	var response apicompat.ResponsesResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Equal(t, "response", response.Object)
	require.Equal(t, "completed", response.Status)
	require.Len(t, response.Output, 2)
	require.Equal(t, "reasoning", response.Output[0].Type)
	require.Equal(t, "step ", response.Output[0].Summary[0].Text)
	require.Equal(t, "message", response.Output[1].Type)
	require.Equal(t, "answer", response.Output[1].Content[0].Text)
}

func TestRelayGrokSessionResponses_BufferedCleansRenderCitationTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	upstream := strings.Join([]string{
		`data: {"result":{"response":{"cardAttachment":{"jsonData":"{\"id\":\"card_1\",\"url\":\"https://example.com/article\"}"}}}}`,
		`data: {"result":{"response":{"token":"Answer<grok:render card_id=\"card_1\" card_type=\"citation\" type=\"render_inline_citation\"></grok:render>","messageTag":"final"}}}`,
		`data: {"result":{"response":{"finalMetadata":{"stop_reason":"end_turn"}}}}`,
	}, "\n")

	err := relayGrokSessionResponses(c, strings.NewReader(upstream), "grok-3", false)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	var response apicompat.ResponsesResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Len(t, response.Output, 1)
	require.Equal(t, "message", response.Output[0].Type)
	require.Equal(t, "Answer [[1]](https://example.com/article)", response.Output[0].Content[0].Text)
}

func decodeResponsesSSEEvents(t *testing.T, raw string) []apicompat.ResponsesStreamEvent {
	t.Helper()

	lines := strings.Split(raw, "\n")
	events := make([]apicompat.ResponsesStreamEvent, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
		if payload == "" {
			continue
		}

		var event apicompat.ResponsesStreamEvent
		require.NoError(t, json.Unmarshal([]byte(payload), &event))
		events = append(events, event)
	}
	return events
}

func decodeChatCompletionsSSEChunks(t *testing.T, raw string) []apicompat.ChatCompletionsChunk {
	t.Helper()

	lines := strings.Split(raw, "\n")
	chunks := make([]apicompat.ChatCompletionsChunk, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
		if payload == "" || payload == "[DONE]" {
			continue
		}

		var chunk apicompat.ChatCompletionsChunk
		require.NoError(t, json.Unmarshal([]byte(payload), &chunk))
		chunks = append(chunks, chunk)
	}
	return chunks
}
