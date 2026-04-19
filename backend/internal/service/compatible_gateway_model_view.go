package service

import (
	"sort"

	"github.com/senran-N/sub2api/internal/pkg/claude"
	"github.com/senran-N/sub2api/internal/pkg/grok"
	"github.com/senran-N/sub2api/internal/pkg/openai"
)

func BuildCompatibleGatewayDiscoveredModels(models []CompatibleUpstreamModel, platform string) any {
	switch NormalizeCompatibleGatewayPlatform(platform) {
	case PlatformOpenAI:
		result := make([]openai.Model, 0, len(models))
		for _, model := range models {
			result = append(result, openai.Model{
				ID:          model.ID,
				Object:      model.Object,
				Created:     model.Created,
				OwnedBy:     model.OwnedBy,
				Type:        model.Type,
				DisplayName: model.DisplayName,
			})
		}
		return result
	case PlatformGrok:
		result := make([]grok.Model, 0, len(models))
		for _, model := range models {
			result = append(result, grok.Model{
				ID:          model.ID,
				Object:      model.Object,
				Created:     model.Created,
				OwnedBy:     model.OwnedBy,
				Type:        model.Type,
				DisplayName: model.DisplayName,
			})
		}
		return result
	default:
		result := make([]claude.Model, 0, len(models))
		for _, model := range models {
			result = append(result, claude.Model{
				ID:          model.ID,
				Type:        model.Type,
				DisplayName: model.DisplayName,
				CreatedAt:   model.CreatedAt,
			})
		}
		return result
	}
}

func BuildCompatibleGatewayMappedModels(modelIDs []string, platform string) any {
	ids := append([]string(nil), modelIDs...)
	sort.Strings(ids)

	switch NormalizeCompatibleGatewayPlatform(platform) {
	case PlatformOpenAI:
		models := make([]openai.Model, 0, len(ids))
		for _, modelID := range ids {
			models = append(models, openai.Model{
				ID:          modelID,
				Object:      "model",
				Type:        "model",
				DisplayName: modelID,
			})
		}
		return models
	case PlatformGrok:
		models := make([]grok.Model, 0, len(ids))
		for _, modelID := range ids {
			models = append(models, grok.Model{
				ID:          modelID,
				Object:      "model",
				Type:        "model",
				DisplayName: modelID,
				OwnedBy:     "xai",
			})
		}
		return models
	default:
		models := make([]claude.Model, 0, len(ids))
		for _, modelID := range ids {
			models = append(models, claude.Model{
				ID:          modelID,
				Type:        "model",
				DisplayName: modelID,
				CreatedAt:   "2024-01-01T00:00:00Z",
			})
		}
		return models
	}
}

func DefaultCompatibleGatewayModels(platform string) any {
	switch NormalizeCompatibleGatewayPlatform(platform) {
	case PlatformOpenAI:
		return openai.DefaultModels
	case PlatformGrok:
		return grok.DefaultModels
	default:
		return claude.DefaultModels
	}
}
