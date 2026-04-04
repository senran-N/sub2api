package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/senran-N/sub2api/internal/pkg/antigravity"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

// TestConnectionResult 测试连接结果
type TestConnectionResult struct {
	Text        string // 响应文本
	MappedModel string // 实际使用的模型
}

// TestConnection 测试 Antigravity 账号连接。
// 复用 antigravityRetryLoop 的完整重试 / credits overages / 智能重试逻辑，
// 与真实调度行为一致。差异：不做账号切换（测试指定账号）、不记录 ops 错误。
func (s *AntigravityGatewayService) TestConnection(ctx context.Context, account *Account, modelID string) (*TestConnectionResult, error) {
	transport, err := s.resolveForwardTransportContext(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("获取 access_token 失败: %w", err)
	}

	mappedModel := s.getMappedModel(account, modelID)
	if mappedModel == "" {
		return nil, fmt.Errorf("model %s not in whitelist", modelID)
	}

	var requestBody []byte
	if strings.HasPrefix(modelID, "gemini-") {
		requestBody, err = s.buildGeminiTestRequest(transport.projectID, mappedModel)
	} else {
		requestBody, err = s.buildClaudeTestRequest(transport.projectID, mappedModel)
	}
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %w", err)
	}

	prefix := fmt.Sprintf("[antigravity-Test] account=%d(%s)", account.ID, account.Name)
	p := s.newRetryLoopParams(ctx, prefix, account, transport, antigravityUpstreamStreamAction, requestBody, nil, modelID, false, testConnectionHandleError)

	result, err := s.antigravityRetryLoop(p)
	if err != nil {
		var switchErr *AntigravityAccountSwitchError
		if errors.As(err, &switchErr) {
			return nil, fmt.Errorf("该账号模型 %s 当前限流中，请稍后重试", switchErr.RateLimitedModel)
		}
		return nil, err
	}

	if result == nil || result.resp == nil {
		return nil, errors.New("upstream returned empty response")
	}
	defer func() { _ = result.resp.Body.Close() }()

	respBody, err := io.ReadAll(io.LimitReader(result.resp.Body, 2<<20))
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	if result.resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API 返回 %d: %s", result.resp.StatusCode, string(respBody))
	}

	return &TestConnectionResult{
		Text:        extractTextFromSSEResponse(respBody),
		MappedModel: mappedModel,
	}, nil
}

// testConnectionHandleError 是 TestConnection 使用的轻量 handleError 回调。
// 仅记录日志，不做 ops 错误追踪或粘性会话清除。
func testConnectionHandleError(
	_ context.Context, prefix string, account *Account,
	statusCode int, _ http.Header, body []byte,
	requestedModel string, _ int64, _ string, _ bool,
) *handleModelRateLimitResult {
	logger.LegacyPrintf("service.antigravity_gateway",
		"%s test_handle_error status=%d model=%s account=%d body=%s",
		prefix, statusCode, requestedModel, account.ID, truncateForLog(body, 200))
	return nil
}

// buildGeminiTestRequest 构建 Gemini 格式测试请求
// 使用最小 token 消耗：输入 "." + maxOutputTokens: 1
func (s *AntigravityGatewayService) buildGeminiTestRequest(projectID, model string) ([]byte, error) {
	payload := map[string]any{
		"contents": []map[string]any{
			{
				"role": "user",
				"parts": []map[string]any{
					{"text": "."},
				},
			},
		},
		"systemInstruction": map[string]any{
			"parts": []map[string]any{
				{"text": antigravity.GetDefaultIdentityPatch()},
			},
		},
		"generationConfig": map[string]any{
			"maxOutputTokens": 1,
		},
	}
	payloadBytes, _ := json.Marshal(payload)
	return s.wrapV1InternalRequest(projectID, model, payloadBytes)
}

// buildClaudeTestRequest 构建 Claude 格式测试请求并转换为 Gemini 格式
// 使用最小 token 消耗：输入 "." + MaxTokens: 1
func (s *AntigravityGatewayService) buildClaudeTestRequest(projectID, mappedModel string) ([]byte, error) {
	claudeReq := &antigravity.ClaudeRequest{
		Model: mappedModel,
		Messages: []antigravity.ClaudeMessage{
			{
				Role:    "user",
				Content: json.RawMessage(`"."`),
			},
		},
		MaxTokens: 1,
		Stream:    false,
	}
	return antigravity.TransformClaudeToGemini(claudeReq, projectID, mappedModel)
}
