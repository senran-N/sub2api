package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/apicompat"
	"github.com/tidwall/gjson"
)

type GrokTextRuntime struct {
	gatewayService    *GatewayService
	compatibleRuntime grokTextCompatibleRuntime
	sessionRuntime    grokTextSessionRuntime
}

type grokTextCompatibleRuntime interface {
	Execute(*gin.Context, *grokTextPreparation) error
}

type grokTextSessionRuntime interface {
	Execute(*gin.Context, *grokTextPreparation) error
}

type grokTextPreparation struct {
	protocolFamily CompatibleGatewayProtocolFamily
	requestedModel string
	stream         bool
	includeUsage   bool
	toolNames      []string
	account        *Account
	compatibleBody []byte
	usesCompatible bool
	target         grokTransportTarget
	payload        []byte
}

type grokSessionTextPreparedPayload struct {
	requestedModel   string
	stream           bool
	includeUsage     bool
	toolNames        []string
	payloadMap       map[string]any
	payload          []byte
	attachmentInputs []grokSessionUploadInput
}

type grokResponsesHTTPError struct {
	statusCode int
	code       string
	message    string
}

func (e *grokResponsesHTTPError) Error() string {
	if e == nil {
		return ""
	}
	return e.message
}

func newGrokResponsesHTTPError(statusCode int, code, message string) error {
	return &grokResponsesHTTPError{
		statusCode: statusCode,
		code:       strings.TrimSpace(code),
		message:    strings.TrimSpace(message),
	}
}

func NewGrokTextRuntime(
	gatewayService *GatewayService,
	compatibleRuntime *GrokCompatibleRuntime,
	sessionRuntime *GrokSessionRuntime,
) *GrokTextRuntime {
	return &GrokTextRuntime{
		gatewayService:    gatewayService,
		compatibleRuntime: compatibleRuntime,
		sessionRuntime:    sessionRuntime,
	}
}

func (r *GrokTextRuntime) HandleResponses(c *gin.Context, groupID *int64, body []byte) bool {
	return r.handleTextRequest(c, groupID, body, CompatibleGatewayProtocolFamilyResponses, prepareGrokResponsesPayload)
}

func (r *GrokTextRuntime) HandleChatCompletions(c *gin.Context, groupID *int64, body []byte) bool {
	return r.handleTextRequest(c, groupID, body, CompatibleGatewayProtocolFamilyChatCompletions, prepareGrokChatCompletionsPayload)
}

func (r *GrokTextRuntime) HandleMessages(c *gin.Context, groupID *int64, body []byte) bool {
	return r.handleTextRequest(c, groupID, body, CompatibleGatewayProtocolFamilyMessages, prepareGrokMessagesPayload)
}

func (r *GrokTextRuntime) handleTextRequest(
	c *gin.Context,
	groupID *int64,
	body []byte,
	protocolFamily CompatibleGatewayProtocolFamily,
	buildPreparedPayload func([]byte) (*grokSessionTextPreparedPayload, error),
) bool {
	if c == nil {
		return false
	}

	excludedIDs := make(map[int64]struct{})
	var lastFailoverErr *UpstreamFailoverError

	for attempt := 1; attempt <= maxRetryAttempts; attempt++ {
		preparation, handled, err := r.prepareTextRequest(
			c.Request.Context(),
			groupID,
			body,
			protocolFamily,
			buildPreparedPayload,
			excludedIDs,
		)
		if err != nil {
			if lastFailoverErr != nil && grokShouldReturnFailoverError(err) {
				writeGrokTextFailoverError(c, protocolFamily, lastFailoverErr)
				return true
			}
			writeGrokTextPreparationError(c, protocolFamily, err)
			return true
		}
		if !handled {
			return false
		}

		err = r.executePreparedText(c, preparation)
		if err == nil || c.Writer.Written() {
			return true
		}

		failoverErr := grokTextFailoverError(err)
		if failoverErr == nil || preparation.account == nil {
			writeGrokTextRuntimeError(c, protocolFamily, err)
			return true
		}

		lastFailoverErr = failoverErr
		excludedIDs[preparation.account.ID] = struct{}{}
		if len(excludedIDs) >= maxRetryAttempts {
			break
		}
	}

	writeGrokTextFailoverError(c, protocolFamily, lastFailoverErr)
	return true
}

func (r *GrokTextRuntime) prepareTextRequest(
	ctx context.Context,
	groupID *int64,
	body []byte,
	protocolFamily CompatibleGatewayProtocolFamily,
	buildPreparedPayload func([]byte) (*grokSessionTextPreparedPayload, error),
	excludedIDs map[int64]struct{},
) (*grokTextPreparation, bool, error) {
	if r == nil || r.gatewayService == nil {
		return nil, true, newGrokResponsesHTTPError(http.StatusInternalServerError, "api_error", "Grok gateway service is not configured")
	}
	if len(body) == 0 {
		return nil, true, newGrokResponsesHTTPError(http.StatusBadRequest, "invalid_request_error", "Request body is empty")
	}
	if !gjson.ValidBytes(body) {
		return nil, true, newGrokResponsesHTTPError(http.StatusBadRequest, "invalid_request_error", "Failed to parse request body")
	}

	preparedPayload, err := buildPreparedPayload(body)
	if err != nil {
		return nil, true, err
	}
	requestedModel := preparedPayload.requestedModel

	accounts, _, err := r.gatewayService.listSchedulableAccounts(ctx, groupID, PlatformGrok, true)
	if err != nil {
		return nil, true, newGrokResponsesHTTPError(http.StatusInternalServerError, "api_error", "Failed to query Grok accounts")
	}

	candidates := defaultGrokAccountSelector.FilterSchedulableCandidatesWithContext(ctx, accounts, requestedModel, excludedIDs)
	if len(candidates) == 0 {
		if !defaultGrokAccountSelector.RequestedModelAvailableWithContext(ctx, accounts, requestedModel) {
			return nil, true, newGrokResponsesHTTPError(
				http.StatusBadRequest,
				"invalid_request_error",
				"Requested model is not configured for any available Grok account: "+requestedModel,
			)
		}
		return nil, true, newGrokResponsesHTTPError(http.StatusServiceUnavailable, "api_error", "No available Grok accounts")
	}

	var loadMap map[int64]*AccountLoadInfo
	if r.gatewayService.concurrencyService != nil {
		if snapshot, err := r.gatewayService.concurrencyService.GetAccountsLoadBatch(ctx, buildAccountLoadRequests(candidates)); err == nil {
			loadMap = snapshot
		}
	}

	selected := defaultGrokAccountSelector.SelectBestCandidateWithContext(ctx, candidates, requestedModel, loadMap)
	if selected == nil {
		return nil, true, newGrokResponsesHTTPError(http.StatusServiceUnavailable, "api_error", "No available Grok accounts")
	}
	selected, err = r.hydrateSelectedAccount(ctx, selected)
	if err != nil {
		return nil, true, err
	}
	if selected.SupportsCompatibleGatewaySharedRuntime() {
		return &grokTextPreparation{
			protocolFamily: protocolFamily,
			requestedModel: requestedModel,
			account:        selected,
			compatibleBody: append([]byte(nil), body...),
			usesCompatible: true,
		}, true, nil
	}
	if selected.Type != AccountTypeSession {
		return nil, true, newGrokResponsesHTTPError(http.StatusServiceUnavailable, "api_error", "No available Grok session accounts")
	}

	runtimeSettings := DefaultGrokRuntimeSettings()
	if r.gatewayService != nil && r.gatewayService.settingService != nil {
		runtimeSettings = r.gatewayService.settingService.GetGrokRuntimeSettings(ctx)
	}
	if err := r.prepareSessionAttachments(ctx, selected, preparedPayload); err != nil {
		return nil, true, err
	}
	target, err := resolveGrokTransportTargetWithSettings(
		selected,
		r.gatewayService.validateUpstreamBaseURL,
		runtimeSettings,
	)
	if err != nil {
		return nil, true, newGrokResponsesHTTPError(http.StatusInternalServerError, "api_error", err.Error())
	}

	return &grokTextPreparation{
		protocolFamily: protocolFamily,
		requestedModel: requestedModel,
		stream:         preparedPayload.stream,
		includeUsage:   preparedPayload.includeUsage,
		toolNames:      append([]string(nil), preparedPayload.toolNames...),
		account:        selected,
		target:         target,
		payload:        preparedPayload.payload,
	}, true, nil
}

func (r *GrokTextRuntime) prepareSessionAttachments(
	ctx context.Context,
	account *Account,
	preparedPayload *grokSessionTextPreparedPayload,
) error {
	if preparedPayload == nil {
		return newGrokResponsesHTTPError(http.StatusInternalServerError, "api_error", "Failed to build Grok session request")
	}
	if preparedPayload.payloadMap == nil {
		preparedPayload.payloadMap = map[string]any{}
	}
	if len(preparedPayload.attachmentInputs) == 0 {
		if len(preparedPayload.payload) == 0 {
			payloadBytes, err := json.Marshal(preparedPayload.payloadMap)
			if err != nil {
				return newGrokResponsesHTTPError(http.StatusInternalServerError, "api_error", "Failed to build Grok session request")
			}
			preparedPayload.payload = payloadBytes
		}
		return nil
	}
	if r == nil || r.gatewayService == nil {
		return newGrokResponsesHTTPError(http.StatusInternalServerError, "api_error", "Grok gateway service is not configured")
	}

	uploader := NewGrokSessionMediaRuntime(r.gatewayService, nil, nil)
	attachments := make([]any, 0, len(preparedPayload.attachmentInputs))
	for _, input := range preparedPayload.attachmentInputs {
		uploaded, err := uploader.uploadSessionMediaInput(ctx, account, input)
		if err != nil {
			var upstreamErr *grokSessionMediaUpstreamError
			if errors.As(err, &upstreamErr) {
				return newGrokResponsesHTTPError(upstreamErr.statusCode, upstreamErr.code, upstreamErr.message)
			}
			message := sanitizeUpstreamErrorMessage(strings.TrimSpace(err.Error()))
			if message == "" {
				message = "Failed to upload Grok attachment input"
			}
			return newGrokResponsesHTTPError(http.StatusBadRequest, "invalid_request_error", message)
		}
		if uploaded == nil || strings.TrimSpace(uploaded.FileID) == "" {
			return newGrokResponsesHTTPError(http.StatusBadGateway, "api_error", "Grok attachment upload returned no attachment id")
		}
		attachments = append(attachments, strings.TrimSpace(uploaded.FileID))
	}
	preparedPayload.payloadMap["fileAttachments"] = attachments

	payloadBytes, err := json.Marshal(preparedPayload.payloadMap)
	if err != nil {
		return newGrokResponsesHTTPError(http.StatusInternalServerError, "api_error", "Failed to build Grok session request")
	}
	preparedPayload.payload = payloadBytes
	return nil
}

func (r *GrokTextRuntime) hydrateSelectedAccount(ctx context.Context, account *Account) (*Account, error) {
	if account == nil {
		return nil, newGrokResponsesHTTPError(http.StatusServiceUnavailable, "api_error", "No available Grok accounts")
	}
	if r == nil || r.gatewayService == nil {
		return nil, newGrokResponsesHTTPError(http.StatusInternalServerError, "api_error", "Grok gateway service is not configured")
	}

	hydrated, err := r.gatewayService.hydrateSelectedAccount(ctx, account)
	if err != nil {
		return nil, newGrokResponsesHTTPError(http.StatusInternalServerError, "api_error", "Failed to hydrate selected Grok account")
	}
	if hydrated == nil {
		return nil, newGrokResponsesHTTPError(http.StatusServiceUnavailable, "api_error", "No available Grok accounts")
	}
	return hydrated, nil
}

func prepareGrokResponsesPayload(body []byte) (*grokSessionTextPreparedPayload, error) {
	var req apicompat.ResponsesRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, newGrokResponsesHTTPError(http.StatusBadRequest, "invalid_request_error", "Failed to parse request body")
	}

	requestedModel := strings.TrimSpace(req.Model)
	if requestedModel == "" {
		return nil, newGrokResponsesHTTPError(http.StatusBadRequest, "invalid_request_error", "model is required")
	}

	request, err := grokSessionTextRequestFromResponsesRequest(&req)
	if err != nil {
		return nil, newGrokResponsesHTTPError(http.StatusBadRequest, "invalid_request_error", err.Error())
	}
	payload, err := buildGrokSessionTextPayload(request)
	if err != nil {
		return nil, newGrokResponsesHTTPError(http.StatusBadRequest, "invalid_request_error", err.Error())
	}

	return &grokSessionTextPreparedPayload{
		requestedModel:   requestedModel,
		stream:           req.Stream,
		toolNames:        append([]string(nil), request.ToolNames...),
		payloadMap:       payload,
		attachmentInputs: append([]grokSessionUploadInput(nil), request.AttachmentInputs...),
	}, nil
}

func prepareGrokChatCompletionsPayload(body []byte) (*grokSessionTextPreparedPayload, error) {
	var req apicompat.ChatCompletionsRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, newGrokResponsesHTTPError(http.StatusBadRequest, "invalid_request_error", "Failed to parse request body")
	}

	requestedModel := strings.TrimSpace(req.Model)
	if requestedModel == "" {
		return nil, newGrokResponsesHTTPError(http.StatusBadRequest, "invalid_request_error", "model is required")
	}

	responsesReq, err := apicompat.ChatCompletionsToResponses(&req)
	if err != nil {
		return nil, newGrokResponsesHTTPError(http.StatusBadRequest, "invalid_request_error", err.Error())
	}
	request, err := grokSessionTextRequestFromResponsesRequest(responsesReq)
	if err != nil {
		return nil, newGrokResponsesHTTPError(http.StatusBadRequest, "invalid_request_error", err.Error())
	}
	payload, err := buildGrokSessionTextPayload(request)
	if err != nil {
		return nil, newGrokResponsesHTTPError(http.StatusBadRequest, "invalid_request_error", err.Error())
	}

	includeUsage := req.StreamOptions != nil && req.StreamOptions.IncludeUsage
	return &grokSessionTextPreparedPayload{
		requestedModel:   requestedModel,
		stream:           req.Stream,
		includeUsage:     includeUsage,
		toolNames:        append([]string(nil), request.ToolNames...),
		payloadMap:       payload,
		attachmentInputs: append([]grokSessionUploadInput(nil), request.AttachmentInputs...),
	}, nil
}

func prepareGrokMessagesPayload(body []byte) (*grokSessionTextPreparedPayload, error) {
	var req apicompat.AnthropicRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, newGrokResponsesHTTPError(http.StatusBadRequest, "invalid_request_error", "Failed to parse request body")
	}

	requestedModel := strings.TrimSpace(req.Model)
	if requestedModel == "" {
		return nil, newGrokResponsesHTTPError(http.StatusBadRequest, "invalid_request_error", "model is required")
	}

	responsesReq, err := apicompat.AnthropicToResponses(&req)
	if err != nil {
		return nil, newGrokResponsesHTTPError(http.StatusBadRequest, "invalid_request_error", err.Error())
	}
	request, err := grokSessionTextRequestFromResponsesRequest(responsesReq)
	if err != nil {
		return nil, newGrokResponsesHTTPError(http.StatusBadRequest, "invalid_request_error", err.Error())
	}
	payload, err := buildGrokSessionTextPayload(request)
	if err != nil {
		return nil, newGrokResponsesHTTPError(http.StatusBadRequest, "invalid_request_error", err.Error())
	}

	return &grokSessionTextPreparedPayload{
		requestedModel:   requestedModel,
		stream:           req.Stream,
		toolNames:        append([]string(nil), request.ToolNames...),
		payloadMap:       payload,
		attachmentInputs: append([]grokSessionUploadInput(nil), request.AttachmentInputs...),
	}, nil
}

func grokResponsesErrorCodeForStatus(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "invalid_request_error"
	case http.StatusUnauthorized, http.StatusForbidden:
		return "authentication_error"
	case http.StatusTooManyRequests:
		return "rate_limit_error"
	default:
		return "api_error"
	}
}

func writeGrokTextPreparationError(c *gin.Context, protocolFamily CompatibleGatewayProtocolFamily, err error) {
	if c == nil {
		return
	}
	var httpErr *grokResponsesHTTPError
	if errors.As(err, &httpErr) {
		writeGrokTextError(c, protocolFamily, httpErr.statusCode, httpErr.code, httpErr.message)
		return
	}
	writeGrokTextError(c, protocolFamily, http.StatusInternalServerError, "api_error", "Grok runtime request failed")
}

func (r *GrokTextRuntime) executePreparedText(c *gin.Context, preparation *grokTextPreparation) error {
	if preparation == nil {
		return newGrokResponsesHTTPError(http.StatusServiceUnavailable, "api_error", "No available Grok accounts")
	}
	if preparation.usesCompatible {
		if r.compatibleRuntime == nil {
			return newGrokResponsesHTTPError(http.StatusInternalServerError, "api_error", "Grok compatible runtime is not configured")
		}
		return r.compatibleRuntime.Execute(c, preparation)
	}
	if r.sessionRuntime == nil {
		return newGrokResponsesHTTPError(http.StatusInternalServerError, "api_error", "Grok session runtime is not configured")
	}
	return r.sessionRuntime.Execute(c, preparation)
}

func grokTextFailoverError(err error) *UpstreamFailoverError {
	var failoverErr *UpstreamFailoverError
	if errors.As(err, &failoverErr) {
		return failoverErr
	}
	return nil
}

func grokShouldReturnFailoverError(err error) bool {
	var httpErr *grokResponsesHTTPError
	if !errors.As(err, &httpErr) || httpErr == nil {
		return false
	}
	if httpErr.statusCode == http.StatusServiceUnavailable {
		return true
	}
	if httpErr.statusCode != http.StatusBadRequest {
		return false
	}

	message := strings.ToLower(strings.TrimSpace(httpErr.message))
	return strings.Contains(message, "requested model is not configured for any available grok account") ||
		strings.Contains(message, "no available grok accounts")
}

func writeGrokTextRuntimeError(c *gin.Context, protocolFamily CompatibleGatewayProtocolFamily, err error) {
	if c == nil || c.Writer.Written() {
		return
	}
	var httpErr *grokResponsesHTTPError
	if errors.As(err, &httpErr) && httpErr != nil {
		writeGrokTextError(c, protocolFamily, httpErr.statusCode, httpErr.code, httpErr.message)
		return
	}
	message := sanitizeUpstreamErrorMessage(strings.TrimSpace(err.Error()))
	if message == "" {
		message = "Grok runtime request failed"
	}
	writeGrokTextError(c, protocolFamily, http.StatusBadGateway, "api_error", message)
}

func writeGrokTextFailoverError(c *gin.Context, protocolFamily CompatibleGatewayProtocolFamily, failoverErr *UpstreamFailoverError) {
	if c == nil || c.Writer.Written() {
		return
	}
	statusCode := http.StatusBadGateway
	code := "api_error"
	message := "Grok upstream request failed"
	if failoverErr != nil {
		if failoverErr.StatusCode > 0 {
			statusCode = firstNonZero(mapUpstreamStatusCode(failoverErr.StatusCode), http.StatusBadGateway)
			code = grokResponsesErrorCodeForStatus(failoverErr.StatusCode)
		}
		if extracted := sanitizeUpstreamErrorMessage(strings.TrimSpace(ExtractUpstreamErrorMessage(failoverErr.ResponseBody))); extracted != "" {
			message = extracted
		} else if reason := sanitizeUpstreamErrorMessage(strings.TrimSpace(failoverErr.FailureReason)); reason != "" {
			message = reason
		}
	}
	writeGrokTextError(c, protocolFamily, statusCode, code, message)
}

func writeGrokTextError(c *gin.Context, protocolFamily CompatibleGatewayProtocolFamily, statusCode int, code, message string) {
	if c == nil {
		return
	}
	switch protocolFamily {
	case CompatibleGatewayProtocolFamilyChatCompletions:
		writeChatCompletionsError(c, passthroughRuleResult{
			StatusCode: statusCode,
			ErrType:    code,
			ErrMessage: message,
		})
	case CompatibleGatewayProtocolFamilyMessages:
		writeAnthropicError(c, passthroughRuleResult{
			StatusCode: statusCode,
			ErrType:    code,
			ErrMessage: message,
		})
	default:
		writeResponsesError(c, statusCode, code, message)
	}
}
