package service

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

func setAntigravityRequestIDHeader(c *gin.Context, resp *http.Response) string {
	requestID := resp.Header.Get("x-request-id")
	if requestID != "" {
		c.Header("x-request-id", requestID)
	}
	return requestID
}

func (s *AntigravityGatewayService) mapClaudeRetryLoopError(c *gin.Context, err error) error {
	if switchErr, ok := IsAntigravityAccountSwitchError(err); ok {
		return &UpstreamFailoverError{
			StatusCode:        http.StatusServiceUnavailable,
			ForceCacheBilling: switchErr.IsStickySession,
		}
	}
	if c.Request.Context().Err() != nil {
		return s.writeClaudeError(c, http.StatusBadGateway, "client_disconnected", "Client disconnected before upstream response")
	}
	return s.writeClaudeError(c, http.StatusBadGateway, "upstream_error", "Upstream request failed after retries")
}

func (s *AntigravityGatewayService) mapGeminiRetryLoopError(c *gin.Context, err error) error {
	if switchErr, ok := IsAntigravityAccountSwitchError(err); ok {
		return &UpstreamFailoverError{
			StatusCode:        http.StatusServiceUnavailable,
			ForceCacheBilling: switchErr.IsStickySession,
		}
	}
	if c.Request.Context().Err() != nil {
		return s.writeGoogleError(c, http.StatusBadGateway, "Client disconnected before upstream response")
	}
	return s.writeGoogleError(c, http.StatusBadGateway, "Upstream request failed after retries")
}

func (s *AntigravityGatewayService) completeClaudeForwardSuccess(
	c *gin.Context,
	resp *http.Response,
	startTime time.Time,
	prefix string,
	originalModel string,
	upstreamModel string,
	stream bool,
) (*ForwardResult, error) {
	requestID := setAntigravityRequestIDHeader(c, resp)

	var usage *ClaudeUsage
	var firstTokenMs *int
	var clientDisconnect bool

	if stream {
		streamRes, err := s.handleClaudeStreamingResponse(c, resp, startTime, originalModel)
		if err != nil {
			logger.LegacyPrintf("service.antigravity_gateway", "%s status=stream_error error=%v", prefix, err)
			return nil, err
		}
		usage = streamRes.usage
		firstTokenMs = streamRes.firstTokenMs
		clientDisconnect = streamRes.clientDisconnect
	} else {
		streamRes, err := s.handleClaudeStreamToNonStreaming(c, resp, startTime, originalModel)
		if err != nil {
			logger.LegacyPrintf("service.antigravity_gateway", "%s status=stream_collect_error error=%v", prefix, err)
			return nil, err
		}
		usage = streamRes.usage
		firstTokenMs = streamRes.firstTokenMs
	}

	return &ForwardResult{
		RequestID:        requestID,
		Usage:            *usage,
		Model:            originalModel,
		UpstreamModel:    upstreamModel,
		Stream:           stream,
		Duration:         time.Since(startTime),
		FirstTokenMs:     firstTokenMs,
		ClientDisconnect: clientDisconnect,
	}, nil
}

func (s *AntigravityGatewayService) completeGeminiForwardSuccess(
	c *gin.Context,
	resp *http.Response,
	startTime time.Time,
	prefix string,
	originalModel string,
	upstreamModel string,
	stream bool,
	imageSize string,
) (*ForwardResult, error) {
	requestID := setAntigravityRequestIDHeader(c, resp)

	var usage *ClaudeUsage
	var firstTokenMs *int
	var clientDisconnect bool

	if stream {
		streamRes, err := s.handleGeminiStreamingResponse(c, resp, startTime)
		if err != nil {
			logger.LegacyPrintf("service.antigravity_gateway", "%s status=stream_error error=%v", prefix, err)
			return nil, err
		}
		usage = streamRes.usage
		firstTokenMs = streamRes.firstTokenMs
		clientDisconnect = streamRes.clientDisconnect
	} else {
		streamRes, err := s.handleGeminiStreamToNonStreaming(c, resp, startTime)
		if err != nil {
			logger.LegacyPrintf("service.antigravity_gateway", "%s status=stream_collect_error error=%v", prefix, err)
			return nil, err
		}
		usage = streamRes.usage
		firstTokenMs = streamRes.firstTokenMs
	}

	if usage == nil {
		usage = &ClaudeUsage{}
	}

	imageCount := 0
	if isImageGenerationModel(upstreamModel) {
		imageCount = 1
	}

	return &ForwardResult{
		RequestID:        requestID,
		Usage:            *usage,
		Model:            originalModel,
		UpstreamModel:    upstreamModel,
		Stream:           stream,
		Duration:         time.Since(startTime),
		FirstTokenMs:     firstTokenMs,
		ClientDisconnect: clientDisconnect,
		ImageCount:       imageCount,
		ImageSize:        imageSize,
	}, nil
}
