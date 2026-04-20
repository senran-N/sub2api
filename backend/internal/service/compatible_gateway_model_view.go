package service

import (
	"sort"
	"strings"

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

func LookupCompatibleGatewayDiscoveredModel(models []CompatibleUpstreamModel, modelID, platform string) (any, bool) {
	targetKey := compatibleGatewayModelLookupKey(modelID, platform)
	if targetKey == "" {
		return nil, false
	}

	for _, model := range models {
		if compatibleGatewayModelLookupKey(model.ID, platform) != targetKey {
			continue
		}
		switch NormalizeCompatibleGatewayPlatform(platform) {
		case PlatformOpenAI:
			return openai.Model{
				ID:          model.ID,
				Object:      model.Object,
				Created:     model.Created,
				OwnedBy:     model.OwnedBy,
				Type:        model.Type,
				DisplayName: model.DisplayName,
			}, true
		case PlatformGrok:
			return grok.Model{
				ID:          model.ID,
				Object:      model.Object,
				Created:     model.Created,
				OwnedBy:     model.OwnedBy,
				Type:        model.Type,
				DisplayName: model.DisplayName,
			}, true
		default:
			return claude.Model{
				ID:          model.ID,
				Type:        model.Type,
				DisplayName: model.DisplayName,
				CreatedAt:   model.CreatedAt,
			}, true
		}
	}

	return nil, false
}

func LookupCompatibleGatewayMappedModel(modelIDs []string, modelID, platform string) (any, bool) {
	targetKey := compatibleGatewayModelLookupKey(modelID, platform)
	if targetKey == "" {
		return nil, false
	}

	for _, candidateID := range modelIDs {
		if compatibleGatewayModelLookupKey(candidateID, platform) != targetKey {
			continue
		}
		if model, ok := LookupCompatibleGatewayDefaultModel(candidateID, platform); ok {
			return model, true
		}
		switch NormalizeCompatibleGatewayPlatform(platform) {
		case PlatformOpenAI:
			return openai.Model{
				ID:          candidateID,
				Object:      "model",
				Type:        "model",
				DisplayName: candidateID,
			}, true
		case PlatformGrok:
			return grok.Model{
				ID:          candidateID,
				Object:      "model",
				Type:        "model",
				OwnedBy:     "xai",
				DisplayName: candidateID,
			}, true
		default:
			return claude.Model{
				ID:          candidateID,
				Type:        "model",
				DisplayName: candidateID,
				CreatedAt:   "2024-01-01T00:00:00Z",
			}, true
		}
	}

	return nil, false
}

func LookupCompatibleGatewayDefaultModel(modelID, platform string) (any, bool) {
	targetKey := compatibleGatewayModelLookupKey(modelID, platform)
	if targetKey == "" {
		return nil, false
	}

	switch NormalizeCompatibleGatewayPlatform(platform) {
	case PlatformOpenAI:
		for _, model := range openai.DefaultModels {
			if compatibleGatewayModelLookupKey(model.ID, platform) == targetKey {
				return model, true
			}
		}
	case PlatformGrok:
		spec, ok := grok.LookupModelSpec(modelID)
		if !ok {
			return nil, false
		}
		return grok.Model{
			ID:          spec.ID,
			Object:      spec.Object,
			Created:     spec.Created,
			OwnedBy:     spec.OwnedBy,
			Type:        spec.Type,
			DisplayName: spec.DisplayName,
		}, true
	default:
		for _, model := range claude.DefaultModels {
			if compatibleGatewayModelLookupKey(model.ID, platform) == targetKey {
				return model, true
			}
		}
	}

	return nil, false
}

func compatibleGatewayModelLookupKey(modelID, platform string) string {
	switch NormalizeCompatibleGatewayPlatform(platform) {
	case PlatformGrok:
		return strings.ToLower(strings.TrimSpace(grok.ResolveCanonicalModelID(modelID)))
	case PlatformOpenAI:
		return strings.ToLower(strings.TrimSpace(modelID))
	default:
		return strings.ToLower(strings.TrimSpace(claude.NormalizeModelID(modelID)))
	}
}
