package service

import "github.com/gin-gonic/gin"

// GrokGatewayService stays as the handler-facing Grok entrypoint while
// delegating actual text execution to Grok-owned runtimes.
type GrokGatewayService struct {
	textRuntime  *GrokTextRuntime
	mediaService *GrokMediaService
}

func NewGrokGatewayService(gatewayService *GatewayService, compatibleTextRuntime *CompatibleGatewayTextRuntime) *GrokGatewayService {
	return &GrokGatewayService{
		textRuntime: NewGrokTextRuntime(
			gatewayService,
			ProvideGrokCompatibleRuntime(compatibleTextRuntime),
			NewGrokSessionRuntime(gatewayService),
		),
		mediaService: NewGrokMediaService(gatewayService, nil, nil),
	}
}

func ProvideGrokGatewayService(
	textRuntime *GrokTextRuntime,
	gatewayService *GatewayService,
	videoJobs GrokVideoJobRepository,
	mediaAssets GrokMediaAssetRepository,
) *GrokGatewayService {
	return &GrokGatewayService{
		textRuntime:  textRuntime,
		mediaService: NewGrokMediaService(gatewayService, videoJobs, mediaAssets),
	}
}

func NewGrokGatewayServiceWithCompatibleExecutor(
	gatewayService *GatewayService,
	compatibleTextExecutor grokCompatibleTextExecutor,
) *GrokGatewayService {
	return &GrokGatewayService{
		textRuntime: NewGrokTextRuntime(
			gatewayService,
			NewGrokCompatibleRuntime(compatibleTextExecutor),
			NewGrokSessionRuntime(gatewayService),
		),
		mediaService: NewGrokMediaService(gatewayService, nil, nil),
	}
}

func (s *GrokGatewayService) HandleResponses(c *gin.Context, groupID *int64, body []byte) bool {
	if s == nil || s.textRuntime == nil {
		return false
	}
	return s.textRuntime.HandleResponses(c, groupID, body)
}

func (s *GrokGatewayService) HandleChatCompletions(c *gin.Context, groupID *int64, body []byte) bool {
	if s == nil {
		return false
	}
	if s.handleChatCompletionsMedia(c, groupID, body) {
		return true
	}
	if s.textRuntime == nil {
		return false
	}
	return s.textRuntime.HandleChatCompletions(c, groupID, body)
}

func (s *GrokGatewayService) HandleMessages(c *gin.Context, groupID *int64, body []byte) bool {
	if s == nil || s.textRuntime == nil {
		return false
	}
	return s.textRuntime.HandleMessages(c, groupID, body)
}

func (s *GrokGatewayService) HandleImages(c *gin.Context, groupID *int64, body []byte) bool {
	if s == nil || s.mediaService == nil {
		return false
	}
	return s.mediaService.HandleImages(c, groupID, body)
}

func (s *GrokGatewayService) HandleVideos(c *gin.Context, groupID *int64, body []byte) bool {
	if s == nil || s.mediaService == nil {
		return false
	}
	return s.mediaService.HandleVideos(c, groupID, body)
}

func (s *GrokGatewayService) HandleMediaAssetContent(c *gin.Context, assetID string) bool {
	if s == nil || s.mediaService == nil {
		return false
	}
	return s.mediaService.HandleAssetContent(c, assetID)
}
