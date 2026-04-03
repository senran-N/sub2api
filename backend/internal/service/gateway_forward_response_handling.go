package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

func (s *GatewayService) handleForwardUpstreamResponse(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	resp *http.Response,
) (*ForwardResult, error, bool) {
	if resp.StatusCode >= 400 && s.shouldRetryUpstreamError(account, resp.StatusCode) {
		if s.shouldFailoverUpstreamError(resp.StatusCode) {
			respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
			_ = resp.Body.Close()
			resp.Body = io.NopCloser(bytes.NewReader(respBody))

			logger.LegacyPrintf("service.gateway", "[Forward] Upstream error (retry exhausted, failover): Account=%d(%s) Status=%d RequestID=%s Body=%s",
				account.ID, account.Name, resp.StatusCode, resp.Header.Get("x-request-id"), truncateString(string(respBody), 1000))

			s.handleRetryExhaustedSideEffects(ctx, resp, account)
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform:           account.Platform,
				AccountID:          account.ID,
				AccountName:        account.Name,
				UpstreamStatusCode: resp.StatusCode,
				UpstreamRequestID:  resp.Header.Get("x-request-id"),
				Kind:               "retry_exhausted_failover",
				Message:            extractUpstreamErrorMessage(respBody),
				Detail: func() string {
					if s.cfg != nil && s.cfg.Gateway.LogUpstreamErrorBody {
						return truncateString(string(respBody), s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes)
					}
					return ""
				}(),
			})
			return nil, &UpstreamFailoverError{
				StatusCode:             resp.StatusCode,
				ResponseBody:           respBody,
				RetryableOnSameAccount: account.IsPoolMode() && isPoolModeRetryableStatus(resp.StatusCode),
			}, true
		}
		result, err := s.handleRetryExhaustedError(ctx, resp, c, account)
		return result, err, true
	}

	if resp.StatusCode >= 400 && s.shouldFailoverUpstreamError(resp.StatusCode) {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(respBody))

		logger.LegacyPrintf("service.gateway", "[Forward] Upstream error (failover): Account=%d(%s) Status=%d RequestID=%s Body=%s",
			account.ID, account.Name, resp.StatusCode, resp.Header.Get("x-request-id"), truncateString(string(respBody), 1000))

		s.handleFailoverSideEffects(ctx, resp, account)
		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			Platform:           account.Platform,
			AccountID:          account.ID,
			UpstreamStatusCode: resp.StatusCode,
			UpstreamRequestID:  resp.Header.Get("x-request-id"),
			Kind:               "failover",
			Message:            extractUpstreamErrorMessage(respBody),
			Detail: func() string {
				if s.cfg != nil && s.cfg.Gateway.LogUpstreamErrorBody {
					return truncateString(string(respBody), s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes)
				}
				return ""
			}(),
		})
		return nil, &UpstreamFailoverError{
			StatusCode:             resp.StatusCode,
			ResponseBody:           respBody,
			RetryableOnSameAccount: account.IsPoolMode() && isPoolModeRetryableStatus(resp.StatusCode),
		}, true
	}

	if resp.StatusCode < 400 {
		return nil, nil, false
	}

	if resp.StatusCode == http.StatusBadRequest && s.cfg != nil && s.cfg.Gateway.FailoverOn400 {
		respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		if readErr != nil {
			result, err := s.handleErrorResponse(ctx, resp, c, account)
			return result, err, true
		}
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(respBody))

		if s.shouldFailoverOn400(respBody) {
			upstreamMsg := strings.TrimSpace(extractUpstreamErrorMessage(respBody))
			upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
			upstreamDetail := ""
			if s.cfg != nil && s.cfg.Gateway.LogUpstreamErrorBody {
				maxBytes := s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes
				if maxBytes <= 0 {
					maxBytes = 2048
				}
				upstreamDetail = truncateString(string(respBody), maxBytes)
			}
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform:           account.Platform,
				AccountID:          account.ID,
				AccountName:        account.Name,
				UpstreamStatusCode: resp.StatusCode,
				UpstreamRequestID:  resp.Header.Get("x-request-id"),
				Kind:               "failover_on_400",
				Message:            upstreamMsg,
				Detail:             upstreamDetail,
			})

			if s.cfg.Gateway.LogUpstreamErrorBody {
				logger.LegacyPrintf("service.gateway",
					"Account %d: 400 error, attempting failover: %s",
					account.ID,
					truncateForLog(respBody, s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes),
				)
			} else {
				logger.LegacyPrintf("service.gateway", "Account %d: 400 error, attempting failover", account.ID)
			}
			s.handleFailoverSideEffects(ctx, resp, account)
			return nil, &UpstreamFailoverError{StatusCode: resp.StatusCode, ResponseBody: respBody}, true
		}
	}

	result, err := s.handleErrorResponse(ctx, resp, c, account)
	return result, err, true
}

func (s *GatewayService) finalizeForwardSuccess(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	resp *http.Response,
	parsed *ParsedRequest,
	preparation *forwardRequestPreparation,
	startTime time.Time,
) (*ForwardResult, error) {
	if parsed.OnUpstreamAccepted != nil {
		parsed.OnUpstreamAccepted()
	}

	var usage *ClaudeUsage
	var firstTokenMs *int
	var clientDisconnect bool
	if preparation.reqStream {
		streamResult, err := s.handleStreamingResponse(
			ctx,
			resp,
			c,
			account,
			startTime,
			preparation.originalModel,
			preparation.reqModel,
			preparation.shouldMimicClaudeCode,
		)
		if err != nil {
			var streamErr *upstreamStreamEventError
			if errors.As(err, &streamErr) {
				return nil, &UpstreamFailoverError{
					StatusCode:   streamErr.StatusCode(),
					ResponseBody: streamErr.ResponseBody(),
				}
			}
			return nil, err
		}
		usage = streamResult.usage
		firstTokenMs = streamResult.firstTokenMs
		clientDisconnect = streamResult.clientDisconnect
	} else {
		var err error
		usage, err = s.handleNonStreamingResponse(ctx, resp, c, account, preparation.originalModel, preparation.reqModel)
		if err != nil {
			return nil, err
		}
	}

	return &ForwardResult{
		RequestID:        resp.Header.Get("x-request-id"),
		Usage:            *usage,
		Model:            preparation.originalModel,
		UpstreamModel:    preparation.reqModel,
		Stream:           preparation.reqStream,
		Duration:         time.Since(startTime),
		FirstTokenMs:     firstTokenMs,
		ClientDisconnect: clientDisconnect,
	}, nil
}
