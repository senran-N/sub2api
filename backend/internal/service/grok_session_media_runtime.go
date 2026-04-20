package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

const (
	grokSessionUploadFileEndpoint      = "/rest/app-chat/upload-file"
	grokSessionMediaPostCreateEndpoint = "/rest/media/post/create"

	grokSessionMediaPostTypeImage = "MEDIA_POST_TYPE_IMAGE"
	grokSessionMediaPostTypeVideo = "MEDIA_POST_TYPE_VIDEO"

	grokSessionImageEditModelName         = "imagine-image-edit"
	grokSessionImageEditModelKind         = "imagine"
	grokSessionVideoModelName             = "grok-3"
	grokSessionVideoExtensionRefType      = "ORIGINAL_REF_TYPE_VIDEO_EXTENSION"
	grokSessionDefaultImageEditCount      = 2
	grokSessionDefaultVideoSeconds        = 6
	grokSessionDefaultVideoSize           = "720x1280"
	grokSessionDefaultVideoQuality        = "standard"
	grokSessionDefaultVideoResolutionName = "720p"
	grokSessionDefaultVideoPreset         = "custom"
	grokSessionVideoPollInterval          = 5 * time.Second
	grokSessionUploadedAssetBaseURL       = "https://assets.grok.com"
)

var (
	errGrokSessionVideoReferenceFileIDUnsupported = errors.New("input_reference.file_id is not supported yet")
	grokSessionVideoSizeMap                       = map[string][2]string{
		"720x1280":  {"9:16", "720p"},
		"1280x720":  {"16:9", "720p"},
		"1024x1024": {"1:1", "720p"},
		"1024x1792": {"9:16", "720p"},
		"1792x1024": {"16:9", "720p"},
	}
	grokSessionVideoPresetFlags = map[string]string{
		"fun":    "--mode=extremely-crazy",
		"normal": "--mode=normal",
		"spicy":  "--mode=extremely-spicy-or-crazy",
		"custom": "--mode=custom",
	}
)

type GrokSessionMediaRuntime struct {
	gatewayService *GatewayService
	videoJobs      GrokVideoJobRepository
	mediaAssets    *GrokMediaAssetService
	spawn          func(func())
	now            func() time.Time
}

type grokSessionImageGenerationRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	N              int    `json:"n"`
	ResponseFormat string `json:"response_format"`
}

type grokSessionImageEditRequest struct {
	Model          string
	Prompt         string
	N              int
	ResponseFormat string
	InputImages    []grokSessionUploadInput
	HasMask        bool
	MaskProvided   string
}

type grokSessionVideoCreateRequest struct {
	Model          string                     `json:"model"`
	Prompt         string                     `json:"prompt"`
	Seconds        int                        `json:"seconds"`
	Size           string                     `json:"size"`
	Quality        string                     `json:"quality"`
	Preset         string                     `json:"preset"`
	ResolutionName string                     `json:"resolution_name"`
	InputReference *grokSessionInputReference `json:"input_reference"`
}

type grokSessionInputReference struct {
	FileID   string `json:"file_id"`
	ImageURL string `json:"image_url"`
}

type grokSessionUploadInput struct {
	FileName string
	MimeType string
	Base64   string
	Source   string
}

type grokSessionUploadedAsset struct {
	FileID     string
	FileURI    string
	ContentURL string
}

type grokSessionVideoReference struct {
	ContentURL string
	PostID     string
}

type grokSessionVideoArtifact struct {
	VideoURL     string
	VideoPostID  string
	AssetID      string
	ThumbnailURL string
}

type grokSessionMediaUpstreamError struct {
	statusCode int
	code       string
	message    string
}

func (e *grokSessionMediaUpstreamError) Error() string {
	if e == nil {
		return ""
	}
	return e.message
}

func NewGrokSessionMediaRuntime(
	gatewayService *GatewayService,
	videoJobs GrokVideoJobRepository,
	mediaAssets *GrokMediaAssetService,
) *GrokSessionMediaRuntime {
	return &GrokSessionMediaRuntime{
		gatewayService: gatewayService,
		videoJobs:      videoJobs,
		mediaAssets:    mediaAssets,
		spawn: func(fn func()) {
			go fn()
		},
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (r *GrokSessionMediaRuntime) HandleImages(
	c *gin.Context,
	account *Account,
	requestedModel string,
	canonicalModel string,
	body []byte,
) bool {
	if c == nil {
		return false
	}
	if account == nil || account.Type != AccountTypeSession {
		writeCompatibleGatewayMediaError(c, http.StatusServiceUnavailable, "api_error", "No available Grok session media account")
		return true
	}

	if grokSessionImageRouteIsEdit(c.Request.URL.Path) {
		return r.handleImageEdit(c, account, requestedModel, canonicalModel, body)
	}
	return r.handleImageGeneration(c, account, requestedModel, canonicalModel, body)
}

func (r *GrokSessionMediaRuntime) HandleVideoCreate(
	c *gin.Context,
	groupID *int64,
	account *Account,
	requestedModel string,
	canonicalModel string,
	body []byte,
) bool {
	if c == nil {
		return false
	}
	if r == nil || r.videoJobs == nil {
		writeCompatibleGatewayMediaError(c, http.StatusInternalServerError, "api_error", "Grok session video runtime is not configured")
		return true
	}
	if account == nil || account.Type != AccountTypeSession {
		writeCompatibleGatewayMediaError(c, http.StatusServiceUnavailable, "api_error", "No available Grok session video account")
		return true
	}

	req, err := parseGrokSessionVideoCreateRequest(body, firstNonEmpty(canonicalModel, requestedModel))
	if err != nil {
		writeCompatibleGatewayMediaError(c, http.StatusBadRequest, "invalid_request_error", err.Error())
		return true
	}

	now := r.now()
	pollAfter := now.Add(grokSessionVideoPollInterval)
	jobID := "job_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	record := GrokVideoJobRecord{
		JobID:                  jobID,
		AccountID:              account.ID,
		GroupID:                groupID,
		RequestedModel:         requestedModel,
		CanonicalModel:         firstNonEmpty(canonicalModel, requestedModel, req.Model),
		RequestPayloadSnapshot: cloneJSONBody(body),
		UpstreamStatus:         "queued",
		NormalizedStatus:       "queued",
		PollAfter:              &pollAfter,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
	if err := r.videoJobs.Upsert(c.Request.Context(), record); err != nil {
		writeCompatibleGatewayMediaError(c, http.StatusInternalServerError, "api_error", "Failed to create Grok session video job")
		return true
	}

	if r.spawn != nil {
		recordCopy := record
		reqCopy := req
		r.spawn(func() {
			r.runVideoJob(account, &recordCopy, reqCopy)
		})
	}

	c.JSON(http.StatusOK, buildGrokSessionVideoJobResponse(record, req))
	return true
}

func (r *GrokSessionMediaRuntime) HandleVideoFollowup(
	c *gin.Context,
	_ *Account,
	record *GrokVideoJobRecord,
	contentRequest bool,
) bool {
	if c == nil {
		return false
	}
	if record == nil {
		writeCompatibleGatewayMediaError(c, http.StatusNotFound, "not_found_error", "Grok video job is not known to this gateway")
		return true
	}

	if contentRequest {
		if strings.TrimSpace(record.OutputAssetID) == "" {
			status := normalizeGrokVideoStatus(record.NormalizedStatus)
			if status == "" {
				status = normalizeGrokVideoStatus(record.UpstreamStatus)
			}
			if status != "completed" {
				writeCompatibleGatewayMediaError(c, http.StatusConflict, "invalid_request_error", "Grok video content is not ready yet")
				return true
			}
			writeCompatibleGatewayMediaError(c, http.StatusNotFound, "not_found_error", "Grok video content is not known to this gateway")
			return true
		}
		if r == nil || r.mediaAssets == nil {
			writeCompatibleGatewayMediaError(c, http.StatusInternalServerError, "api_error", "Grok media asset service is not configured")
			return true
		}
		rendered, upstreamURL, err := r.mediaAssets.RenderExistingAssetValue(c, record.OutputAssetID, "video")
		if err != nil {
			if errors.Is(err, ErrGrokMediaAssetNotFound) {
				writeCompatibleGatewayMediaError(c, http.StatusNotFound, "not_found_error", "Grok video content is not known to this gateway")
				return true
			}
			writeCompatibleGatewayMediaError(c, http.StatusBadGateway, "api_error", "Failed to render Grok video content")
			return true
		}
		c.JSON(http.StatusOK, gin.H{
			"id":                   strings.TrimSpace(record.JobID),
			"object":               "video.content",
			"status":               firstNonEmpty(strings.TrimSpace(record.NormalizedStatus), normalizeGrokVideoStatus(record.UpstreamStatus), "completed"),
			"content_url":          rendered,
			"upstream_content_url": upstreamURL,
		})
		return true
	}

	req, _ := parseGrokSessionVideoCreateRequest(record.RequestPayloadSnapshot, firstNonEmpty(record.CanonicalModel, record.RequestedModel))
	c.JSON(http.StatusOK, buildGrokSessionVideoJobResponse(*record, req))
	return true
}

func (r *GrokSessionMediaRuntime) handleImageGeneration(
	c *gin.Context,
	account *Account,
	requestedModel string,
	canonicalModel string,
	body []byte,
) bool {
	req, err := parseGrokSessionImageGenerationRequest(body, firstNonEmpty(canonicalModel, requestedModel))
	if err != nil {
		writeCompatibleGatewayMediaError(c, http.StatusBadRequest, "invalid_request_error", err.Error())
		return true
	}

	modeID := resolveGrokSessionMediaModeID(req.Model)
	payload, err := buildGrokSessionTextPayload(grokSessionTextRequest{
		ModelID: req.Model,
		ModeID:  modeID,
		Message: "Drawing: " + strings.TrimSpace(req.Prompt),
	})
	if err != nil {
		writeCompatibleGatewayMediaError(c, http.StatusBadRequest, "invalid_request_error", err.Error())
		return true
	}

	payload["disableTextFollowUps"] = true
	payload["enableImageGeneration"] = true
	payload["enableImageStreaming"] = true
	payload["imageGenerationCount"] = req.N
	payload["toolOverrides"] = map[string]any{"imageGen": true}
	payload["returnImageBytes"] = false
	payload["returnRawGrokInXaiRequest"] = false

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		writeCompatibleGatewayMediaError(c, http.StatusInternalServerError, "api_error", "Failed to build Grok image generation request")
		return true
	}

	resp, target, err := r.doSessionJSONRequest(
		c.Request.Context(),
		account,
		grokSessionConversationEndpoint,
		payloadBytes,
		"application/json, text/event-stream, text/plain, */*",
		0,
	)
	if err != nil {
		r.persistSessionMediaRuntimeFeedback(c.Request.Context(), account, requestedModel, c.Request.URL.Path, err)
		r.writeMediaRuntimeError(c, err)
		return true
	}
	defer func() { _ = resp.Body.Close() }()

	imageURLs, err := collectGrokSessionImageURLs(resp, target, c.Request.URL.Path)
	if err != nil {
		r.persistSessionMediaRuntimeFeedback(c.Request.Context(), account, requestedModel, c.Request.URL.Path, err)
		r.writeMediaRuntimeError(c, err)
		return true
	}
	if len(imageURLs) == 0 {
		r.persistSessionMediaRuntimeFeedback(c.Request.Context(), account, requestedModel, c.Request.URL.Path, errors.New("grok session image generation returned no images"))
		writeCompatibleGatewayMediaError(c, http.StatusBadGateway, "api_error", "Grok session image generation returned no images")
		return true
	}

	responseBody, err := marshalGrokSessionImageResponse(imageURLs, r.now())
	if err != nil {
		writeCompatibleGatewayMediaError(c, http.StatusInternalServerError, "api_error", "Failed to build Grok image response")
		return true
	}
	if r.mediaAssets != nil {
		responseBody, _, _ = r.mediaAssets.RewriteResponse(c, account, responseBody, "image", req.ResponseFormat, requestedModel, canonicalModel, "")
	}
	r.persistSessionMediaRuntimeFeedback(c.Request.Context(), account, requestedModel, c.Request.URL.Path, nil)
	c.Data(http.StatusOK, "application/json", responseBody)
	return true
}

func (r *GrokSessionMediaRuntime) handleImageEdit(
	c *gin.Context,
	account *Account,
	requestedModel string,
	canonicalModel string,
	body []byte,
) bool {
	req, err := parseGrokSessionImageEditRequest(c, body, firstNonEmpty(canonicalModel, requestedModel))
	if err != nil {
		writeCompatibleGatewayMediaError(c, http.StatusBadRequest, "invalid_request_error", err.Error())
		return true
	}
	if req.HasMask {
		writeCompatibleGatewayMediaError(c, http.StatusBadRequest, "invalid_request_error", "Grok session image edit does not support mask uploads")
		return true
	}

	uploaded, err := r.uploadSessionMediaInputs(c.Request.Context(), account, req.InputImages)
	if err != nil {
		r.persistSessionMediaRuntimeFeedback(c.Request.Context(), account, requestedModel, c.Request.URL.Path, err)
		r.writeMediaRuntimeError(c, err)
		return true
	}
	imageReferences := collectGrokSessionUploadedContentURLs(uploaded)
	if len(imageReferences) == 0 {
		err = errors.New("grok session image edit has no uploaded image references")
		r.persistSessionMediaRuntimeFeedback(c.Request.Context(), account, requestedModel, c.Request.URL.Path, err)
		r.writeMediaRuntimeError(c, err)
		return true
	}

	postID, err := r.createSessionMediaPost(
		c.Request.Context(),
		account,
		grokSessionMediaPostTypeImage,
		imageReferences[0],
		"",
	)
	if err != nil {
		r.persistSessionMediaRuntimeFeedback(c.Request.Context(), account, requestedModel, c.Request.URL.Path, err)
		r.writeMediaRuntimeError(c, err)
		return true
	}

	modeID := resolveGrokSessionMediaModeID(req.Model)
	payload, err := buildGrokSessionTextPayload(grokSessionTextRequest{
		ModelID: req.Model,
		ModeID:  modeID,
		Message: strings.TrimSpace(req.Prompt),
	})
	if err != nil {
		writeCompatibleGatewayMediaError(c, http.StatusBadRequest, "invalid_request_error", err.Error())
		return true
	}

	payload["modelName"] = grokSessionImageEditModelName
	payload["enableImageGeneration"] = true
	payload["enableImageStreaming"] = true
	payload["imageGenerationCount"] = req.N
	payload["returnImageBytes"] = false
	payload["returnRawGrokInXaiRequest"] = false
	payload["disableTextFollowUps"] = true
	payload["toolOverrides"] = map[string]any{"imageGen": true}
	payload["responseMetadata"] = map[string]any{
		"modelConfigOverride": map[string]any{
			"modelMap": map[string]any{
				"imageEditModel": grokSessionImageEditModelKind,
				"imageEditModelConfig": map[string]any{
					"imageReferences": grokSessionStringSliceToAny(imageReferences),
					"parentPostId":    postID,
				},
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		writeCompatibleGatewayMediaError(c, http.StatusInternalServerError, "api_error", "Failed to build Grok image edit request")
		return true
	}

	resp, target, err := r.doSessionJSONRequest(
		c.Request.Context(),
		account,
		grokSessionConversationEndpoint,
		payloadBytes,
		"application/json, text/event-stream, text/plain, */*",
		0,
	)
	if err != nil {
		r.persistSessionMediaRuntimeFeedback(c.Request.Context(), account, requestedModel, c.Request.URL.Path, err)
		r.writeMediaRuntimeError(c, err)
		return true
	}
	defer func() { _ = resp.Body.Close() }()

	imageURLs, err := collectGrokSessionImageURLs(resp, target, c.Request.URL.Path)
	if err != nil {
		r.persistSessionMediaRuntimeFeedback(c.Request.Context(), account, requestedModel, c.Request.URL.Path, err)
		r.writeMediaRuntimeError(c, err)
		return true
	}
	if len(imageURLs) == 0 {
		r.persistSessionMediaRuntimeFeedback(c.Request.Context(), account, requestedModel, c.Request.URL.Path, errors.New("grok session image edit returned no images"))
		writeCompatibleGatewayMediaError(c, http.StatusBadGateway, "api_error", "Grok session image edit returned no images")
		return true
	}

	responseBody, err := marshalGrokSessionImageResponse(imageURLs, r.now())
	if err != nil {
		writeCompatibleGatewayMediaError(c, http.StatusInternalServerError, "api_error", "Failed to build Grok image response")
		return true
	}
	if r.mediaAssets != nil {
		responseBody, _, _ = r.mediaAssets.RewriteResponse(c, account, responseBody, "image", req.ResponseFormat, requestedModel, canonicalModel, "")
	}
	r.persistSessionMediaRuntimeFeedback(c.Request.Context(), account, requestedModel, c.Request.URL.Path, nil)
	c.Data(http.StatusOK, "application/json", responseBody)
	return true
}

func (r *GrokSessionMediaRuntime) runVideoJob(account *Account, record *GrokVideoJobRecord, req grokSessionVideoCreateRequest) {
	if r == nil || r.videoJobs == nil || record == nil || account == nil {
		return
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			r.failVideoJob(context.Background(), record.JobID, "api_error", fmt.Sprintf("panic: %v", recovered))
		}
	}()

	ctx := context.Background()
	timeout := DefaultGrokRuntimeSettings().VideoTimeout()
	if r.gatewayService != nil && r.gatewayService.settingService != nil {
		timeout = r.gatewayService.settingService.GetGrokRuntimeSettings(ctx).VideoTimeout()
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	pollAfter := r.now().Add(grokSessionVideoPollInterval)
	_ = r.videoJobs.UpdateStatus(ctx, GrokVideoJobStatusPatch{
		JobID:            record.JobID,
		UpstreamStatus:   "running",
		NormalizedStatus: "in_progress",
		PollAfter:        &pollAfter,
	})

	artifact, err := r.generateSessionVideo(ctx, account, req)
	if err != nil {
		r.persistSessionMediaRuntimeFeedback(ctx, account, firstNonEmpty(record.CanonicalModel, record.RequestedModel, req.Model), "/v1/videos", err)
		r.failVideoJob(ctx, record.JobID, grokSessionMediaFeedbackCode(err), sanitizeUpstreamErrorMessage(err.Error()))
		return
	}

	outputAssetID := strings.TrimSpace(artifact.AssetID)
	if r.mediaAssets != nil {
		assetRecord, assetErr := r.mediaAssets.UpsertRemoteAssetRecord(
			ctx,
			account,
			"video",
			record.RequestedModel,
			record.CanonicalModel,
			record.JobID,
			artifact.AssetID,
			artifact.VideoURL,
		)
		if assetErr != nil {
			r.failVideoJob(ctx, record.JobID, "api_error", "Failed to persist Grok video asset")
			return
		}
		if assetRecord != nil {
			outputAssetID = strings.TrimSpace(assetRecord.AssetID)
		}
	}

	_ = r.videoJobs.UpdateStatus(ctx, GrokVideoJobStatusPatch{
		JobID:            record.JobID,
		UpstreamStatus:   "completed",
		NormalizedStatus: "completed",
		PollAfter:        nil,
		ErrorCode:        "",
		ErrorMessage:     "",
		OutputAssetID:    outputAssetID,
	})
	r.persistSessionMediaRuntimeFeedback(ctx, account, firstNonEmpty(record.CanonicalModel, record.RequestedModel, req.Model), "/v1/videos", nil)
}

func (r *GrokSessionMediaRuntime) generateSessionVideo(
	ctx context.Context,
	account *Account,
	req grokSessionVideoCreateRequest,
) (*grokSessionVideoArtifact, error) {
	var (
		reference    *grokSessionVideoReference
		parentPostID string
		err          error
	)

	if req.InputReference != nil {
		reference, err = r.prepareSessionVideoReference(ctx, account, *req.InputReference)
		if err != nil {
			return nil, err
		}
		parentPostID = reference.PostID
	} else {
		parentPostID, err = r.createSessionMediaPost(
			ctx,
			account,
			grokSessionMediaPostTypeVideo,
			"",
			req.Prompt,
		)
		if err != nil {
			return nil, err
		}
	}

	segments := buildGrokSessionVideoSegments(req.Seconds)
	totalSegments := len(segments)
	extendPostID := parentPostID
	elapsedSeconds := 0
	var artifact *grokSessionVideoArtifact

	for index, segmentLength := range segments {
		var (
			payload map[string]any
			referer string
		)
		if index == 0 {
			payload = buildGrokSessionVideoCreatePayload(
				req.Prompt,
				parentPostID,
				grokSessionVideoSizeMap[req.Size][0],
				req.ResolutionName,
				segmentLength,
				req.Preset,
				reference,
			)
			referer = buildGrokSessionReferer(ctx, r.gatewayService, "/imagine")
		} else {
			payload = buildGrokSessionVideoExtendPayload(
				req.Prompt,
				parentPostID,
				extendPostID,
				grokSessionVideoSizeMap[req.Size][0],
				req.ResolutionName,
				segmentLength,
				req.Preset,
				grokSessionVideoExtendStartTime(elapsedSeconds),
			)
			referer = buildGrokSessionReferer(ctx, r.gatewayService, "/imagine/post/"+parentPostID)
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		resp, target, err := r.doSessionJSONRequest(
			ctx,
			account,
			grokSessionConversationEndpoint,
			payloadBytes,
			"application/json, text/event-stream, text/plain, */*",
			0,
			withGrokSessionHeaderOverride("Referer", referer),
		)
		if err != nil {
			return nil, err
		}

		artifact, err = collectGrokSessionVideoArtifact(resp, target, totalSegments, index)
		_ = resp.Body.Close()
		if err != nil {
			return nil, err
		}
		extendPostID = firstNonEmpty(artifact.VideoPostID, artifact.AssetID, parentPostID)
		elapsedSeconds += segmentLength
	}

	if artifact == nil {
		return nil, errors.New("grok session video generation returned no artifact")
	}
	return artifact, nil
}

func (r *GrokSessionMediaRuntime) prepareSessionVideoReference(
	ctx context.Context,
	account *Account,
	input grokSessionInputReference,
) (*grokSessionVideoReference, error) {
	fileID := strings.TrimSpace(input.FileID)
	imageURL := strings.TrimSpace(input.ImageURL)
	switch {
	case fileID != "" && imageURL != "":
		return nil, errors.New("input_reference accepts only one of file_id or image_url")
	case fileID != "":
		return nil, errGrokSessionVideoReferenceFileIDUnsupported
	case imageURL == "":
		return nil, errors.New("input_reference.image_url is required")
	}

	contentURL := imageURL
	if !isGrokSessionAssetContentURL(imageURL) {
		uploaded, err := r.uploadSessionMediaInput(ctx, account, grokSessionUploadInput{Source: imageURL})
		if err != nil {
			return nil, err
		}
		contentURL = uploaded.ContentURL
	}

	postID, err := r.createSessionMediaPost(ctx, account, grokSessionMediaPostTypeImage, contentURL, "")
	if err != nil {
		return nil, err
	}
	return &grokSessionVideoReference{
		ContentURL: contentURL,
		PostID:     postID,
	}, nil
}

func (r *GrokSessionMediaRuntime) failVideoJob(ctx context.Context, jobID string, code string, message string) {
	if r == nil || r.videoJobs == nil || strings.TrimSpace(jobID) == "" {
		return
	}
	_ = r.videoJobs.UpdateStatus(ctx, GrokVideoJobStatusPatch{
		JobID:            jobID,
		UpstreamStatus:   "failed",
		NormalizedStatus: "failed",
		PollAfter:        nil,
		ErrorCode:        firstNonEmpty(strings.TrimSpace(code), "api_error"),
		ErrorMessage:     firstNonEmpty(strings.TrimSpace(message), "Grok session video generation failed"),
	})
}

type grokSessionRequestOption func(*http.Request)

func withGrokSessionHeaderOverride(key string, value string) grokSessionRequestOption {
	return func(req *http.Request) {
		if req == nil {
			return
		}
		if strings.TrimSpace(value) == "" {
			req.Header.Del(key)
			return
		}
		req.Header.Set(key, value)
	}
}

func (r *GrokSessionMediaRuntime) doSessionJSONRequest(
	ctx context.Context,
	account *Account,
	requestPath string,
	payload []byte,
	accept string,
	timeout time.Duration,
	opts ...grokSessionRequestOption,
) (*http.Response, grokTransportTarget, error) {
	target, err := r.resolveSessionTarget(ctx, account, requestPath)
	if err != nil {
		return nil, grokTransportTarget{}, err
	}

	reqCtx := ctx
	if reqCtx == nil {
		reqCtx = context.Background()
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(reqCtx, timeout)
		defer cancel()
	}

	req, err := newGrokSessionJSONRequest(reqCtx, http.MethodPost, target, payload, accept)
	if err != nil {
		return nil, grokTransportTarget{}, err
	}
	for _, opt := range opts {
		if opt != nil {
			opt(req)
		}
	}

	proxyURL := resolveGrokMediaAssetProxyURL(account)
	resp, err := r.gatewayService.httpUpstream.DoWithTLS(
		req,
		proxyURL,
		account.ID,
		resolveAccountConcurrency(account),
		resolveGrokGatewayTLSProfile(r.gatewayService, account),
	)
	if err != nil {
		return nil, grokTransportTarget{}, err
	}
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return resp, target, nil
	}

	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	upstreamMessage := sanitizeUpstreamErrorMessage(strings.TrimSpace(extractUpstreamErrorMessage(respBody)))
	if upstreamMessage == "" {
		upstreamMessage = strings.TrimSpace(string(respBody))
	}
	if upstreamMessage == "" {
		upstreamMessage = http.StatusText(resp.StatusCode)
	}
	return nil, grokTransportTarget{}, &grokSessionMediaUpstreamError{
		statusCode: mapUpstreamStatusCode(resp.StatusCode),
		code:       grokResponsesErrorCodeForStatus(resp.StatusCode),
		message:    upstreamMessage,
	}
}

func (r *GrokSessionMediaRuntime) doRawSessionRequest(
	ctx context.Context,
	account *Account,
	method string,
	rawURL string,
	body io.Reader,
	contentType string,
	accept string,
	applySessionHeaders bool,
) (*http.Response, error) {
	if r == nil || r.gatewayService == nil || r.gatewayService.httpUpstream == nil {
		return nil, errors.New("grok session media runtime is not configured")
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(contentType) != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if strings.TrimSpace(accept) != "" {
		req.Header.Set("Accept", accept)
	}

	if applySessionHeaders {
		target, targetErr := r.resolveSessionTarget(ctx, account, grokSessionConversationEndpoint)
		if targetErr != nil {
			return nil, targetErr
		}
		target.URL = rawURL
		applyGrokSessionBrowserHeaders(req.Header, target, accept)
		target.Apply(req)
	}

	resp, err := r.gatewayService.httpUpstream.DoWithTLS(
		req,
		resolveGrokMediaAssetProxyURL(account),
		account.ID,
		resolveAccountConcurrency(account),
		resolveGrokGatewayTLSProfile(r.gatewayService, account),
	)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return resp, nil
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	upstreamMessage := sanitizeUpstreamErrorMessage(strings.TrimSpace(extractUpstreamErrorMessage(respBody)))
	if upstreamMessage == "" {
		upstreamMessage = strings.TrimSpace(string(respBody))
	}
	if upstreamMessage == "" {
		upstreamMessage = http.StatusText(resp.StatusCode)
	}
	return nil, &grokSessionMediaUpstreamError{
		statusCode: mapUpstreamStatusCode(resp.StatusCode),
		code:       grokResponsesErrorCodeForStatus(resp.StatusCode),
		message:    upstreamMessage,
	}
}

func (r *GrokSessionMediaRuntime) resolveSessionTarget(ctx context.Context, account *Account, requestPath string) (grokTransportTarget, error) {
	if r == nil || r.gatewayService == nil {
		return grokTransportTarget{}, errors.New("grok session media runtime is not configured")
	}
	if account == nil || account.Type != AccountTypeSession {
		return grokTransportTarget{}, errors.New("account is not a grok session account")
	}

	runtimeSettings := DefaultGrokRuntimeSettings()
	if r.gatewayService.settingService != nil {
		runtimeSettings = r.gatewayService.settingService.GetGrokRuntimeSettings(ctx)
	}
	cookieHeader, err := BuildGrokSessionCookieHeader(
		account.GetGrokSessionToken(),
		account.GetGrokSessionCFCookies(),
		account.GetGrokSessionCFClearance(),
	)
	if err != nil {
		return grokTransportTarget{}, err
	}
	targetURL, sessionBaseURL, err := buildGrokSessionTransportURLForPathWithSettings(
		r.gatewayService.validateUpstreamBaseURL,
		runtimeSettings,
		requestPath,
	)
	if err != nil {
		return grokTransportTarget{}, err
	}
	return grokTransportTarget{
		Kind:           grokTransportKindSession,
		URL:            targetURL,
		SessionBaseURL: sessionBaseURL,
		CookieHeader:   cookieHeader,
		UserAgent:      account.GetGrokSessionUserAgent(),
		AcceptLang:     account.GetGrokSessionAcceptLanguage(),
	}, nil
}

func (r *GrokSessionMediaRuntime) uploadSessionMediaInput(
	ctx context.Context,
	account *Account,
	input grokSessionUploadInput,
) (*grokSessionUploadedAsset, error) {
	normalized, err := r.normalizeSessionUploadInput(ctx, account, input)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(map[string]any{
		"fileName":     normalized.FileName,
		"fileMimeType": normalized.MimeType,
		"content":      normalized.Base64,
	})
	if err != nil {
		return nil, err
	}

	resp, target, err := r.doSessionJSONRequest(
		ctx,
		account,
		grokSessionUploadFileEndpoint,
		payload,
		"application/json, text/plain, */*",
		0,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}

	fileID := firstNonEmpty(
		strings.TrimSpace(gjson.GetBytes(respBody, "fileMetadataId").String()),
		strings.TrimSpace(gjson.GetBytes(respBody, "fileId").String()),
	)
	fileURI := strings.TrimSpace(gjson.GetBytes(respBody, "fileUri").String())
	contentURL := resolveGrokSessionUploadedAssetURL(fileID, fileURI, grokSessionUserIDFromTargetAndResponse(target, resp))
	if contentURL == "" {
		return nil, errors.New("grok session upload returned no resolvable asset reference")
	}
	return &grokSessionUploadedAsset{
		FileID:     fileID,
		FileURI:    fileURI,
		ContentURL: contentURL,
	}, nil
}

func (r *GrokSessionMediaRuntime) uploadSessionMediaInputs(
	ctx context.Context,
	account *Account,
	inputs []grokSessionUploadInput,
) ([]*grokSessionUploadedAsset, error) {
	uploaded := make([]*grokSessionUploadedAsset, 0, len(inputs))
	for _, input := range inputs {
		asset, err := r.uploadSessionMediaInput(ctx, account, input)
		if err != nil {
			return nil, err
		}
		if asset != nil {
			uploaded = append(uploaded, asset)
		}
	}
	if len(uploaded) == 0 {
		return nil, errors.New("image is required")
	}
	return uploaded, nil
}

func (r *GrokSessionMediaRuntime) normalizeSessionUploadInput(
	ctx context.Context,
	account *Account,
	input grokSessionUploadInput,
) (*grokSessionUploadInput, error) {
	source := strings.TrimSpace(input.Source)
	if strings.TrimSpace(input.Base64) != "" {
		fileName := firstNonEmpty(strings.TrimSpace(input.FileName), "upload.bin")
		mimeType := firstNonEmpty(strings.TrimSpace(input.MimeType), mime.TypeByExtension(path.Ext(fileName)), "application/octet-stream")
		return &grokSessionUploadInput{
			FileName: fileName,
			MimeType: mimeType,
			Base64:   strings.TrimSpace(input.Base64),
		}, nil
	}

	if source == "" {
		return nil, errors.New("image is required")
	}
	if parsed, ok := parseGrokSessionDataURI(source); ok {
		return &grokSessionUploadInput{
			FileName: parsed.FileName,
			MimeType: parsed.MimeType,
			Base64:   parsed.Base64,
		}, nil
	}

	parsedURL, err := url.Parse(source)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, errors.New("image must be a URL, data URI, or uploaded file")
	}
	resp, err := r.doRawSessionRequest(
		ctx,
		account,
		http.MethodGet,
		source,
		nil,
		"",
		"*/*",
		isGrokSessionRequestHost(parsedURL.Host),
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 25<<20))
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, errors.New("image download returned empty content")
	}

	mimeType := strings.TrimSpace(strings.Split(resp.Header.Get("Content-Type"), ";")[0])
	if mimeType == "" {
		mimeType = http.DetectContentType(raw)
	}
	fileName := path.Base(parsedURL.Path)
	if fileName == "" || fileName == "." || fileName == "/" {
		fileName = grokSessionDefaultFileName(mimeType)
	}
	return &grokSessionUploadInput{
		FileName: fileName,
		MimeType: mimeType,
		Base64:   base64.StdEncoding.EncodeToString(raw),
	}, nil
}

func (r *GrokSessionMediaRuntime) createSessionMediaPost(
	ctx context.Context,
	account *Account,
	mediaType string,
	mediaURL string,
	prompt string,
) (string, error) {
	payload := map[string]any{
		"mediaType": mediaType,
	}
	if strings.TrimSpace(mediaURL) != "" {
		payload["mediaUrl"] = strings.TrimSpace(mediaURL)
	}
	if strings.TrimSpace(prompt) != "" {
		payload["prompt"] = strings.TrimSpace(prompt)
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, _, err := r.doSessionJSONRequest(
		ctx,
		account,
		grokSessionMediaPostCreateEndpoint,
		payloadBytes,
		"application/json, text/plain, */*",
		0,
		withGrokSessionHeaderOverride("Referer", buildGrokSessionReferer(ctx, r.gatewayService, "/imagine")),
	)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", err
	}
	postID := firstNonEmpty(
		strings.TrimSpace(gjson.GetBytes(respBody, "post.id").String()),
		strings.TrimSpace(gjson.GetBytes(respBody, "id").String()),
		strings.TrimSpace(gjson.GetBytes(respBody, "postId").String()),
	)
	if postID == "" {
		return "", errors.New("grok session media post returned no post id")
	}
	return postID, nil
}

func collectGrokSessionImageURLs(resp *http.Response, target grokTransportTarget, requestPath string) ([]string, error) {
	if resp == nil || resp.Body == nil {
		return nil, errors.New("grok session response is empty")
	}

	userID := grokSessionUserIDFromTargetAndResponse(target, resp)
	scanner := bufio.NewScanner(resp.Body)
	scanBuf := getSSEScannerBuf64K()
	scanner.Buffer(scanBuf[:0], defaultMaxLineSize)
	defer putSSEScannerBuf64K(scanBuf)

	imageURLs := make([]string, 0, 4)
	seen := make(map[string]struct{})

	appendURL := func(raw string) {
		value := strings.TrimSpace(raw)
		if value == "" {
			return
		}
		if _, exists := seen[value]; exists {
			return
		}
		seen[value] = struct{}{}
		imageURLs = append(imageURLs, value)
	}

	for scanner.Scan() {
		payload, ok := normalizeGrokSessionResponseLine(scanner.Text())
		if !ok {
			continue
		}
		raw := []byte(payload)
		if message := sanitizeUpstreamErrorMessage(strings.TrimSpace(extractUpstreamErrorMessage(raw))); message != "" {
			return nil, &grokSessionMediaUpstreamError{
				statusCode: http.StatusBadGateway,
				code:       "api_error",
				message:    message,
			}
		}

		appendURL(strings.TrimSpace(gjson.GetBytes(raw, "result.response.streamingImageGenerationResponse.imageUrl").String()))

		if assetID := strings.TrimSpace(gjson.GetBytes(raw, "result.response.streamingImageGenerationResponse.assetId").String()); assetID != "" && userID != "" {
			appendURL(resolveGrokSessionUploadedAssetURL(assetID, "", userID))
		}
		for _, imageURL := range extractGrokSessionImageCardURLs(raw) {
			appendURL(imageURL)
		}

		for _, value := range gjson.GetBytes(raw, "result.response.modelResponse.generatedImageUrls").Array() {
			appendURL(value.String())
		}
		for _, value := range gjson.GetBytes(raw, "result.response.modelResponse.fileAttachments").Array() {
			assetID := strings.TrimSpace(value.String())
			if assetID == "" || userID == "" {
				continue
			}
			appendURL(resolveGrokSessionUploadedAssetURL(assetID, "", userID))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(imageURLs) == 0 && strings.Contains(strings.ToLower(strings.TrimSpace(requestPath)), "/images/") {
		return nil, errors.New("grok session media stream ended without an image result")
	}
	return imageURLs, nil
}

func extractGrokSessionImageCardURLs(raw []byte) []string {
	if len(raw) == 0 {
		return nil
	}

	payloads := make([][]byte, 0, 3)
	if cardJSON := strings.TrimSpace(gjson.GetBytes(raw, "result.response.cardAttachment.jsonData").String()); cardJSON != "" {
		payloads = append(payloads, []byte(cardJSON))
	}
	for _, item := range gjson.GetBytes(raw, "result.response.modelResponse.cardAttachmentsJson").Array() {
		cardJSON := strings.TrimSpace(item.String())
		if cardJSON == "" {
			continue
		}
		payloads = append(payloads, []byte(cardJSON))
	}
	if len(payloads) == 0 {
		return nil
	}

	urls := make([]string, 0, len(payloads))
	seen := make(map[string]struct{}, len(payloads))
	for _, payload := range payloads {
		if !gjson.ValidBytes(payload) {
			continue
		}
		imageURL := firstNonEmpty(
			strings.TrimSpace(gjson.GetBytes(payload, "image_chunk.imageUrl").String()),
			strings.TrimSpace(gjson.GetBytes(payload, "image_chunk.image_url").String()),
		)
		if imageURL == "" {
			continue
		}
		progress := gjson.GetBytes(payload, "image_chunk.progress")
		if progress.Exists() && progress.Type == gjson.Number && progress.Int() < 100 {
			continue
		}
		if gjson.GetBytes(payload, "image_chunk.moderated").Bool() {
			continue
		}

		resolved := absolutizeGrokSessionAssetURL(imageURL)
		if resolved == "" {
			continue
		}
		if _, ok := seen[resolved]; ok {
			continue
		}
		seen[resolved] = struct{}{}
		urls = append(urls, resolved)
	}
	return urls
}

func collectGrokSessionVideoArtifact(
	resp *http.Response,
	target grokTransportTarget,
	totalSegments int,
	index int,
) (*grokSessionVideoArtifact, error) {
	if resp == nil || resp.Body == nil {
		return nil, errors.New("grok session response is empty")
	}

	userID := grokSessionUserIDFromTargetAndResponse(target, resp)
	scanner := bufio.NewScanner(resp.Body)
	scanBuf := getSSEScannerBuf64K()
	scanner.Buffer(scanBuf[:0], defaultMaxLineSize)
	defer putSSEScannerBuf64K(scanBuf)

	artifact := &grokSessionVideoArtifact{}
	for scanner.Scan() {
		payload, ok := normalizeGrokSessionResponseLine(scanner.Text())
		if !ok {
			continue
		}
		raw := []byte(payload)
		if message := sanitizeUpstreamErrorMessage(strings.TrimSpace(extractUpstreamErrorMessage(raw))); message != "" {
			return nil, &grokSessionMediaUpstreamError{
				statusCode: http.StatusBadGateway,
				code:       "api_error",
				message:    message,
			}
		}

		stream := gjson.GetBytes(raw, "result.response.streamingVideoGenerationResponse")
		if stream.Exists() {
			artifact.VideoPostID = firstNonEmpty(
				strings.TrimSpace(stream.Get("videoPostId").String()),
				strings.TrimSpace(stream.Get("videoId").String()),
				artifact.VideoPostID,
			)
			if progressValue := stream.Get("progress"); progressValue.Exists() && totalSegments > 0 {
				_ = index // reserved for future progress persistence
				_ = progressValue.Int()
			}
			if rawURL := strings.TrimSpace(stream.Get("videoUrl").String()); rawURL != "" {
				artifact.VideoURL = absolutizeGrokSessionAssetURL(rawURL)
			}
			if assetID := strings.TrimSpace(stream.Get("assetId").String()); assetID != "" {
				artifact.AssetID = assetID
				if artifact.VideoURL == "" && userID != "" {
					artifact.VideoURL = resolveGrokSessionUploadedAssetURL(assetID, "", userID)
				}
			}
			if thumbnail := strings.TrimSpace(stream.Get("thumbnailImageUrl").String()); thumbnail != "" {
				artifact.ThumbnailURL = absolutizeGrokSessionAssetURL(thumbnail)
			}
		}

		for _, attachment := range gjson.GetBytes(raw, "result.response.modelResponse.fileAttachments").Array() {
			assetID := strings.TrimSpace(attachment.String())
			if assetID == "" {
				continue
			}
			if artifact.AssetID == "" {
				artifact.AssetID = assetID
			}
			if artifact.VideoURL == "" && userID != "" {
				artifact.VideoURL = resolveGrokSessionUploadedAssetURL(assetID, "", userID)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if artifact.VideoURL == "" {
		return nil, errors.New("grok session video generation returned no final video URL")
	}
	if artifact.VideoPostID == "" {
		artifact.VideoPostID = firstNonEmpty(artifact.AssetID, uuid.NewString())
	}
	return artifact, nil
}

func parseGrokSessionImageGenerationRequest(body []byte, defaultModel string) (grokSessionImageGenerationRequest, error) {
	var req grokSessionImageGenerationRequest
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return req, errors.New("failed to parse image generation request body")
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return req, errors.New("failed to parse image generation request body")
	}
	responseFormat, err := normalizeGrokOpenAIImageResponseFormat(req.ResponseFormat)
	if err != nil {
		return req, err
	}
	req.Model = firstNonEmpty(strings.TrimSpace(req.Model), strings.TrimSpace(defaultModel))
	req.Prompt = strings.TrimSpace(req.Prompt)
	req.ResponseFormat = responseFormat
	if req.Model == "" {
		return req, errors.New("model is required")
	}
	if req.Prompt == "" {
		return req, errors.New("prompt is required")
	}
	if req.N <= 0 {
		req.N = 1
	}
	return req, nil
}

func parseGrokSessionImageEditRequest(c *gin.Context, body []byte, defaultModel string) (grokSessionImageEditRequest, error) {
	contentType := ""
	if c != nil && c.Request != nil {
		contentType = c.Request.Header.Get("Content-Type")
	}
	mediaType, params, _ := mime.ParseMediaType(contentType)
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))

	if mediaType == "multipart/form-data" {
		return parseMultipartGrokSessionImageEditRequest(body, params["boundary"], defaultModel)
	}

	req := grokSessionImageEditRequest{
		Model:  firstNonEmpty(strings.TrimSpace(gjson.GetBytes(body, "model").String()), strings.TrimSpace(defaultModel)),
		Prompt: strings.TrimSpace(gjson.GetBytes(body, "prompt").String()),
		N:      int(gjson.GetBytes(body, "n").Int()),
	}
	responseFormat, err := normalizeGrokOpenAIImageResponseFormat(gjson.GetBytes(body, "response_format").String())
	if err != nil {
		return req, err
	}
	req.ResponseFormat = responseFormat
	if req.N <= 0 {
		req.N = grokSessionDefaultImageEditCount
	}
	if req.N > 2 {
		req.N = 2
	}
	if req.Model == "" {
		return req, errors.New("model is required")
	}
	if req.Prompt == "" {
		return req, errors.New("prompt is required")
	}

	if mask := strings.TrimSpace(gjson.GetBytes(body, "mask").String()); mask != "" {
		req.HasMask = true
		req.MaskProvided = mask
	}

	req.InputImages = append(req.InputImages, grokSessionImageEditSourcesFromJSON(body, "image")...)
	req.InputImages = append(req.InputImages, grokSessionImageEditSourcesFromJSON(body, "image_url")...)
	req.InputImages = append(req.InputImages, grokSessionImageEditSourcesFromJSON(body, "input_image")...)
	if len(req.InputImages) == 0 {
		return req, errors.New("image is required")
	}
	return req, nil
}

func parseMultipartGrokSessionImageEditRequest(body []byte, boundary string, defaultModel string) (grokSessionImageEditRequest, error) {
	req := grokSessionImageEditRequest{
		Model: firstNonEmpty(strings.TrimSpace(defaultModel)),
		N:     grokSessionDefaultImageEditCount,
	}
	if strings.TrimSpace(boundary) == "" {
		return req, errors.New("multipart image edit request boundary is missing")
	}

	reader := multipart.NewReader(bytes.NewReader(body), boundary)
	for {
		part, err := reader.NextPart()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return req, errors.New("failed to parse multipart image edit request")
		}

		formName := strings.TrimSpace(part.FormName())
		fileName := strings.TrimSpace(part.FileName())
		switch formName {
		case "model":
			value, readErr := readMultipartTextField(part, 8<<10)
			if readErr == nil {
				req.Model = firstNonEmpty(strings.TrimSpace(value), req.Model)
			}
		case "prompt":
			value, readErr := readMultipartTextField(part, 128<<10)
			if readErr == nil {
				req.Prompt = strings.TrimSpace(value)
			}
		case "n":
			value, readErr := readMultipartTextField(part, 64)
			if readErr == nil {
				if parsed, convErr := strconv.Atoi(strings.TrimSpace(value)); convErr == nil && parsed > 0 {
					req.N = parsed
				}
			}
		case "response_format":
			value, readErr := readMultipartTextField(part, 64)
			if readErr == nil {
				req.ResponseFormat = strings.TrimSpace(value)
			}
		case "mask":
			value, readErr := readMultipartTextField(part, 128<<10)
			if readErr == nil && strings.TrimSpace(value) != "" {
				req.HasMask = true
				req.MaskProvided = strings.TrimSpace(value)
			}
		case "image", "image[]":
			raw, readErr := io.ReadAll(io.LimitReader(part, 25<<20))
			if readErr == nil && len(raw) > 0 {
				mimeType := strings.TrimSpace(strings.Split(part.Header.Get("Content-Type"), ";")[0])
				if mimeType == "" {
					mimeType = http.DetectContentType(raw)
				}
				name := firstNonEmpty(fileName, grokSessionDefaultFileName(mimeType))
				req.InputImages = append(req.InputImages, grokSessionUploadInput{
					FileName: name,
					MimeType: mimeType,
					Base64:   base64.StdEncoding.EncodeToString(raw),
				})
			}
		}
		_ = part.Close()
	}

	req.Model = firstNonEmpty(strings.TrimSpace(req.Model), strings.TrimSpace(defaultModel))
	req.Prompt = strings.TrimSpace(req.Prompt)
	responseFormat, err := normalizeGrokOpenAIImageResponseFormat(req.ResponseFormat)
	if err != nil {
		return req, err
	}
	req.ResponseFormat = responseFormat
	if req.N <= 0 {
		req.N = grokSessionDefaultImageEditCount
	}
	if req.N > 2 {
		req.N = 2
	}
	if req.Model == "" {
		return req, errors.New("model is required")
	}
	if req.Prompt == "" {
		return req, errors.New("prompt is required")
	}
	if len(req.InputImages) == 0 {
		return req, errors.New("image is required")
	}
	return req, nil
}

func grokSessionImageEditSourcesFromJSON(body []byte, field string) []grokSessionUploadInput {
	field = strings.TrimSpace(field)
	if field == "" {
		return nil
	}

	result := make([]grokSessionUploadInput, 0, 4)
	appendSource := func(raw string) {
		if trimmed := strings.TrimSpace(raw); trimmed != "" {
			result = append(result, grokSessionUploadInput{Source: trimmed})
		}
	}

	value := gjson.GetBytes(body, field)
	if !value.Exists() {
		return result
	}
	if value.IsArray() {
		for _, item := range value.Array() {
			appendSource(item.String())
		}
		return result
	}
	appendSource(value.String())
	return result
}

func parseGrokSessionVideoCreateRequest(body []byte, defaultModel string) (grokSessionVideoCreateRequest, error) {
	req := grokSessionVideoCreateRequest{
		Model:          strings.TrimSpace(defaultModel),
		Seconds:        grokSessionDefaultVideoSeconds,
		Size:           grokSessionDefaultVideoSize,
		Quality:        grokSessionDefaultVideoQuality,
		ResolutionName: grokSessionDefaultVideoResolutionName,
		Preset:         grokSessionDefaultVideoPreset,
	}
	if len(body) == 0 {
		return req, nil
	}
	if !gjson.ValidBytes(body) {
		return req, errors.New("failed to parse video request body")
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return req, errors.New("failed to parse video request body")
	}

	req.Model = firstNonEmpty(strings.TrimSpace(req.Model), strings.TrimSpace(defaultModel))
	req.Prompt = strings.TrimSpace(req.Prompt)
	req.Quality = firstNonEmpty(strings.TrimSpace(req.Quality), grokSessionDefaultVideoQuality)
	if req.Model == "" {
		return req, errors.New("model is required")
	}
	if req.Prompt == "" {
		return req, errors.New("prompt is required")
	}
	if req.Seconds == 0 {
		req.Seconds = grokSessionDefaultVideoSeconds
	}
	if !isSupportedGrokSessionVideoLength(req.Seconds) {
		return req, errors.New("seconds must be one of [6, 10, 12, 16, 20]")
	}
	req.Size = firstNonEmpty(strings.TrimSpace(req.Size), grokSessionDefaultVideoSize)
	sizeInfo, ok := grokSessionVideoSizeMap[req.Size]
	if !ok {
		return req, fmt.Errorf("size must be one of [%s]", strings.Join(sortedGrokSessionVideoSizes(), ", "))
	}
	req.ResolutionName = firstNonEmpty(strings.TrimSpace(req.ResolutionName), sizeInfo[1], grokSessionDefaultVideoResolutionName)
	switch req.ResolutionName {
	case "480p", "720p":
	default:
		return req, errors.New("resolution_name must be one of [480p, 720p]")
	}
	req.Preset = firstNonEmpty(strings.ToLower(strings.TrimSpace(req.Preset)), grokSessionDefaultVideoPreset)
	if _, ok := grokSessionVideoPresetFlags[req.Preset]; !ok {
		return req, errors.New("preset must be one of [custom, fun, normal, spicy]")
	}
	if req.InputReference != nil {
		req.InputReference.FileID = strings.TrimSpace(req.InputReference.FileID)
		req.InputReference.ImageURL = strings.TrimSpace(req.InputReference.ImageURL)
	}
	return req, nil
}

func buildGrokSessionVideoJobResponse(record GrokVideoJobRecord, req grokSessionVideoCreateRequest) gin.H {
	status := firstNonEmpty(strings.TrimSpace(record.NormalizedStatus), normalizeGrokVideoStatus(record.UpstreamStatus), "queued")
	progress := 0
	switch status {
	case "in_progress":
		progress = 50
	case "completed":
		progress = 100
	}

	response := gin.H{
		"id":         strings.TrimSpace(record.JobID),
		"object":     "video",
		"created_at": record.CreatedAt.Unix(),
		"status":     status,
		"model":      firstNonEmpty(strings.TrimSpace(record.RequestedModel), strings.TrimSpace(req.Model)),
		"progress":   progress,
		"prompt":     strings.TrimSpace(req.Prompt),
		"seconds":    strconv.Itoa(maxInt(req.Seconds, grokSessionDefaultVideoSeconds)),
		"size":       firstNonEmpty(strings.TrimSpace(req.Size), grokSessionDefaultVideoSize),
		"quality":    firstNonEmpty(strings.TrimSpace(req.Quality), grokSessionDefaultVideoQuality),
	}
	if record.PollAfter != nil {
		secondsUntil := int(time.Until(record.PollAfter.UTC()).Seconds())
		if secondsUntil < 1 {
			secondsUntil = 1
		}
		if status != "completed" && status != "failed" {
			response["poll_after"] = secondsUntil
		}
	}
	if status == "completed" || status == "failed" {
		response["completed_at"] = record.UpdatedAt.Unix()
	}
	if strings.TrimSpace(record.ErrorCode) != "" || strings.TrimSpace(record.ErrorMessage) != "" {
		response["error"] = gin.H{
			"code":    firstNonEmpty(strings.TrimSpace(record.ErrorCode), "api_error"),
			"message": firstNonEmpty(strings.TrimSpace(record.ErrorMessage), "Grok session video generation failed"),
		}
	}
	return response
}

func buildGrokSessionVideoCreatePayload(
	prompt string,
	parentPostID string,
	aspectRatio string,
	resolutionName string,
	videoLength int,
	preset string,
	reference *grokSessionVideoReference,
) map[string]any {
	message := buildGrokSessionVideoMessage(prompt, preset, reference)
	payload := map[string]any{
		"temporary":        true,
		"modelName":        grokSessionVideoModelName,
		"message":          message,
		"toolOverrides":    map[string]any{"videoGen": true},
		"enableSideBySide": true,
		"responseMetadata": map[string]any{
			"experiments": []any{},
			"modelConfigOverride": map[string]any{
				"modelMap": map[string]any{
					"videoGenModelConfig": map[string]any{
						"parentPostId":   parentPostID,
						"aspectRatio":    aspectRatio,
						"videoLength":    videoLength,
						"resolutionName": resolutionName,
					},
				},
			},
		},
	}
	if reference != nil && strings.TrimSpace(reference.PostID) != "" {
		payload["fileAttachments"] = []any{reference.PostID}
	}
	return payload
}

func buildGrokSessionVideoExtendPayload(
	prompt string,
	parentPostID string,
	extendPostID string,
	aspectRatio string,
	resolutionName string,
	videoLength int,
	preset string,
	startTimeS float64,
) map[string]any {
	return map[string]any{
		"temporary":        true,
		"modelName":        grokSessionVideoModelName,
		"message":          buildGrokSessionVideoMessage(prompt, preset, nil),
		"toolOverrides":    map[string]any{"videoGen": true},
		"enableSideBySide": true,
		"responseMetadata": map[string]any{
			"experiments": []any{},
			"modelConfigOverride": map[string]any{
				"modelMap": map[string]any{
					"videoGenModelConfig": map[string]any{
						"isVideoExtension":        true,
						"videoExtensionStartTime": startTimeS,
						"extendPostId":            extendPostID,
						"stitchWithExtendPostId":  true,
						"originalPrompt":          prompt,
						"originalPostId":          parentPostID,
						"originalRefType":         grokSessionVideoExtensionRefType,
						"mode":                    preset,
						"aspectRatio":             aspectRatio,
						"videoLength":             videoLength,
						"resolutionName":          resolutionName,
						"parentPostId":            parentPostID,
						"isVideoEdit":             false,
					},
				},
			},
		},
	}
}

func buildGrokSessionVideoMessage(prompt string, preset string, reference *grokSessionVideoReference) string {
	message := strings.TrimSpace(prompt)
	flag := grokSessionVideoPresetFlags[strings.ToLower(strings.TrimSpace(preset))]
	if flag == "" {
		flag = grokSessionVideoPresetFlags[grokSessionDefaultVideoPreset]
	}
	message = strings.TrimSpace(message + " " + flag)
	if reference != nil && strings.TrimSpace(reference.ContentURL) != "" {
		return strings.TrimSpace(reference.ContentURL) + "  " + message
	}
	return message
}

func buildGrokSessionVideoSegments(seconds int) []int {
	switch seconds {
	case 6:
		return []int{6}
	case 10:
		return []int{10}
	case 12:
		return []int{6, 6}
	case 16:
		return []int{10, 6}
	case 20:
		return []int{10, 10}
	default:
		return []int{grokSessionDefaultVideoSeconds}
	}
}

func grokSessionVideoExtendStartTime(seconds int) float64 {
	return float64(seconds) + (1.0 / 24.0)
}

func marshalGrokSessionImageResponse(imageURLs []string, now time.Time) ([]byte, error) {
	data := make([]map[string]any, 0, len(imageURLs))
	for _, imageURL := range imageURLs {
		data = append(data, map[string]any{
			"url": strings.TrimSpace(imageURL),
		})
	}
	return json.Marshal(map[string]any{
		"created": now.Unix(),
		"data":    data,
	})
}

func resolveGrokSessionMediaModeID(model string) string {
	normalized := strings.ToLower(strings.TrimSpace(model))
	switch normalized {
	case "grok-2-image", "grok-imagine-image-lite":
		return grokSessionModeFast
	default:
		return grokSessionModeAuto
	}
}

func grokSessionImageRouteIsEdit(requestPath string) bool {
	normalized := strings.ToLower(strings.TrimSpace(requestPath))
	return strings.Contains(normalized, "/images/edits")
}

type grokSessionParsedDataURI struct {
	FileName string
	MimeType string
	Base64   string
}

func parseGrokSessionDataURI(raw string) (*grokSessionParsedDataURI, bool) {
	if !strings.HasPrefix(strings.TrimSpace(raw), "data:") {
		return nil, false
	}
	parts := strings.SplitN(raw, ",", 2)
	if len(parts) != 2 {
		return nil, false
	}
	header := strings.TrimSpace(parts[0])
	if !strings.Contains(header, ";base64") {
		return nil, false
	}
	mimeType := strings.TrimSpace(strings.TrimPrefix(strings.SplitN(header, ";", 2)[0], "data:"))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	base64Payload := strings.TrimSpace(parts[1])
	if base64Payload == "" {
		return nil, false
	}
	return &grokSessionParsedDataURI{
		FileName: grokSessionDefaultFileName(mimeType),
		MimeType: mimeType,
		Base64:   base64Payload,
	}, true
}

func grokSessionDefaultFileName(mimeType string) string {
	extension := ".bin"
	if exts, _ := mime.ExtensionsByType(mimeType); len(exts) > 0 && strings.TrimSpace(exts[0]) != "" {
		extension = exts[0]
	}
	return "upload" + extension
}

func resolveGrokSessionUploadedAssetURL(fileID string, fileURI string, userID string) string {
	if parsed, err := url.Parse(strings.TrimSpace(fileURI)); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		return parsed.String()
	}
	if strings.TrimSpace(fileURI) != "" {
		return absolutizeGrokSessionAssetURL(fileURI)
	}
	fileID = strings.TrimSpace(fileID)
	userID = strings.TrimSpace(userID)
	if fileID != "" && userID != "" {
		return strings.TrimRight(grokSessionUploadedAssetBaseURL, "/") + "/users/" + userID + "/" + fileID + "/content"
	}
	return ""
}

func absolutizeGrokSessionAssetURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	parsed, err := url.Parse(trimmed)
	if err == nil && parsed.Scheme != "" && parsed.Host != "" {
		return parsed.String()
	}
	if strings.HasPrefix(trimmed, "/") {
		return strings.TrimRight(grokSessionUploadedAssetBaseURL, "/") + trimmed
	}
	return strings.TrimRight(grokSessionUploadedAssetBaseURL, "/") + "/" + trimmed
}

func collectGrokSessionUploadedContentURLs(assets []*grokSessionUploadedAsset) []string {
	if len(assets) == 0 {
		return nil
	}
	urls := make([]string, 0, len(assets))
	for _, asset := range assets {
		if asset == nil {
			continue
		}
		if trimmed := strings.TrimSpace(asset.ContentURL); trimmed != "" {
			urls = append(urls, trimmed)
		}
	}
	return urls
}

func grokSessionStringSliceToAny(values []string) []any {
	result := make([]any, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func grokSessionUserIDFromTargetAndResponse(target grokTransportTarget, resp *http.Response) string {
	if userID := grokSessionExtractCookieValue(target.CookieHeader, "x-userid"); userID != "" {
		return userID
	}
	if resp == nil {
		return ""
	}
	for _, setCookie := range resp.Header.Values("Set-Cookie") {
		if userID := grokSessionExtractCookieValue(setCookie, "x-userid"); userID != "" {
			return userID
		}
	}
	return ""
}

func grokSessionExtractCookieValue(raw string, name string) string {
	jar, err := parseOptionalGrokSessionCookieHeader(raw)
	if err == nil {
		if value := strings.TrimSpace(jar.Get(name)); value != "" {
			return value
		}
	}
	normalizedName := strings.ToLower(strings.TrimSpace(name)) + "="
	for _, part := range strings.Split(raw, ";") {
		part = strings.TrimSpace(part)
		if !strings.HasPrefix(strings.ToLower(part), normalizedName) {
			continue
		}
		return strings.TrimSpace(part[len(normalizedName):])
	}
	return ""
}

func buildGrokSessionReferer(ctx context.Context, gatewayService *GatewayService, requestPath string) string {
	settings := DefaultGrokRuntimeSettings()
	if gatewayService != nil && gatewayService.settingService != nil {
		settings = gatewayService.settingService.GetGrokRuntimeSettings(ctx)
	}
	baseURL := normalizeGrokRuntimeBaseURL(settings.SessionBaseURL, grokWebBaseURL)
	return strings.TrimRight(baseURL, "/") + normalizeGrokSessionTransportPath(requestPath)
}

func isGrokSessionRequestHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	return host == "grok.com" || host == "assets.grok.com" || strings.HasSuffix(host, ".grok.com")
}

func isGrokSessionAssetContentURL(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	return parsed.Scheme == "https" && strings.EqualFold(parsed.Host, "assets.grok.com") && strings.HasSuffix(parsed.Path, "/content")
}

func isSupportedGrokSessionVideoLength(seconds int) bool {
	switch seconds {
	case 6, 10, 12, 16, 20:
		return true
	default:
		return false
	}
}

func sortedGrokSessionVideoSizes() []string {
	keys := make([]string, 0, len(grokSessionVideoSizeMap))
	for key := range grokSessionVideoSizeMap {
		keys = append(keys, key)
	}
	sortStrings(keys)
	return keys
}

func sortStrings(values []string) {
	if len(values) < 2 {
		return
	}
	for i := 0; i < len(values)-1; i++ {
		for j := i + 1; j < len(values); j++ {
			if values[j] < values[i] {
				values[i], values[j] = values[j], values[i]
			}
		}
	}
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

func (r *GrokSessionMediaRuntime) writeMediaRuntimeError(c *gin.Context, err error) {
	if c == nil {
		return
	}
	var upstreamErr *grokSessionMediaUpstreamError
	if errors.As(err, &upstreamErr) {
		writeCompatibleGatewayMediaError(
			c,
			firstNonZero(upstreamErr.statusCode, http.StatusBadGateway),
			firstNonEmpty(strings.TrimSpace(upstreamErr.code), "api_error"),
			firstNonEmpty(strings.TrimSpace(upstreamErr.message), "Grok session upstream request failed"),
		)
		return
	}
	writeCompatibleGatewayMediaError(c, http.StatusBadGateway, "api_error", firstNonEmpty(strings.TrimSpace(err.Error()), "Grok session upstream request failed"))
}

func (r *GrokSessionMediaRuntime) persistSessionMediaRuntimeFeedback(
	ctx context.Context,
	account *Account,
	requestedModel string,
	endpoint string,
	err error,
) {
	if r == nil || r.gatewayService == nil || r.gatewayService.accountRepo == nil || account == nil {
		return
	}
	feedback := (&OpenAIGatewayService{accountRepo: r.gatewayService.accountRepo})
	feedback.PersistGrokRuntimeFeedback(ctx, GrokRuntimeFeedbackInput{
		Account:        account,
		RequestedModel: strings.TrimSpace(requestedModel),
		StatusCode:     grokSessionMediaFeedbackStatusCode(err),
		Endpoint:       strings.TrimSpace(endpoint),
		Err:            err,
	})
}

func grokSessionMediaFeedbackStatusCode(err error) int {
	var upstreamErr *grokSessionMediaUpstreamError
	if errors.As(err, &upstreamErr) && upstreamErr != nil {
		return upstreamErr.statusCode
	}
	if err == nil {
		return http.StatusOK
	}
	return 0
}

func grokSessionMediaFeedbackCode(err error) string {
	var upstreamErr *grokSessionMediaUpstreamError
	if errors.As(err, &upstreamErr) && upstreamErr != nil {
		return firstNonEmpty(strings.TrimSpace(upstreamErr.code), "api_error")
	}
	return "api_error"
}

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}
