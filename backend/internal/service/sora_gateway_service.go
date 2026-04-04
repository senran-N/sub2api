package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

var soraImageSizeMap = map[string]string{
	"gpt-image":           "360",
	"gpt-image-landscape": "540",
	"gpt-image-portrait":  "540",
}

// SoraGatewayService handles forwarding requests to Sora upstream.
type SoraGatewayService struct {
	soraClient       SoraClient
	rateLimitService *RateLimitService
	httpUpstream     HTTPUpstream // 用于 apikey 类型账号的 HTTP 透传
	cfg              *config.Config
}

type soraWatermarkOptions struct {
	Enabled           bool
	ParseMethod       string
	ParseURL          string
	ParseToken        string
	FallbackOnFailure bool
	DeletePost        bool
}

type soraCharacterOptions struct {
	SetPublic           bool
	DeleteAfterGenerate bool
}

type soraPreflightChecker interface {
	PreflightCheck(ctx context.Context, account *Account, requestedModel string, modelCfg SoraModelConfig) error
}

func NewSoraGatewayService(
	soraClient SoraClient,
	rateLimitService *RateLimitService,
	httpUpstream HTTPUpstream,
	cfg *config.Config,
) *SoraGatewayService {
	return &SoraGatewayService{
		soraClient:       soraClient,
		rateLimitService: rateLimitService,
		httpUpstream:     httpUpstream,
		cfg:              cfg,
	}
}

func (s *SoraGatewayService) Forward(ctx context.Context, c *gin.Context, account *Account, body []byte, clientStream bool) (*ForwardResult, error) {
	startTime := time.Now()

	// apikey 类型账号：HTTP 透传到上游，不走 SoraSDKClient
	if account.Type == AccountTypeAPIKey && account.GetBaseURL() != "" {
		if s.httpUpstream == nil {
			s.writeSoraError(c, http.StatusInternalServerError, "api_error", "HTTP upstream client not configured", clientStream)
			return nil, errors.New("httpUpstream not configured for sora apikey forwarding")
		}
		return s.forwardToUpstream(ctx, c, account, body, clientStream, startTime)
	}

	if s.soraClient == nil || !s.soraClient.Enabled() {
		if c != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": gin.H{
					"type":    "api_error",
					"message": "Sora 上游未配置",
				},
			})
		}
		return nil, errors.New("sora upstream not configured")
	}

	var reqBody map[string]any
	if err := json.Unmarshal(body, &reqBody); err != nil {
		s.writeSoraError(c, http.StatusBadRequest, "invalid_request_error", "Failed to parse request body", clientStream)
		return nil, fmt.Errorf("parse request: %w", err)
	}
	reqModel, _ := reqBody["model"].(string)
	reqStream, _ := reqBody["stream"].(bool)
	if strings.TrimSpace(reqModel) == "" {
		s.writeSoraError(c, http.StatusBadRequest, "invalid_request_error", "model is required", clientStream)
		return nil, errors.New("model is required")
	}
	originalModel := reqModel

	mappedModel := account.GetMappedModel(reqModel)
	var upstreamModel string
	if mappedModel != "" && mappedModel != reqModel {
		reqModel = mappedModel
		upstreamModel = mappedModel
	}

	modelCfg, ok := GetSoraModelConfig(reqModel)
	if !ok {
		s.writeSoraError(c, http.StatusBadRequest, "invalid_request_error", "Unsupported Sora model", clientStream)
		return nil, fmt.Errorf("unsupported model: %s", reqModel)
	}
	prompt, imageInput, videoInput, remixTargetID := extractSoraInput(reqBody)
	prompt = strings.TrimSpace(prompt)
	imageInput = strings.TrimSpace(imageInput)
	videoInput = strings.TrimSpace(videoInput)
	remixTargetID = strings.TrimSpace(remixTargetID)

	if videoInput != "" && modelCfg.Type != "video" {
		s.writeSoraError(c, http.StatusBadRequest, "invalid_request_error", "video input only supports video models", clientStream)
		return nil, errors.New("video input only supports video models")
	}
	if videoInput != "" && imageInput != "" {
		s.writeSoraError(c, http.StatusBadRequest, "invalid_request_error", "image input and video input cannot be used together", clientStream)
		return nil, errors.New("image input and video input cannot be used together")
	}
	characterOnly := videoInput != "" && prompt == ""
	if modelCfg.Type == "prompt_enhance" && prompt == "" {
		s.writeSoraError(c, http.StatusBadRequest, "invalid_request_error", "prompt is required", clientStream)
		return nil, errors.New("prompt is required")
	}
	if modelCfg.Type != "prompt_enhance" && prompt == "" && !characterOnly {
		s.writeSoraError(c, http.StatusBadRequest, "invalid_request_error", "prompt is required", clientStream)
		return nil, errors.New("prompt is required")
	}

	reqCtx, cancel := s.withSoraTimeout(ctx, reqStream)
	if cancel != nil {
		defer cancel()
	}
	if checker, ok := s.soraClient.(soraPreflightChecker); ok && !characterOnly {
		if err := checker.PreflightCheck(reqCtx, account, reqModel, modelCfg); err != nil {
			return nil, s.handleSoraRequestError(ctx, account, err, reqModel, c, clientStream)
		}
	}

	if modelCfg.Type == "prompt_enhance" {
		enhancedPrompt, err := s.soraClient.EnhancePrompt(reqCtx, account, prompt, modelCfg.ExpansionLevel, modelCfg.DurationS)
		if err != nil {
			return nil, s.handleSoraRequestError(ctx, account, err, reqModel, c, clientStream)
		}
		content := strings.TrimSpace(enhancedPrompt)
		if content == "" {
			content = prompt
		}
		firstTokenMs, responseErr := s.writeSoraCompletionResponse(c, reqModel, content, startTime, clientStream, nil)
		if responseErr != nil {
			return nil, responseErr
		}
		return buildSoraPromptForwardResult(startTime, originalModel, upstreamModel, clientStream, firstTokenMs), nil
	}

	characterOpts := parseSoraCharacterOptions(reqBody)
	watermarkOpts := parseSoraWatermarkOptions(reqBody)
	var characterResult *soraCharacterFlowResult
	if videoInput != "" {
		videoData, videoErr := decodeSoraVideoInput(reqCtx, videoInput)
		if videoErr != nil {
			s.writeSoraError(c, http.StatusBadRequest, "invalid_request_error", videoErr.Error(), clientStream)
			return nil, videoErr
		}
		characterResult, videoErr = s.createCharacterFromVideo(reqCtx, account, videoData, characterOpts)
		if videoErr != nil {
			return nil, s.handleSoraRequestError(ctx, account, videoErr, reqModel, c, clientStream)
		}
		if characterResult != nil && characterOpts.DeleteAfterGenerate && strings.TrimSpace(characterResult.CharacterID) != "" && !characterOnly {
			characterID := strings.TrimSpace(characterResult.CharacterID)
			defer func() {
				cleanupCtx, cancelCleanup := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancelCleanup()
				if err := s.soraClient.DeleteCharacter(cleanupCtx, account, characterID); err != nil {
					log.Printf("[Sora] cleanup character failed, character_id=%s err=%v", characterID, err)
				}
			}()
		}
		if characterOnly {
			content := "角色创建成功"
			if characterResult != nil && strings.TrimSpace(characterResult.Username) != "" {
				content = fmt.Sprintf("角色创建成功，角色名@%s", strings.TrimSpace(characterResult.Username))
			}
			firstTokenMs, responseErr := s.writeSoraCompletionResponse(c, reqModel, content, startTime, clientStream, soraCharacterResponseFields(characterResult))
			if responseErr != nil {
				return nil, responseErr
			}
			return buildSoraPromptForwardResult(startTime, originalModel, upstreamModel, clientStream, firstTokenMs), nil
		}
		if characterResult != nil && strings.TrimSpace(characterResult.Username) != "" {
			prompt = fmt.Sprintf("@%s %s", characterResult.Username, prompt)
		}
	}

	var imageData []byte
	imageFilename := ""
	if imageInput != "" {
		decoded, filename, err := decodeSoraImageInput(reqCtx, imageInput)
		if err != nil {
			s.writeSoraError(c, http.StatusBadRequest, "invalid_request_error", err.Error(), clientStream)
			return nil, err
		}
		imageData = decoded
		imageFilename = filename
	}

	mediaID := ""
	if len(imageData) > 0 {
		uploadID, err := s.soraClient.UploadImage(reqCtx, account, imageData, imageFilename)
		if err != nil {
			return nil, s.handleSoraRequestError(ctx, account, err, reqModel, c, clientStream)
		}
		mediaID = uploadID
	}

	taskID := ""
	var err error
	videoCount := parseSoraVideoCount(reqBody)
	switch modelCfg.Type {
	case "image":
		taskID, err = s.soraClient.CreateImageTask(reqCtx, account, SoraImageRequest{
			Prompt:  prompt,
			Width:   modelCfg.Width,
			Height:  modelCfg.Height,
			MediaID: mediaID,
		})
	case "video":
		if remixTargetID == "" && isSoraStoryboardPrompt(prompt) {
			taskID, err = s.soraClient.CreateStoryboardTask(reqCtx, account, SoraStoryboardRequest{
				Prompt:      formatSoraStoryboardPrompt(prompt),
				Orientation: modelCfg.Orientation,
				Frames:      modelCfg.Frames,
				Model:       modelCfg.Model,
				Size:        modelCfg.Size,
				MediaID:     mediaID,
			})
		} else {
			taskID, err = s.soraClient.CreateVideoTask(reqCtx, account, SoraVideoRequest{
				Prompt:        prompt,
				Orientation:   modelCfg.Orientation,
				Frames:        modelCfg.Frames,
				Model:         modelCfg.Model,
				Size:          modelCfg.Size,
				VideoCount:    videoCount,
				MediaID:       mediaID,
				RemixTargetID: remixTargetID,
				CameoIDs:      extractSoraCameoIDs(reqBody),
			})
		}
	default:
		err = fmt.Errorf("unsupported model type: %s", modelCfg.Type)
	}
	if err != nil {
		return nil, s.handleSoraRequestError(ctx, account, err, reqModel, c, clientStream)
	}

	if clientStream && c != nil {
		s.prepareSoraStream(c, taskID)
	}

	var mediaURLs []string
	videoGenerationID := ""
	mediaType := modelCfg.Type
	imageCount := 0
	imageSize := ""
	switch modelCfg.Type {
	case "image":
		urls, pollErr := s.pollImageTask(reqCtx, c, account, taskID, clientStream)
		if pollErr != nil {
			return nil, s.handleSoraRequestError(ctx, account, pollErr, reqModel, c, clientStream)
		}
		mediaURLs = urls
		imageCount = len(urls)
		imageSize = soraImageSizeFromModel(reqModel)
	case "video":
		videoStatus, pollErr := s.pollVideoTaskDetailed(reqCtx, c, account, taskID, clientStream)
		if pollErr != nil {
			return nil, s.handleSoraRequestError(ctx, account, pollErr, reqModel, c, clientStream)
		}
		if videoStatus != nil {
			mediaURLs = videoStatus.URLs
			videoGenerationID = strings.TrimSpace(videoStatus.GenerationID)
		}
	default:
		mediaType = "prompt"
	}

	watermarkPostID := ""
	if modelCfg.Type == "video" && watermarkOpts.Enabled {
		watermarkURL, postID, watermarkErr := s.resolveWatermarkFreeURL(reqCtx, account, videoGenerationID, watermarkOpts)
		if watermarkErr != nil {
			if !watermarkOpts.FallbackOnFailure {
				return nil, s.handleSoraRequestError(ctx, account, watermarkErr, reqModel, c, clientStream)
			}
			log.Printf("[Sora] watermark-free fallback to original URL, task_id=%s err=%v", taskID, watermarkErr)
		} else if strings.TrimSpace(watermarkURL) != "" {
			mediaURLs = []string{strings.TrimSpace(watermarkURL)}
			watermarkPostID = strings.TrimSpace(postID)
		}
	}

	// 直调路径（/sora/v1/chat/completions）保持纯透传，不执行本地/S3 媒体落盘。
	// 媒体存储由客户端 API 路径（/api/v1/sora/generate）的异步流程负责。
	finalURLs := s.normalizeSoraMediaURLs(mediaURLs)
	if watermarkPostID != "" && watermarkOpts.DeletePost {
		if deleteErr := s.soraClient.DeletePost(reqCtx, account, watermarkPostID); deleteErr != nil {
			log.Printf("[Sora] delete post failed, post_id=%s err=%v", watermarkPostID, deleteErr)
		}
	}

	content := buildSoraContent(mediaType, finalURLs)
	firstTokenMs, responseErr := s.writeSoraCompletionResponse(c, reqModel, content, startTime, clientStream, soraMediaResponseFields(finalURLs))
	if responseErr != nil {
		return nil, responseErr
	}

	return &ForwardResult{
		RequestID:     taskID,
		Model:         originalModel,
		UpstreamModel: upstreamModel,
		Stream:        clientStream,
		Duration:      time.Since(startTime),
		FirstTokenMs:  firstTokenMs,
		Usage:         ClaudeUsage{},
		MediaType:     mediaType,
		MediaURL:      firstMediaURL(finalURLs),
		ImageCount:    imageCount,
		ImageSize:     imageSize,
	}, nil
}

func (s *SoraGatewayService) withSoraTimeout(ctx context.Context, stream bool) (context.Context, context.CancelFunc) {
	if s == nil || s.cfg == nil {
		return ctx, nil
	}
	timeoutSeconds := s.cfg.Gateway.SoraRequestTimeoutSeconds
	if stream {
		timeoutSeconds = s.cfg.Gateway.SoraStreamTimeoutSeconds
	}
	if timeoutSeconds <= 0 {
		return ctx, nil
	}
	return context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
}

func (s *SoraGatewayService) resolveWatermarkFreeURL(ctx context.Context, account *Account, generationID string, opts soraWatermarkOptions) (string, string, error) {
	generationID = strings.TrimSpace(generationID)
	if generationID == "" {
		return "", "", errors.New("generation id is required for watermark-free mode")
	}
	postID, err := s.soraClient.PostVideoForWatermarkFree(ctx, account, generationID)
	if err != nil {
		return "", "", err
	}
	postID = strings.TrimSpace(postID)
	if postID == "" {
		return "", "", errors.New("watermark-free publish returned empty post id")
	}

	switch opts.ParseMethod {
	case "custom":
		urlVal, parseErr := s.soraClient.GetWatermarkFreeURLCustom(ctx, account, opts.ParseURL, opts.ParseToken, postID)
		if parseErr != nil {
			return "", postID, parseErr
		}
		return strings.TrimSpace(urlVal), postID, nil
	case "", "third_party":
		return fmt.Sprintf("https://oscdn2.dyysy.com/MP4/%s.mp4", postID), postID, nil
	default:
		return "", postID, fmt.Errorf("unsupported watermark parse method: %s", opts.ParseMethod)
	}
}

func (s *SoraGatewayService) shouldFailoverUpstreamError(statusCode int) bool {
	switch statusCode {
	case 401, 402, 403, 404, 429, 529:
		return true
	default:
		return statusCode >= 500
	}
}

func soraProErrorMessage(model, upstreamMsg string) string {
	modelLower := strings.ToLower(model)
	if strings.Contains(modelLower, "sora2pro-hd") {
		return "当前账号无法使用 Sora Pro-HD 模型，请更换模型或账号"
	}
	if strings.Contains(modelLower, "sora2pro") {
		return "当前账号无法使用 Sora Pro 模型，请更换模型或账号"
	}
	return ""
}

func (s *SoraGatewayService) prepareSoraStream(c *gin.Context, requestID string) {
	if c == nil {
		return
	}
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	if strings.TrimSpace(requestID) != "" {
		c.Header("x-request-id", requestID)
	}
}

func (s *SoraGatewayService) writeSoraStream(c *gin.Context, model, content string, startTime time.Time) (*int, error) {
	if c == nil {
		return nil, nil
	}
	writer := c.Writer
	flusher, _ := writer.(http.Flusher)

	chunk := map[string]any{
		"id":      fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
		"object":  "chat.completion.chunk",
		"created": time.Now().Unix(),
		"model":   model,
		"choices": []any{
			map[string]any{
				"index": 0,
				"delta": map[string]any{
					"content": content,
				},
			},
		},
	}
	encoded, _ := jsonMarshalRaw(chunk)
	if _, err := fmt.Fprintf(writer, "data: %s\n\n", encoded); err != nil {
		return nil, err
	}
	if flusher != nil {
		flusher.Flush()
	}
	ms := int(time.Since(startTime).Milliseconds())
	finalChunk := map[string]any{
		"id":      chunk["id"],
		"object":  "chat.completion.chunk",
		"created": time.Now().Unix(),
		"model":   model,
		"choices": []any{
			map[string]any{
				"index":         0,
				"delta":         map[string]any{},
				"finish_reason": "stop",
			},
		},
	}
	finalEncoded, _ := jsonMarshalRaw(finalChunk)
	if _, err := fmt.Fprintf(writer, "data: %s\n\n", finalEncoded); err != nil {
		return &ms, err
	}
	if _, err := fmt.Fprint(writer, "data: [DONE]\n\n"); err != nil {
		return &ms, err
	}
	if flusher != nil {
		flusher.Flush()
	}
	return &ms, nil
}

func (s *SoraGatewayService) writeSoraError(c *gin.Context, status int, errType, message string, stream bool) {
	if c == nil {
		return
	}
	if stream {
		flusher, _ := c.Writer.(http.Flusher)
		errorData := map[string]any{
			"error": map[string]string{
				"type":    errType,
				"message": message,
			},
		}
		jsonBytes, err := json.Marshal(errorData)
		if err != nil {
			_ = c.Error(err)
			return
		}
		errorEvent := fmt.Sprintf("event: error\ndata: %s\n\n", string(jsonBytes))
		_, _ = fmt.Fprint(c.Writer, errorEvent)
		_, _ = fmt.Fprint(c.Writer, "data: [DONE]\n\n")
		if flusher != nil {
			flusher.Flush()
		}
		return
	}
	c.JSON(status, gin.H{
		"error": gin.H{
			"type":    errType,
			"message": message,
		},
	})
}

func (s *SoraGatewayService) handleSoraRequestError(ctx context.Context, account *Account, err error, model string, c *gin.Context, stream bool) error {
	if err == nil {
		return nil
	}
	var upstreamErr *SoraUpstreamError
	if errors.As(err, &upstreamErr) {
		accountID := int64(0)
		if account != nil {
			accountID = account.ID
		}
		logger.LegacyPrintf(
			"service.sora",
			"[SoraRawError] account_id=%d model=%s status=%d request_id=%s cf_ray=%s message=%s raw_body=%s",
			accountID,
			model,
			upstreamErr.StatusCode,
			strings.TrimSpace(upstreamErr.Headers.Get("x-request-id")),
			strings.TrimSpace(upstreamErr.Headers.Get("cf-ray")),
			strings.TrimSpace(upstreamErr.Message),
			truncateForLog(upstreamErr.Body, 1024),
		)
		if s.rateLimitService != nil && account != nil {
			s.rateLimitService.HandleUpstreamError(ctx, account, upstreamErr.StatusCode, upstreamErr.Headers, upstreamErr.Body)
		}
		if s.shouldFailoverUpstreamError(upstreamErr.StatusCode) {
			var responseHeaders http.Header
			if upstreamErr.Headers != nil {
				responseHeaders = upstreamErr.Headers.Clone()
			}
			return &UpstreamFailoverError{
				StatusCode:      upstreamErr.StatusCode,
				ResponseBody:    upstreamErr.Body,
				ResponseHeaders: responseHeaders,
			}
		}
		msg := upstreamErr.Message
		if override := soraProErrorMessage(model, msg); override != "" {
			msg = override
		}
		s.writeSoraError(c, upstreamErr.StatusCode, "upstream_error", msg, stream)
		return err
	}
	if errors.Is(err, context.DeadlineExceeded) {
		s.writeSoraError(c, http.StatusGatewayTimeout, "timeout_error", "Sora generation timeout", stream)
		return err
	}
	s.writeSoraError(c, http.StatusBadGateway, "api_error", err.Error(), stream)
	return err
}
