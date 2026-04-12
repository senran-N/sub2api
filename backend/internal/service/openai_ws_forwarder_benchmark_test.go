package service

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/config"
)

var (
	benchmarkOpenAIWSPayloadJSONSink string
	benchmarkOpenAIWSStringSink      string
	benchmarkOpenAIWSBoolSink        bool
	benchmarkOpenAIWSBytesSink       []byte
)

func BenchmarkOpenAIWSForwarderHotPath(b *testing.B) {
	cfg := &config.Config{}
	svc := &OpenAIGatewayService{cfg: cfg}
	account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	reqBody := benchmarkOpenAIWSHotPathRequest()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		payload := svc.buildOpenAIWSCreatePayload(reqBody, account)
		_, _ = applyOpenAIWSRetryPayloadStrategy(payload, 2)
		setOpenAIWSTurnMetadata(payload, `{"trace":"bench","turn":"1"}`)

		benchmarkOpenAIWSStringSink = openAIWSPayloadString(payload, "previous_response_id")
		benchmarkOpenAIWSBoolSink = payload["tools"] != nil
		benchmarkOpenAIWSStringSink = summarizeOpenAIWSPayloadKeySizes(payload, openAIWSPayloadKeySizeTopN)
		benchmarkOpenAIWSStringSink = summarizeOpenAIWSInput(payload["input"])
		benchmarkOpenAIWSPayloadJSONSink = payloadAsJSON(payload)
	}
}

func BenchmarkBuildOpenAIWSIngressPayloadMeta(b *testing.B) {
	svc := &OpenAIGatewayService{}
	account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	payload := []byte(`{"type":"response.create","model":"gpt-5.3-codex","stream":false,"previous_response_id":"resp_benchmark_prev","prompt_cache_key":"bench-cache-key","reasoning":{"effort":"medium"},"input":[{"type":"function_call_output","call_id":"call_1"}],"store":false}`)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		meta := svc.buildOpenAIWSIngressPayloadMeta(payload, account, true)
		benchmarkOpenAIWSStringSink = meta.previousResponseID
		benchmarkOpenAIWSStringSink = meta.promptCacheKey
		benchmarkOpenAIWSBoolSink = meta.hasFunctionCallOutput
		benchmarkOpenAIWSBoolSink = meta.strictAffinityTurn
	}
}

func BenchmarkNormalizeOpenAIWSPayloadWithoutInputAndPreviousResponseID(b *testing.B) {
	payload := []byte(`{"type":"response.create","model":"gpt-5.3-codex","stream":false,"previous_response_id":"resp_benchmark_prev","prompt_cache_key":"bench-cache-key","reasoning":{"effort":"medium"},"input":[{"type":"function_call_output","call_id":"call_1","output":{"status":"ok"}}],"metadata":{"a":1,"b":2},"tools":[{"type":"function","name":"tool_1"}],"store":false}`)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		normalized, err := normalizeOpenAIWSPayloadWithoutInputAndPreviousResponseID(payload)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkOpenAIWSBytesSink = normalized
	}
}

func BenchmarkParseAndPrepareOpenAIWSIngressClientPayload(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	req := httptest.NewRequest("GET", "/v1/responses/ws", nil)
	req.Header.Set(openAIWSTurnMetadataHeader, `{"trace":"bench"}`)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = req

	account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	svc := &OpenAIGatewayService{}
	payload := []byte(`{"type":"response.create","model":"gpt-5.3-codex","stream":false,"previous_response_id":"resp_benchmark_prev","prompt_cache_key":"bench-cache-key","reasoning":{"effort":"medium"},"input":[{"type":"function_call_output","call_id":"call_1","output":{"status":"ok"}}],"metadata":{"a":1,"b":2},"tools":[{"type":"function","name":"tool_1"}],"store":false}`)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		parsed, err := svc.parseOpenAIWSIngressClientPayload(ctx, account, payload)
		if err != nil {
			b.Fatal(err)
		}
		prepared := svc.prepareOpenAIWSClientPayload(account, parsed)
		benchmarkOpenAIWSBytesSink = prepared.payloadRaw
		benchmarkOpenAIWSBoolSink = prepared.storeDisabled
		benchmarkOpenAIWSStringSink = prepared.payloadMeta.previousResponseID
	}
}

func BenchmarkBuildOpenAIWSReplayInputSequence_PrefixReuse(b *testing.B) {
	previous := []json.RawMessage{
		json.RawMessage(`{"type":"message","role":"user","content":"hello"}`),
		json.RawMessage(`{"type":"message","role":"assistant","content":"world"}`),
	}
	currentPayload := []byte(`{"previous_response_id":"resp_1","input":[{"type":"message","role":"user","content":"hello"},{"type":"message","role":"assistant","content":"world"},{"type":"message","role":"user","content":"next"}]}`)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		items, exists, err := buildOpenAIWSReplayInputSequence(previous, true, currentPayload, true)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkOpenAIWSBoolSink = exists
		benchmarkOpenAIWSBytesSink = items[len(items)-1]
	}
}

func BenchmarkBuildOpenAIWSReplayInputSequence_Merge(b *testing.B) {
	previous := []json.RawMessage{
		json.RawMessage(`{"type":"message","role":"user","content":"hello"}`),
		json.RawMessage(`{"type":"message","role":"assistant","content":"world"}`),
	}
	currentPayload := []byte(`{"previous_response_id":"resp_1","input":[{"type":"function_call_output","call_id":"call_1","output":"ok"}]}`)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		items, exists, err := buildOpenAIWSReplayInputSequence(previous, true, currentPayload, true)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkOpenAIWSBoolSink = exists
		benchmarkOpenAIWSBytesSink = items[len(items)-1]
	}
}

func BenchmarkRewriteOpenAIWSPayload_DropPreviousAndSetInput(b *testing.B) {
	payload := []byte(`{"type":"response.create","model":"gpt-5.3-codex","previous_response_id":"resp_old","input":[{"type":"input_text","text":"stale"}],"store":false}`)
	items := []json.RawMessage{
		json.RawMessage(`{"type":"input_text","text":"hello"}`),
		json.RawMessage(`{"type":"function_call_output","call_id":"call_1","output":"ok"}`),
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		updated, err := rewriteOpenAIWSPayload(payload, openAIWSPayloadRewriteOptions{
			dropPreviousResponseID: true,
			setInput:               true,
			input:                  items,
		})
		if err != nil {
			b.Fatal(err)
		}
		benchmarkOpenAIWSBytesSink = updated
	}
}

func BenchmarkSetPreviousResponseIDToRawPayload_SameValue(b *testing.B) {
	payload := []byte(`{"type":"response.create","model":"gpt-5.3-codex","previous_response_id":"resp_same","input":[{"type":"input_text","text":"hello"}]}`)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		updated, err := setPreviousResponseIDToRawPayload(payload, "resp_same")
		if err != nil {
			b.Fatal(err)
		}
		benchmarkOpenAIWSBytesSink = updated
	}
}

func benchmarkOpenAIWSHotPathRequest() map[string]any {
	tools := make([]map[string]any, 0, 24)
	for i := 0; i < 24; i++ {
		tools = append(tools, map[string]any{
			"type":        "function",
			"name":        fmt.Sprintf("tool_%02d", i),
			"description": "benchmark tool schema",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{"type": "string"},
					"limit": map[string]any{"type": "number"},
				},
				"required": []string{"query"},
			},
		})
	}

	input := make([]map[string]any, 0, 16)
	for i := 0; i < 16; i++ {
		input = append(input, map[string]any{
			"role":    "user",
			"type":    "message",
			"content": fmt.Sprintf("benchmark message %d", i),
		})
	}

	return map[string]any{
		"type":                 "response.create",
		"model":                "gpt-5.3-codex",
		"input":                input,
		"tools":                tools,
		"parallel_tool_calls":  true,
		"previous_response_id": "resp_benchmark_prev",
		"prompt_cache_key":     "bench-cache-key",
		"reasoning":            map[string]any{"effort": "medium"},
		"instructions":         "benchmark instructions",
		"store":                false,
	}
}

func BenchmarkOpenAIWSEventEnvelopeParse(b *testing.B) {
	event := []byte(`{"type":"response.completed","response":{"id":"resp_bench_1","model":"gpt-5.1","usage":{"input_tokens":12,"output_tokens":8}}}`)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		eventType, responseID, response := parseOpenAIWSEventEnvelope(event)
		benchmarkOpenAIWSStringSink = eventType
		benchmarkOpenAIWSStringSink = responseID
		benchmarkOpenAIWSBoolSink = response.Exists()
	}
}

func BenchmarkOpenAIWSErrorEventFieldReuse(b *testing.B) {
	event := []byte(`{"type":"error","error":{"type":"invalid_request_error","code":"invalid_request","message":"invalid input"}}`)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		codeRaw, errTypeRaw, errMsgRaw := parseOpenAIWSErrorEventFields(event)
		benchmarkOpenAIWSStringSink, benchmarkOpenAIWSBoolSink = classifyOpenAIWSErrorEventFromRaw(codeRaw, errTypeRaw, errMsgRaw)
		code, errType, errMsg := summarizeOpenAIWSErrorEventFieldsFromRaw(codeRaw, errTypeRaw, errMsgRaw)
		benchmarkOpenAIWSStringSink = code
		benchmarkOpenAIWSStringSink = errType
		benchmarkOpenAIWSStringSink = errMsg
		benchmarkOpenAIWSBoolSink = openAIWSErrorHTTPStatusFromRaw(codeRaw, errTypeRaw) > 0
	}
}

func BenchmarkReplaceOpenAIWSMessageModel_NoMatchFastPath(b *testing.B) {
	event := []byte(`{"type":"response.output_text.delta","delta":"hello world"}`)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkOpenAIWSBytesSink = replaceOpenAIWSMessageModel(event, "gpt-5.1", "custom-model")
	}
}

func BenchmarkReplaceOpenAIWSMessageModel_DualReplace(b *testing.B) {
	event := []byte(`{"type":"response.completed","model":"gpt-5.1","response":{"id":"resp_1","model":"gpt-5.1"}}`)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkOpenAIWSBytesSink = replaceOpenAIWSMessageModel(event, "gpt-5.1", "custom-model")
	}
}
