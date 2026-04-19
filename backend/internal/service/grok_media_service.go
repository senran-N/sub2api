package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/util/responseheaders"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type GrokMediaService struct {
	gatewayService *GatewayService
	videoJobs      GrokVideoJobRepository
	mediaAssets    *GrokMediaAssetService
}

type grokMediaForwardRequest struct {
	account            *Account
	body               []byte
	requestMeta        openAICompatiblePassthroughRequestMeta
	defaultMappedModel string
}

type grokMediaForwardResponse struct {
	StatusCode    int
	Header        http.Header
	Body          []byte
	UpstreamModel string
}

func NewGrokMediaService(
	gatewayService *GatewayService,
	videoJobs GrokVideoJobRepository,
	mediaAssets GrokMediaAssetRepository,
) *GrokMediaService {
	return &GrokMediaService{
		gatewayService: gatewayService,
		videoJobs:      videoJobs,
		mediaAssets:    NewGrokMediaAssetService(gatewayService, mediaAssets),
	}
}

func (s *GrokMediaService) HandleImages(c *gin.Context, groupID *int64, body []byte) bool {
	if c == nil {
		return false
	}

	reqMeta := GetOpenAICompatiblePassthroughRequestMeta(c, body)
	requestedModel, schedulingModel, restricted := s.resolveMappedModel(c.Request.Context(), groupID, reqMeta.Model)
	if restricted {
		writeCompatibleGatewayMediaError(c, http.StatusBadRequest, "invalid_request_error", "Requested model is not allowed for this channel")
		return true
	}

	account, err := s.selectCompatibleAccount(c.Request.Context(), groupID, schedulingModel)
	if err != nil {
		s.writeSelectionError(c, requestedModel, schedulingModel, err)
		return true
	}

	resp, err := s.forwardCompatibleRequest(c, grokMediaForwardRequest{
		account:            account,
		body:               body,
		requestMeta:        reqMeta,
		defaultMappedModel: schedulingModel,
	})
	if err != nil {
		writeCompatibleGatewayMediaError(c, http.StatusBadGateway, "api_error", "Grok upstream request failed")
		return true
	}
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		resp.Body, _, _ = s.rewriteMediaResponse(c, account, requestedModel, firstNonEmpty(resp.UpstreamModel, schedulingModel, requestedModel), "", "image", resp.Body)
	}
	s.writeForwardResponse(c, resp)
	return true
}

func (s *GrokMediaService) HandleVideos(c *gin.Context, groupID *int64, body []byte) bool {
	if c == nil || c.Request == nil || c.Request.URL == nil {
		return false
	}

	jobID, hasJobID, contentRequest := resolveGrokVideoRoute(c.Request.URL.Path)
	if !hasJobID {
		return s.handleVideoCreate(c, groupID, body)
	}
	return s.handleVideoFollowup(c, jobID, contentRequest, body)
}

func (s *GrokMediaService) handleVideoCreate(c *gin.Context, groupID *int64, body []byte) bool {
	reqMeta := GetOpenAICompatiblePassthroughRequestMeta(c, body)
	requestedModel, schedulingModel, restricted := s.resolveMappedModel(c.Request.Context(), groupID, reqMeta.Model)
	if restricted {
		writeCompatibleGatewayMediaError(c, http.StatusBadRequest, "invalid_request_error", "Requested model is not allowed for this channel")
		return true
	}

	account, err := s.selectCompatibleAccount(c.Request.Context(), groupID, schedulingModel)
	if err != nil {
		s.writeSelectionError(c, requestedModel, schedulingModel, err)
		return true
	}

	resp, err := s.forwardCompatibleRequest(c, grokMediaForwardRequest{
		account:            account,
		body:               body,
		requestMeta:        reqMeta,
		defaultMappedModel: schedulingModel,
	})
	if err != nil {
		writeCompatibleGatewayMediaError(c, http.StatusBadGateway, "api_error", "Grok upstream request failed")
		return true
	}

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices && s.videoJobs != nil {
		jobID := extractGrokVideoJobID(resp.Body)
		if jobID != "" {
			rewrittenBody, outputAssetID, _ := s.rewriteMediaResponse(c, account, requestedModel, firstNonEmpty(resp.UpstreamModel, schedulingModel, requestedModel), jobID, "video", resp.Body)
			resp.Body = rewrittenBody
			_ = s.videoJobs.Upsert(c.Request.Context(), GrokVideoJobRecord{
				JobID:                  jobID,
				AccountID:              account.ID,
				GroupID:                groupID,
				RequestedModel:         requestedModel,
				CanonicalModel:         firstNonEmpty(resp.UpstreamModel, schedulingModel, requestedModel),
				OutputAssetID:          outputAssetID,
				RequestPayloadSnapshot: cloneJSONBody(body),
				UpstreamStatus:         extractGrokVideoStatus(resp.Body),
				NormalizedStatus:       normalizeGrokVideoStatus(extractGrokVideoStatus(resp.Body)),
				PollAfter:              extractGrokVideoPollAfter(resp.Body),
				ErrorCode:              extractUpstreamErrorCode(resp.Body),
				ErrorMessage:           ExtractUpstreamErrorMessage(resp.Body),
			})
		}
	} else if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		resp.Body, _, _ = s.rewriteMediaResponse(c, account, requestedModel, firstNonEmpty(resp.UpstreamModel, schedulingModel, requestedModel), "", "video", resp.Body)
	}

	s.writeForwardResponse(c, resp)
	return true
}

func (s *GrokMediaService) handleVideoFollowup(c *gin.Context, jobID string, contentRequest bool, body []byte) bool {
	if s == nil || s.videoJobs == nil {
		writeCompatibleGatewayMediaError(c, http.StatusNotFound, "not_found_error", "Grok video job binding is not configured")
		return true
	}

	record, err := s.videoJobs.GetByJobID(c.Request.Context(), jobID)
	if err != nil {
		if errors.Is(err, ErrGrokVideoJobNotFound) {
			writeCompatibleGatewayMediaError(c, http.StatusNotFound, "not_found_error", "Grok video job is not known to this gateway")
			return true
		}
		writeCompatibleGatewayMediaError(c, http.StatusInternalServerError, "api_error", "Failed to load Grok video job binding")
		return true
	}
	if record == nil {
		writeCompatibleGatewayMediaError(c, http.StatusNotFound, "not_found_error", "Grok video job is not known to this gateway")
		return true
	}

	account, err := s.gatewayService.accountRepo.GetByID(c.Request.Context(), record.AccountID)
	if err != nil || account == nil {
		writeCompatibleGatewayMediaError(c, http.StatusServiceUnavailable, "api_error", "Bound Grok video account is unavailable")
		return true
	}
	if !account.SupportsCompatibleGatewaySharedRuntime() {
		writeCompatibleGatewayMediaError(c, http.StatusServiceUnavailable, "api_error", "Bound Grok video account does not support provider media replay")
		return true
	}

	resp, err := s.forwardCompatibleRequest(c, grokMediaForwardRequest{
		account:            account,
		body:               body,
		requestMeta:        GetOpenAICompatiblePassthroughRequestMeta(c, body),
		defaultMappedModel: record.CanonicalModel,
	})
	if err != nil {
		writeCompatibleGatewayMediaError(c, http.StatusBadGateway, "api_error", "Grok upstream request failed")
		return true
	}

	if !contentRequest {
		rewrittenBody, outputAssetID, _ := s.rewriteMediaResponse(c, account, record.RequestedModel, record.CanonicalModel, record.JobID, "video", resp.Body)
		resp.Body = rewrittenBody
		_ = s.videoJobs.UpdateStatus(c.Request.Context(), GrokVideoJobStatusPatch{
			JobID:            record.JobID,
			UpstreamStatus:   extractGrokVideoStatus(resp.Body),
			NormalizedStatus: normalizeGrokVideoStatus(extractGrokVideoStatus(resp.Body)),
			PollAfter:        extractGrokVideoPollAfter(resp.Body),
			ErrorCode:        extractUpstreamErrorCode(resp.Body),
			ErrorMessage:     ExtractUpstreamErrorMessage(resp.Body),
			OutputAssetID:    outputAssetID,
		})
	} else {
		resp.Body, _, _ = s.rewriteMediaResponse(c, account, record.RequestedModel, record.CanonicalModel, record.JobID, "video", resp.Body)
	}

	s.writeForwardResponse(c, resp)
	return true
}

func (s *GrokMediaService) HandleAssetContent(c *gin.Context, assetID string) bool {
	if s == nil || s.mediaAssets == nil {
		return false
	}
	return s.mediaAssets.Serve(c, assetID)
}

func (s *GrokMediaService) selectCompatibleAccount(ctx context.Context, groupID *int64, requestedModel string) (*Account, error) {
	if s == nil || s.gatewayService == nil {
		return nil, errors.New("grok media service is not configured")
	}

	accounts, _, err := s.gatewayService.listSchedulableAccounts(ctx, groupID, PlatformGrok, true)
	if err != nil {
		return nil, err
	}

	candidates := defaultGrokAccountSelector.FilterSchedulableCandidatesWithContext(ctx, accounts, requestedModel, nil)
	if len(candidates) == 0 {
		if !defaultGrokAccountSelector.RequestedModelAvailableWithContext(ctx, accounts, requestedModel) {
			return nil, fmt.Errorf("requested model unavailable:%s", requestedModel)
		}
		return nil, errors.New("no compatible grok media accounts")
	}

	var loadMap map[int64]*AccountLoadInfo
	if s.gatewayService.concurrencyService != nil {
		if snapshot, loadErr := s.gatewayService.concurrencyService.GetAccountsLoadBatch(ctx, buildAccountLoadRequests(candidates)); loadErr == nil {
			loadMap = snapshot
		}
	}

	selected := defaultGrokAccountSelector.SelectBestCandidateWithContext(ctx, candidates, requestedModel, loadMap)
	if selected == nil {
		return nil, errors.New("no compatible grok media accounts")
	}
	if !selected.SupportsCompatibleGatewaySharedRuntime() {
		return nil, errors.New("selected grok account does not support compatible media transport")
	}
	return selected, nil
}

func (s *GrokMediaService) resolveMappedModel(ctx context.Context, groupID *int64, requestedModel string) (string, string, bool) {
	requestedModel = strings.TrimSpace(requestedModel)
	if s == nil || s.gatewayService == nil || requestedModel == "" {
		return requestedModel, requestedModel, false
	}

	channelMapping, restricted := s.gatewayService.ResolveChannelMappingAndRestrict(ctx, groupID, requestedModel)
	if restricted {
		return requestedModel, requestedModel, true
	}
	if mapped := strings.TrimSpace(channelMapping.MappedModel); mapped != "" {
		return requestedModel, mapped, false
	}
	return requestedModel, requestedModel, false
}

func (s *GrokMediaService) forwardCompatibleRequest(c *gin.Context, input grokMediaForwardRequest) (*grokMediaForwardResponse, error) {
	if s == nil || s.gatewayService == nil || s.gatewayService.httpUpstream == nil {
		return nil, errors.New("grok media service is not configured")
	}
	if c == nil || c.Request == nil || input.account == nil {
		return nil, errors.New("invalid grok media forward request")
	}

	token := strings.TrimSpace(input.account.GetOpenAIApiKey())
	if token == "" {
		return nil, errors.New("api_key not found in credentials")
	}

	baseURL := strings.TrimSpace(input.account.GetOpenAIBaseURL())
	if baseURL == "" {
		baseURL = CompatibleGatewayDefaultBaseURL(input.account.Platform)
	}
	validatedBaseURL, err := s.gatewayService.validateUpstreamBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	mappedModel := resolveOpenAIForwardModel(input.account, input.requestMeta.Model, input.defaultMappedModel)
	mappedModel = normalizeOpenAIModelForUpstream(input.account, mappedModel)
	forwardBody := input.body
	if input.requestMeta.JSONBody && input.requestMeta.Model != "" && mappedModel != "" && mappedModel != input.requestMeta.Model {
		patchedBody, err := sjson.SetBytes(input.body, "model", mappedModel)
		if err != nil {
			return nil, fmt.Errorf("patch media model: %w", err)
		}
		forwardBody = patchedBody
	}

	upstreamTarget := newCompatiblePassthroughUpstreamTargetWithOptions(
		validatedBaseURL,
		normalizeGrokMediaUpstreamPath(c.Request.URL.Path),
		input.account.GetCompatibleAuthMode(""),
		input.account.GetCompatibleEndpointOverride("responses"),
		input.account.GetCompatibleEndpointOverride("chat_completions"),
	)
	targetURL := upstreamTarget.URL
	if rawQuery := strings.TrimSpace(c.Request.URL.RawQuery); rawQuery != "" {
		parsedTarget, err := url.Parse(targetURL)
		if err != nil {
			return nil, err
		}
		parsedTarget.RawQuery = rawQuery
		targetURL = parsedTarget.String()
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, targetURL, bytes.NewReader(forwardBody))
	if err != nil {
		return nil, err
	}

	allowTimeoutHeaders := s.gatewayService.cfg != nil && s.gatewayService.cfg.Gateway.OpenAIPassthroughAllowTimeoutHeaders
	for key, values := range c.Request.Header {
		lower := strings.ToLower(strings.TrimSpace(key))
		if !isOpenAIPassthroughAllowedRequestHeader(lower, allowTimeoutHeaders) {
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	req.Header.Del("authorization")
	req.Header.Del("api-key")
	req.Header.Del("x-api-key")
	upstreamTarget.ApplyAuthHeader(req.Header, token)
	if req.Header.Get("content-type") == "" && len(forwardBody) > 0 {
		req.Header.Set("content-type", "application/json")
	}

	proxyURL := ""
	if input.account.Proxy != nil {
		proxyURL = input.account.Proxy.URL()
	}
	resp, err := s.gatewayService.httpUpstream.DoWithTLS(
		req,
		proxyURL,
		input.account.ID,
		input.account.Concurrency,
		resolveGrokGatewayTLSProfile(s.gatewayService, input.account),
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &grokMediaForwardResponse{
		StatusCode:    resp.StatusCode,
		Header:        resp.Header.Clone(),
		Body:          respBody,
		UpstreamModel: firstNonEmpty(mappedModel, input.defaultMappedModel, input.requestMeta.Model),
	}, nil
}

func (s *GrokMediaService) writeSelectionError(c *gin.Context, requestedModel, schedulingModel string, err error) {
	if c == nil {
		return
	}
	if err == nil {
		writeCompatibleGatewayMediaError(c, http.StatusServiceUnavailable, "api_error", "No available Grok media accounts")
		return
	}
	if strings.HasPrefix(err.Error(), "requested model unavailable:") {
		model := strings.TrimPrefix(err.Error(), "requested model unavailable:")
		model = strings.TrimSpace(firstNonEmpty(model, schedulingModel, requestedModel))
		writeCompatibleGatewayMediaError(c, http.StatusBadRequest, "invalid_request_error", "Requested model is not configured for any available Grok account: "+model)
		return
	}
	writeCompatibleGatewayMediaError(c, http.StatusServiceUnavailable, "api_error", "No available Grok media accounts")
}

func (s *GrokMediaService) writeForwardResponse(c *gin.Context, resp *grokMediaForwardResponse) {
	if c == nil || resp == nil {
		return
	}
	responseheaders.WriteFilteredHeaders(c.Writer.Header(), resp.Header, s.gatewayService.responseHeaderFilter)
	c.Status(resp.StatusCode)
	if len(resp.Body) == 0 {
		return
	}
	_, _ = c.Writer.Write(resp.Body)
}

func (s *GrokMediaService) rewriteMediaResponse(
	c *gin.Context,
	account *Account,
	requestedModel string,
	canonicalModel string,
	jobID string,
	assetType string,
	body []byte,
) ([]byte, string, error) {
	if s == nil || s.mediaAssets == nil {
		return body, "", nil
	}
	return s.mediaAssets.RewriteResponse(c, account, body, assetType, requestedModel, canonicalModel, jobID)
}

func writeCompatibleGatewayMediaError(c *gin.Context, statusCode int, code, message string) {
	if c == nil {
		return
	}
	c.JSON(statusCode, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}

func resolveGrokVideoRoute(path string) (jobID string, hasJobID bool, contentRequest bool) {
	trimmed := strings.Trim(strings.TrimSpace(path), "/")
	if trimmed == "" {
		return "", false, false
	}
	parts := strings.Split(trimmed, "/")
	for i := 0; i < len(parts); i++ {
		if parts[i] != "videos" {
			continue
		}
		if i+1 >= len(parts) || strings.TrimSpace(parts[i+1]) == "" {
			return "", false, false
		}
		jobID = strings.TrimSpace(parts[i+1])
		contentRequest = i+2 < len(parts) && strings.TrimSpace(parts[i+2]) == "content"
		return jobID, true, contentRequest
	}
	return "", false, false
}

func normalizeGrokMediaUpstreamPath(path string) string {
	trimmed := strings.TrimSpace(path)
	switch {
	case trimmed == "/grok/v1":
		return "/v1"
	case strings.HasPrefix(trimmed, "/grok/v1/"):
		return "/v1/" + strings.TrimPrefix(trimmed, "/grok/v1/")
	default:
		return trimmed
	}
}

func extractGrokVideoJobID(body []byte) string {
	for _, path := range []string{"job_id", "id", "data.job_id", "data.id"} {
		if value := strings.TrimSpace(gjson.GetBytes(body, path).String()); value != "" {
			return value
		}
	}
	return ""
}

func extractGrokVideoStatus(body []byte) string {
	for _, path := range []string{"status", "data.status", "job.status"} {
		if value := strings.TrimSpace(gjson.GetBytes(body, path).String()); value != "" {
			return value
		}
	}
	return ""
}

func normalizeGrokVideoStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "queued", "pending", "submitted":
		return "queued"
	case "running", "processing", "in_progress":
		return "in_progress"
	case "completed", "succeeded", "success":
		return "completed"
	case "failed", "error", "cancelled", "canceled":
		return "failed"
	default:
		return strings.ToLower(strings.TrimSpace(status))
	}
}

func extractGrokVideoPollAfter(body []byte) *time.Time {
	for _, path := range []string{"poll_after", "data.poll_after", "job.poll_after"} {
		value := gjson.GetBytes(body, path)
		if !value.Exists() {
			continue
		}
		switch value.Type {
		case gjson.Number:
			seconds := value.Int()
			if seconds <= 0 {
				continue
			}
			ts := time.Now().UTC().Add(time.Duration(seconds) * time.Second)
			return &ts
		case gjson.String:
			text := strings.TrimSpace(value.String())
			if text == "" {
				continue
			}
			if parsed, err := time.Parse(time.RFC3339, text); err == nil {
				ts := parsed.UTC()
				return &ts
			}
			if seconds, err := strconv.ParseInt(text, 10, 64); err == nil && seconds > 0 {
				ts := time.Now().UTC().Add(time.Duration(seconds) * time.Second)
				return &ts
			}
		}
	}
	return nil
}

func cloneJSONBody(body []byte) []byte {
	if !gjson.ValidBytes(body) {
		return nil
	}
	cloned := append([]byte(nil), body...)
	if json.Valid(cloned) {
		return cloned
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
