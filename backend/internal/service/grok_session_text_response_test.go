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

	err := relayGrokSessionResponses(c, strings.NewReader(upstream), "grok-3", true, nil)
	require.NoError(t, err)
	require.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))

	events := decodeResponsesSSEEvents(t, rec.Body.String())
	require.Len(t, events, 14)
	require.Equal(t, "response.created", events[0].Type)
	require.Equal(t, "response.output_item.added", events[1].Type)
	require.Equal(t, "response.reasoning_summary_part.added", events[2].Type)
	require.NotNil(t, events[2].Part)
	require.Equal(t, "summary_text", events[2].Part.Type)
	require.Equal(t, "response.reasoning_summary_text.delta", events[3].Type)
	require.Equal(t, "response.reasoning_summary_text.done", events[4].Type)
	require.Equal(t, "response.reasoning_summary_part.done", events[5].Type)
	require.Equal(t, "response.output_item.done", events[6].Type)
	require.Equal(t, "response.output_item.added", events[7].Type)
	require.Equal(t, "response.content_part.added", events[8].Type)
	require.NotNil(t, events[8].Part)
	require.Equal(t, "output_text", events[8].Part.Type)
	require.Equal(t, "response.output_text.delta", events[9].Type)
	require.Equal(t, "response.output_text.done", events[10].Type)
	require.Equal(t, "response.content_part.done", events[11].Type)
	require.Equal(t, "response.output_item.done", events[12].Type)
	require.Equal(t, "response.completed", events[13].Type)
	require.Equal(t, "done", events[9].Delta)
	require.Equal(t, "think ", events[4].Text)
	require.NotNil(t, events[6].Item)
	require.Equal(t, "reasoning", events[6].Item.Type)
	require.NotNil(t, events[12].Item)
	require.Equal(t, "message", events[12].Item.Type)
	require.Equal(t, "done", events[10].Text)
	require.NotNil(t, events[11].Part)
	require.Equal(t, "done", events[11].Part.Text)
	require.NotNil(t, events[13].Response)
	require.Len(t, events[13].Response.Output, 2)
	require.Equal(t, "reasoning", events[13].Response.Output[0].Type)
	require.Equal(t, "message", events[13].Response.Output[1].Type)
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

	err := relayGrokSessionResponses(c, strings.NewReader(upstream), "grok-3", true, nil)
	require.NoError(t, err)

	events := decodeResponsesSSEEvents(t, rec.Body.String())
	require.Len(t, events, 16)
	require.Equal(t, "response.reasoning_summary_part.added", events[2].Type)
	require.Equal(t, "response.reasoning_summary_text.delta", events[3].Type)
	require.Equal(t, "Analyzing", events[3].Delta)
	require.Equal(t, "response.reasoning_summary_text.delta", events[4].Type)
	require.Equal(t, " the", events[4].Delta)
	require.Equal(t, "response.reasoning_summary_text.delta", events[5].Type)
	require.Equal(t, " request.", events[5].Delta)
	require.Equal(t, "response.reasoning_summary_text.done", events[6].Type)
	require.Equal(t, "Analyzing the request.", events[6].Text)
	require.Equal(t, "response.reasoning_summary_part.done", events[7].Type)
	require.Equal(t, "response.output_item.done", events[8].Type)
	require.Equal(t, "response.content_part.added", events[10].Type)
	require.Equal(t, "response.output_text.delta", events[11].Type)
	require.Equal(t, "done", events[11].Delta)
	require.Equal(t, "response.output_text.done", events[12].Type)
	require.Equal(t, "done", events[12].Text)
	require.Equal(t, "response.content_part.done", events[13].Type)
	require.Equal(t, "response.output_item.done", events[14].Type)
	require.Equal(t, "response.completed", events[15].Type)
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

	err := relayGrokSessionChatCompletions(c, strings.NewReader(upstream), "grok-3", true, false, nil)
	require.NoError(t, err)
	require.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))

	chunks := decodeChatCompletionsSSEChunks(t, rec.Body.String())
	require.Len(t, chunks, 6)
	require.Equal(t, "assistant", chunks[0].Choices[0].Delta.Role)
	require.NotNil(t, chunks[1].Choices[0].Delta.ReasoningContent)
	require.Equal(t, "Analyzing", *chunks[1].Choices[0].Delta.ReasoningContent)
	require.NotNil(t, chunks[2].Choices[0].Delta.ReasoningContent)
	require.Equal(t, " the", *chunks[2].Choices[0].Delta.ReasoningContent)
	require.NotNil(t, chunks[3].Choices[0].Delta.ReasoningContent)
	require.Equal(t, " request.", *chunks[3].Choices[0].Delta.ReasoningContent)
	require.NotNil(t, chunks[4].Choices[0].Delta.Content)
	require.Equal(t, "done", *chunks[4].Choices[0].Delta.Content)
	require.NotNil(t, chunks[5].Choices[0].FinishReason)
	require.Equal(t, "stop", *chunks[5].Choices[0].FinishReason)
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

	err := relayGrokSessionResponses(c, strings.NewReader(upstream), "grok-3", false, nil)
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
		`data: {"result":{"response":{"webSearchResults":{"results":[{"url":"https://example.com/article","title":"Example Article"}]}}}}`,
		`data: {"result":{"response":{"cardAttachment":{"jsonData":"{\"id\":\"card_1\",\"url\":\"https://example.com/article\"}"}}}}`,
		`data: {"result":{"response":{"token":"Answer<grok:render card_id=\"card_1\" card_type=\"citation\" type=\"render_inline_citation\"></grok:render>","messageTag":"final"}}}`,
		`data: {"result":{"response":{"finalMetadata":{"stop_reason":"end_turn"}}}}`,
	}, "\n")

	err := relayGrokSessionResponses(c, strings.NewReader(upstream), "grok-3", false, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	var response apicompat.ResponsesResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Len(t, response.Output, 1)
	require.Equal(t, "message", response.Output[0].Type)
	require.Equal(
		t,
		"Answer [[1]](https://example.com/article)\n\n## Sources\n[grok2api-sources]: #\n- [Example Article](https://example.com/article)\n",
		response.Output[0].Content[0].Text,
	)
	require.Len(t, response.Output[0].Content[0].Annotations, 1)
	require.Equal(t, "url_citation", response.Output[0].Content[0].Annotations[0].Type)
	require.Equal(t, "https://example.com/article", response.Output[0].Content[0].Annotations[0].URL)
	require.Equal(t, "Example Article", response.Output[0].Content[0].Annotations[0].Title)
	require.Equal(t, 6, response.Output[0].Content[0].Annotations[0].StartIndex)
	require.Equal(t, 41, response.Output[0].Content[0].Annotations[0].EndIndex)
	require.Len(t, response.Output[0].SearchSources, 1)
	require.Equal(t, "https://example.com/article", response.Output[0].SearchSources[0].URL)
	require.Equal(t, "Example Article", response.Output[0].SearchSources[0].Title)
	require.Equal(t, "web", response.Output[0].SearchSources[0].Type)
}

func TestRelayGrokSessionResponses_BufferedAppendsSourcesSuffix(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	upstream := strings.Join([]string{
		`data: {"result":{"response":{"token":"Answer","messageTag":"final"}}}`,
		`data: {"result":{"response":{"webSearchResults":{"results":[{"url":"https://example.com/article","title":"Example [Doc]"}]}}}}`,
		`data: {"result":{"response":{"finalMetadata":{"stop_reason":"end_turn"}}}}`,
	}, "\n")

	err := relayGrokSessionResponses(c, strings.NewReader(upstream), "grok-3", false, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	var response apicompat.ResponsesResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Len(t, response.Output, 1)
	require.Equal(
		t,
		"Answer\n\n## Sources\n[grok2api-sources]: #\n- [Example \\[Doc\\]](https://example.com/article)\n",
		response.Output[0].Content[0].Text,
	)
	require.Len(t, response.Output[0].SearchSources, 1)
	require.Equal(t, "https://example.com/article", response.Output[0].SearchSources[0].URL)
	require.Equal(t, "Example [Doc]", response.Output[0].SearchSources[0].Title)
}

func TestRelayGrokSessionResponses_StreamEmitsContentPartAndAnnotationEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	upstream := strings.Join([]string{
		`data: {"result":{"response":{"webSearchResults":{"results":[{"url":"https://example.com/article","title":"Example Article"}]}}}}`,
		`data: {"result":{"response":{"cardAttachment":{"jsonData":"{\"id\":\"card_1\",\"url\":\"https://example.com/article\"}"}}}}`,
		`data: {"result":{"response":{"token":"Answer<grok:render card_id=\"card_1\" card_type=\"citation\" type=\"render_inline_citation\"></grok:render>","messageTag":"final"}}}`,
		`data: {"result":{"response":{"finalMetadata":{"stop_reason":"end_turn"}}}}`,
	}, "\n")

	err := relayGrokSessionResponses(c, strings.NewReader(upstream), "grok-3", true, nil)
	require.NoError(t, err)

	events := decodeResponsesSSEEvents(t, rec.Body.String())
	require.Len(t, events, 10)
	require.Equal(t, "response.created", events[0].Type)
	require.Equal(t, "response.output_item.added", events[1].Type)
	require.Equal(t, "response.content_part.added", events[2].Type)
	require.Equal(t, "response.output_text.delta", events[3].Type)
	require.Equal(t, "Answer [[1]](https://example.com/article)", events[3].Delta)
	require.Equal(t, "response.output_text.annotation.added", events[4].Type)
	require.Equal(t, 0, events[4].AnnotationIndex)
	require.NotNil(t, events[4].Annotation)
	require.Equal(t, "url_citation", events[4].Annotation.Type)
	require.Equal(t, "https://example.com/article", events[4].Annotation.URL)
	require.Equal(t, "Example Article", events[4].Annotation.Title)
	require.Equal(t, 6, events[4].Annotation.StartIndex)
	require.Equal(t, 41, events[4].Annotation.EndIndex)
	require.Equal(t, "response.output_text.delta", events[5].Type)
	require.Equal(
		t,
		"\n\n## Sources\n[grok2api-sources]: #\n- [Example Article](https://example.com/article)\n",
		events[5].Delta,
	)
	require.Equal(t, "response.output_text.done", events[6].Type)
	require.Equal(
		t,
		"Answer [[1]](https://example.com/article)\n\n## Sources\n[grok2api-sources]: #\n- [Example Article](https://example.com/article)\n",
		events[6].Text,
	)
	require.Equal(t, "response.content_part.done", events[7].Type)
	require.NotNil(t, events[7].Part)
	require.Len(t, events[7].Part.Annotations, 1)
	require.Equal(t, "https://example.com/article", events[7].Part.Annotations[0].URL)
	require.Equal(t, "response.output_item.done", events[8].Type)
	require.Equal(t, "response.completed", events[9].Type)
	require.NotNil(t, events[9].Response)
	require.Len(t, events[9].Response.Output, 1)
	require.Len(t, events[9].Response.Output[0].Content, 1)
	require.Len(t, events[9].Response.Output[0].Content[0].Annotations, 1)
}

func TestRelayGrokSessionResponses_StreamEmitsSourcesSuffixBeforeDone(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	upstream := strings.Join([]string{
		`data: {"result":{"response":{"webSearchResults":{"results":[{"url":"https://example.com/article","title":"Example [Doc]"}]}}}}`,
		`data: {"result":{"response":{"token":"Answer","messageTag":"final"}}}`,
		`data: {"result":{"response":{"finalMetadata":{"stop_reason":"end_turn"}}}}`,
	}, "\n")

	err := relayGrokSessionResponses(c, strings.NewReader(upstream), "grok-3", true, nil)
	require.NoError(t, err)

	events := decodeResponsesSSEEvents(t, rec.Body.String())
	require.Len(t, events, 9)
	require.Equal(t, "response.created", events[0].Type)
	require.Equal(t, "response.output_item.added", events[1].Type)
	require.Equal(t, "response.content_part.added", events[2].Type)
	require.Equal(t, "Answer", events[3].Delta)
	require.Equal(
		t,
		"\n\n## Sources\n[grok2api-sources]: #\n- [Example \\[Doc\\]](https://example.com/article)\n",
		events[4].Delta,
	)
	require.Equal(t, "response.output_text.done", events[5].Type)
	require.Equal(
		t,
		"Answer\n\n## Sources\n[grok2api-sources]: #\n- [Example \\[Doc\\]](https://example.com/article)\n",
		events[5].Text,
	)
	require.Equal(t, "response.content_part.done", events[6].Type)
	require.NotNil(t, events[6].Part)
	require.Equal(
		t,
		"Answer\n\n## Sources\n[grok2api-sources]: #\n- [Example \\[Doc\\]](https://example.com/article)\n",
		events[6].Part.Text,
	)
	require.Equal(t, "response.output_item.done", events[7].Type)
	require.Equal(t, "response.completed", events[8].Type)
	require.NotNil(t, events[8].Response)
	require.Equal(
		t,
		"Answer\n\n## Sources\n[grok2api-sources]: #\n- [Example \\[Doc\\]](https://example.com/article)\n",
		events[8].Response.Output[0].Content[0].Text,
	)
}

func TestRelayGrokSessionResponses_BufferedParsesToolCalls(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	upstream := strings.Join([]string{
		`data: {"result":{"response":{"token":"<tool_calls><tool_call><tool_name>get_weather</tool_name><parameters>{\"city\":\"Shanghai\"}</parameters></tool_call></tool_calls>","messageTag":"final"}}}`,
		`data: {"result":{"response":{"finalMetadata":{"stop_reason":"tool_calls"}}}}`,
	}, "\n")

	err := relayGrokSessionResponses(c, strings.NewReader(upstream), "grok-3", false, []string{"get_weather"})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	var response apicompat.ResponsesResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Len(t, response.Output, 1)
	require.Equal(t, "function_call", response.Output[0].Type)
	require.Equal(t, "get_weather", response.Output[0].Name)
	require.Equal(t, `{"city":"Shanghai"}`, response.Output[0].Arguments)
	require.NotEmpty(t, response.Output[0].CallID)
}

func TestRelayGrokSessionResponses_StreamEmitsToolCallEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	upstream := strings.Join([]string{
		`data: {"result":{"response":{"token":"<tool_calls><tool_call><tool_name>get_weather</tool_name><parameters>{\"city\":\"Shanghai\"}</parameters></tool_call></tool_calls>","messageTag":"final"}}}`,
		`data: {"result":{"response":{"isSoftStop":true,"finalMetadata":{"stop_reason":"tool_calls"}}}}`,
	}, "\n")

	err := relayGrokSessionResponses(c, strings.NewReader(upstream), "grok-3", true, []string{"get_weather"})
	require.NoError(t, err)

	events := decodeResponsesSSEEvents(t, rec.Body.String())
	require.Len(t, events, 6)
	require.Equal(t, "response.created", events[0].Type)
	require.Equal(t, "response.output_item.added", events[1].Type)
	require.NotNil(t, events[1].Item)
	require.Equal(t, "function_call", events[1].Item.Type)
	require.Equal(t, "get_weather", events[1].Item.Name)
	require.Equal(t, "response.function_call_arguments.delta", events[2].Type)
	require.Equal(t, `{"city":"Shanghai"}`, events[2].Delta)
	require.Equal(t, "response.function_call_arguments.done", events[3].Type)
	require.Equal(t, "response.output_item.done", events[4].Type)
	require.Equal(t, "response.completed", events[5].Type)
	require.NotNil(t, events[5].Response)
	require.Len(t, events[5].Response.Output, 1)
	require.Equal(t, "function_call", events[5].Response.Output[0].Type)
}

func TestRelayGrokSessionResponses_StreamClosesReasoningBeforeToolCalls(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	upstream := strings.Join([]string{
		`data: {"result":{"response":{"token":"planning","isThinking":true}}}`,
		`data: {"result":{"response":{"token":"<tool_calls><tool_call><tool_name>get_weather</tool_name><parameters>{\"city\":\"Shanghai\"}</parameters></tool_call></tool_calls>","messageTag":"final"}}}`,
		`data: {"result":{"response":{"isSoftStop":true,"finalMetadata":{"stop_reason":"tool_calls"}}}}`,
	}, "\n")

	err := relayGrokSessionResponses(c, strings.NewReader(upstream), "grok-3", true, []string{"get_weather"})
	require.NoError(t, err)

	events := decodeResponsesSSEEvents(t, rec.Body.String())
	require.Len(t, events, 12)
	require.Equal(t, "response.created", events[0].Type)
	require.Equal(t, "response.output_item.added", events[1].Type)
	require.Equal(t, "response.reasoning_summary_part.added", events[2].Type)
	require.Equal(t, "response.reasoning_summary_text.delta", events[3].Type)
	require.Equal(t, "response.reasoning_summary_text.done", events[4].Type)
	require.Equal(t, "planning", events[4].Text)
	require.Equal(t, "response.reasoning_summary_part.done", events[5].Type)
	require.Equal(t, "response.output_item.done", events[6].Type)
	require.NotNil(t, events[6].Item)
	require.Equal(t, "reasoning", events[6].Item.Type)
	require.Equal(t, "response.output_item.added", events[7].Type)
	require.NotNil(t, events[7].Item)
	require.Equal(t, "function_call", events[7].Item.Type)
	require.Equal(t, "response.function_call_arguments.delta", events[8].Type)
	require.Equal(t, "response.function_call_arguments.done", events[9].Type)
	require.Equal(t, "response.output_item.done", events[10].Type)
	require.NotNil(t, events[10].Item)
	require.Equal(t, "function_call", events[10].Item.Type)
	require.Equal(t, "response.completed", events[11].Type)
	require.NotNil(t, events[11].Response)
	require.Len(t, events[11].Response.Output, 2)
	require.Equal(t, "reasoning", events[11].Response.Output[0].Type)
	require.Equal(t, "function_call", events[11].Response.Output[1].Type)
}

func TestRelayGrokSessionChatCompletions_StreamEmitsToolCallChunks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	upstream := strings.Join([]string{
		`data: {"result":{"response":{"token":"<tool_calls><tool_call><tool_name>get_weather</tool_name><parameters>{\"city\":\"Shanghai\"}</parameters></tool_call></tool_calls>","messageTag":"final"}}}`,
		`data: {"result":{"response":{"isSoftStop":true,"finalMetadata":{"stop_reason":"tool_calls"}}}}`,
	}, "\n")

	err := relayGrokSessionChatCompletions(c, strings.NewReader(upstream), "grok-3", true, false, []string{"get_weather"})
	require.NoError(t, err)

	chunks := decodeChatCompletionsSSEChunks(t, rec.Body.String())
	require.Len(t, chunks, 4)
	require.Equal(t, "assistant", chunks[0].Choices[0].Delta.Role)
	require.Len(t, chunks[1].Choices[0].Delta.ToolCalls, 1)
	require.Equal(t, "get_weather", chunks[1].Choices[0].Delta.ToolCalls[0].Function.Name)
	require.Len(t, chunks[2].Choices[0].Delta.ToolCalls, 1)
	require.Equal(t, `{"city":"Shanghai"}`, chunks[2].Choices[0].Delta.ToolCalls[0].Function.Arguments)
	require.NotNil(t, chunks[3].Choices[0].FinishReason)
	require.Equal(t, "tool_calls", *chunks[3].Choices[0].FinishReason)
}

func TestRelayGrokSessionChatCompletions_StreamPreservesSearchSourcesForToolCalls(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	upstream := strings.Join([]string{
		`data: {"result":{"response":{"webSearchResults":{"results":[{"url":"https://example.com/article","title":"Example Article"}]}}}}`,
		`data: {"result":{"response":{"token":"<tool_calls><tool_call><tool_name>get_weather</tool_name><parameters>{\"city\":\"Shanghai\"}</parameters></tool_call></tool_calls>","messageTag":"final"}}}`,
		`data: {"result":{"response":{"isSoftStop":true,"finalMetadata":{"stop_reason":"tool_calls"}}}}`,
	}, "\n")

	err := relayGrokSessionChatCompletions(c, strings.NewReader(upstream), "grok-3", true, false, []string{"get_weather"})
	require.NoError(t, err)

	chunks := decodeChatCompletionsSSEChunks(t, rec.Body.String())
	require.Len(t, chunks, 4)
	require.Len(t, chunks[3].SearchSources, 1)
	require.Equal(t, "https://example.com/article", chunks[3].SearchSources[0].URL)
	require.Equal(t, "Example Article", chunks[3].SearchSources[0].Title)
	require.Equal(t, "web", chunks[3].SearchSources[0].Type)
}

func TestRelayGrokSessionChatCompletions_BufferedPreservesSearchSourcesForToolCalls(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	upstream := strings.Join([]string{
		`data: {"result":{"response":{"webSearchResults":{"results":[{"url":"https://example.com/article","title":"Example Article"}]}}}}`,
		`data: {"result":{"response":{"token":"<tool_calls><tool_call><tool_name>get_weather</tool_name><parameters>{\"city\":\"Shanghai\"}</parameters></tool_call></tool_calls>","messageTag":"final"}}}`,
		`data: {"result":{"response":{"isSoftStop":true,"finalMetadata":{"stop_reason":"tool_calls"}}}}`,
	}, "\n")

	err := relayGrokSessionChatCompletions(c, strings.NewReader(upstream), "grok-3", false, false, []string{"get_weather"})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	var response apicompat.ChatCompletionsResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Len(t, response.Choices, 1)
	require.Len(t, response.Choices[0].Message.ToolCalls, 1)
	require.Len(t, response.SearchSources, 1)
	require.Equal(t, "https://example.com/article", response.SearchSources[0].URL)
	require.Equal(t, "Example Article", response.SearchSources[0].Title)
	require.Equal(t, "web", response.SearchSources[0].Type)
}

func TestRelayGrokSessionResponses_StreamEmitsAltFunctionCallEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	upstream := strings.Join([]string{
		`data: {"result":{"response":{"token":"<function_call><name>get_weather</name><arguments>{\"city\":\"Shanghai\"}</arguments></function_call>","messageTag":"final"}}}`,
		`data: {"result":{"response":{"isSoftStop":true,"finalMetadata":{"stop_reason":"tool_calls"}}}}`,
	}, "\n")

	err := relayGrokSessionResponses(c, strings.NewReader(upstream), "grok-3", true, []string{"get_weather"})
	require.NoError(t, err)

	events := decodeResponsesSSEEvents(t, rec.Body.String())
	require.Len(t, events, 6)
	require.Equal(t, "response.created", events[0].Type)
	require.Equal(t, "response.output_item.added", events[1].Type)
	require.NotNil(t, events[1].Item)
	require.Equal(t, "function_call", events[1].Item.Type)
	require.Equal(t, "get_weather", events[1].Item.Name)
	require.Equal(t, "response.function_call_arguments.delta", events[2].Type)
	require.Equal(t, `{"city":"Shanghai"}`, events[2].Delta)
	require.Equal(t, "response.function_call_arguments.done", events[3].Type)
	require.Equal(t, "response.output_item.done", events[4].Type)
	require.Equal(t, "response.completed", events[5].Type)
}

func TestRelayGrokSessionChatCompletions_StreamEmitsInvokeToolCallChunks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	upstream := strings.Join([]string{
		`data: {"result":{"response":{"token":"<invoke name=\"get_weather\">{\"city\":\"Shanghai\"}</invoke>","messageTag":"final"}}}`,
		`data: {"result":{"response":{"isSoftStop":true,"finalMetadata":{"stop_reason":"tool_calls"}}}}`,
	}, "\n")

	err := relayGrokSessionChatCompletions(c, strings.NewReader(upstream), "grok-3", true, false, []string{"get_weather"})
	require.NoError(t, err)

	chunks := decodeChatCompletionsSSEChunks(t, rec.Body.String())
	require.Len(t, chunks, 4)
	require.Equal(t, "assistant", chunks[0].Choices[0].Delta.Role)
	require.Len(t, chunks[1].Choices[0].Delta.ToolCalls, 1)
	require.Equal(t, "get_weather", chunks[1].Choices[0].Delta.ToolCalls[0].Function.Name)
	require.Len(t, chunks[2].Choices[0].Delta.ToolCalls, 1)
	require.Equal(t, `{"city":"Shanghai"}`, chunks[2].Choices[0].Delta.ToolCalls[0].Function.Arguments)
	require.NotNil(t, chunks[3].Choices[0].FinishReason)
	require.Equal(t, "tool_calls", *chunks[3].Choices[0].FinishReason)
}

func TestRelayGrokSessionAnthropic_StreamEmitsToolUseBlocks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	upstream := strings.Join([]string{
		`data: {"result":{"response":{"token":"<tool_calls><tool_call><tool_name>get_weather</tool_name><parameters>{\"city\":\"Shanghai\"}</parameters></tool_call></tool_calls>","messageTag":"final"}}}`,
		`data: {"result":{"response":{"isSoftStop":true,"finalMetadata":{"stop_reason":"tool_calls"}}}}`,
	}, "\n")

	err := relayGrokSessionAnthropic(c, strings.NewReader(upstream), "grok-3", true, []string{"get_weather"})
	require.NoError(t, err)

	events := decodeAnthropicSSEEvents(t, rec.Body.String())
	require.GreaterOrEqual(t, len(events), 5)
	require.Equal(t, "message_start", events[0].Type)
	require.Equal(t, "content_block_start", events[1].Type)
	require.NotNil(t, events[1].ContentBlock)
	require.Equal(t, "tool_use", events[1].ContentBlock.Type)
	require.Equal(t, "get_weather", events[1].ContentBlock.Name)
	require.Equal(t, "message_delta", events[len(events)-2].Type)
	require.NotNil(t, events[len(events)-2].Delta)
	require.Equal(t, "tool_use", events[len(events)-2].Delta.StopReason)
	require.Equal(t, "message_stop", events[len(events)-1].Type)
}

func TestRelayGrokSessionAnthropic_StreamPreservesSearchSourcesForToolUse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	upstream := strings.Join([]string{
		`data: {"result":{"response":{"webSearchResults":{"results":[{"url":"https://example.com/article","title":"Example Article"}]}}}}`,
		`data: {"result":{"response":{"token":"<tool_calls><tool_call><tool_name>get_weather</tool_name><parameters>{\"city\":\"Shanghai\"}</parameters></tool_call></tool_calls>","messageTag":"final"}}}`,
		`data: {"result":{"response":{"isSoftStop":true,"finalMetadata":{"stop_reason":"tool_calls"}}}}`,
	}, "\n")

	err := relayGrokSessionAnthropic(c, strings.NewReader(upstream), "grok-3", true, []string{"get_weather"})
	require.NoError(t, err)

	events := decodeAnthropicSSEEvents(t, rec.Body.String())
	require.GreaterOrEqual(t, len(events), 2)
	require.Equal(t, "message_delta", events[len(events)-2].Type)
	require.NotNil(t, events[len(events)-2].Delta)
	require.Len(t, events[len(events)-2].Delta.SearchSources, 1)
	require.Equal(t, "https://example.com/article", events[len(events)-2].Delta.SearchSources[0].URL)
	require.Equal(t, "Example Article", events[len(events)-2].Delta.SearchSources[0].Title)
	require.Equal(t, "web", events[len(events)-2].Delta.SearchSources[0].Type)
}

func TestRelayGrokSessionAnthropic_BufferedPreservesSearchSourcesForToolUse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	upstream := strings.Join([]string{
		`data: {"result":{"response":{"webSearchResults":{"results":[{"url":"https://example.com/article","title":"Example Article"}]}}}}`,
		`data: {"result":{"response":{"token":"<tool_calls><tool_call><tool_name>get_weather</tool_name><parameters>{\"city\":\"Shanghai\"}</parameters></tool_call></tool_calls>","messageTag":"final"}}}`,
		`data: {"result":{"response":{"isSoftStop":true,"finalMetadata":{"stop_reason":"tool_calls"}}}}`,
	}, "\n")

	err := relayGrokSessionAnthropic(c, strings.NewReader(upstream), "grok-3", false, []string{"get_weather"})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	var response apicompat.AnthropicResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Len(t, response.Content, 1)
	require.Equal(t, "tool_use", response.Content[0].Type)
	require.Len(t, response.SearchSources, 1)
	require.Equal(t, "https://example.com/article", response.SearchSources[0].URL)
	require.Equal(t, "Example Article", response.SearchSources[0].Title)
	require.Equal(t, "web", response.SearchSources[0].Type)
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

func decodeAnthropicSSEEvents(t *testing.T, raw string) []apicompat.AnthropicStreamEvent {
	t.Helper()

	frames := strings.Split(strings.TrimSpace(raw), "\n\n")
	events := make([]apicompat.AnthropicStreamEvent, 0, len(frames))
	for _, frame := range frames {
		lines := strings.Split(strings.TrimSpace(frame), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			payload := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
			if payload == "" {
				continue
			}
			var event apicompat.AnthropicStreamEvent
			require.NoError(t, json.Unmarshal([]byte(payload), &event))
			events = append(events, event)
		}
	}
	return events
}
