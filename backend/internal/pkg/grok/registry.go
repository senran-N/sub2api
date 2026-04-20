package grok

import (
	"slices"
	"strings"
)

type Capability string

const (
	CapabilityChat      Capability = "chat"
	CapabilityImage     Capability = "image"
	CapabilityImageEdit Capability = "image_edit"
	CapabilityVideo     Capability = "video"
	CapabilityVoice     Capability = "voice"
)

type Tier string

const (
	TierUnknown Tier = "unknown"
	TierBasic   Tier = "basic"
	TierSuper   Tier = "super"
	TierHeavy   Tier = "heavy"
)

type ProtocolFamily string

const (
	ProtocolFamilyChatCompletions ProtocolFamily = "chat_completions"
	ProtocolFamilyResponses       ProtocolFamily = "responses"
	ProtocolFamilyMessages        ProtocolFamily = "messages"
	ProtocolFamilyMediaJob        ProtocolFamily = "media_job"
)

type ModelSpec struct {
	ID                  string
	Aliases             []string
	Object              string
	Created             int64
	OwnedBy             string
	Type                string
	DisplayName         string
	Capability          Capability
	RequiredTier        Tier
	AllowedAccountTypes []string
	ProtocolFamily      ProtocolFamily
	QuotaWindow         string
	SupportsStream      bool
	SupportsTools       bool
}

var modelSpecs = []ModelSpec{
	{
		ID:                  "grok-4.20-0309-non-reasoning",
		Object:              "model",
		Created:             1741478400,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok 4.20 0309 Non-Reasoning",
		Capability:          CapabilityChat,
		RequiredTier:        TierBasic,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyChatCompletions,
		QuotaWindow:         QuotaWindowFast,
		SupportsStream:      true,
		SupportsTools:       true,
	},
	{
		ID:                  "grok-4.20-0309",
		Object:              "model",
		Created:             1741478400,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok 4.20 0309",
		Capability:          CapabilityChat,
		RequiredTier:        TierBasic,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyResponses,
		QuotaWindow:         QuotaWindowAuto,
		SupportsStream:      true,
		SupportsTools:       true,
	},
	{
		ID:                  "grok-4.20-0309-reasoning",
		Object:              "model",
		Created:             1741478400,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok 4.20 0309 Reasoning",
		Capability:          CapabilityChat,
		RequiredTier:        TierBasic,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyResponses,
		QuotaWindow:         QuotaWindowExpert,
		SupportsStream:      true,
		SupportsTools:       true,
	},
	{
		ID:                  "grok-4.20-0309-non-reasoning-super",
		Object:              "model",
		Created:             1741478400,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok 4.20 0309 Non-Reasoning Super",
		Capability:          CapabilityChat,
		RequiredTier:        TierSuper,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyChatCompletions,
		QuotaWindow:         QuotaWindowFast,
		SupportsStream:      true,
		SupportsTools:       true,
	},
	{
		ID:                  "grok-4.20-0309-super",
		Object:              "model",
		Created:             1741478400,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok 4.20 0309 Super",
		Capability:          CapabilityChat,
		RequiredTier:        TierSuper,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyResponses,
		QuotaWindow:         QuotaWindowAuto,
		SupportsStream:      true,
		SupportsTools:       true,
	},
	{
		ID:                  "grok-4.20-0309-reasoning-super",
		Object:              "model",
		Created:             1741478400,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok 4.20 0309 Reasoning Super",
		Capability:          CapabilityChat,
		RequiredTier:        TierSuper,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyResponses,
		QuotaWindow:         QuotaWindowExpert,
		SupportsStream:      true,
		SupportsTools:       true,
	},
	{
		ID:                  "grok-4.20-0309-non-reasoning-heavy",
		Object:              "model",
		Created:             1741478400,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok 4.20 0309 Non-Reasoning Heavy",
		Capability:          CapabilityChat,
		RequiredTier:        TierHeavy,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyChatCompletions,
		QuotaWindow:         QuotaWindowFast,
		SupportsStream:      true,
		SupportsTools:       true,
	},
	{
		ID:                  "grok-4.20-0309-heavy",
		Object:              "model",
		Created:             1741478400,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok 4.20 0309 Heavy",
		Capability:          CapabilityChat,
		RequiredTier:        TierHeavy,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyResponses,
		QuotaWindow:         QuotaWindowAuto,
		SupportsStream:      true,
		SupportsTools:       true,
	},
	{
		ID:                  "grok-4.20-0309-reasoning-heavy",
		Object:              "model",
		Created:             1741478400,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok 4.20 0309 Reasoning Heavy",
		Capability:          CapabilityChat,
		RequiredTier:        TierHeavy,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyResponses,
		QuotaWindow:         QuotaWindowExpert,
		SupportsStream:      true,
		SupportsTools:       true,
	},
	{
		ID:                  "grok-4.20-multi-agent-0309",
		Object:              "model",
		Created:             1741478400,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok 4.20 Multi-Agent 0309",
		Capability:          CapabilityChat,
		RequiredTier:        TierHeavy,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyResponses,
		QuotaWindow:         QuotaWindowHeavy,
		SupportsStream:      true,
		SupportsTools:       true,
	},
	{
		ID:                  "grok-4.20-fast",
		Aliases:             []string{"grok-3-fast"},
		Object:              "model",
		Created:             1741478400,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok 4.20 Fast",
		Capability:          CapabilityChat,
		RequiredTier:        TierBasic,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyChatCompletions,
		QuotaWindow:         QuotaWindowFast,
		SupportsStream:      true,
		SupportsTools:       true,
	},
	{
		ID:                  "grok-4.20-auto",
		Aliases:             []string{"grok-3"},
		Object:              "model",
		Created:             1741478400,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok 4.20 Auto",
		Capability:          CapabilityChat,
		RequiredTier:        TierBasic,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyResponses,
		QuotaWindow:         QuotaWindowAuto,
		SupportsStream:      true,
		SupportsTools:       true,
	},
	{
		ID:                  "grok-4.20-expert",
		Aliases:             []string{"grok-4-fast-reasoning"},
		Object:              "model",
		Created:             1741478400,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok 4.20 Expert",
		Capability:          CapabilityChat,
		RequiredTier:        TierBasic,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyResponses,
		QuotaWindow:         QuotaWindowExpert,
		SupportsStream:      true,
		SupportsTools:       true,
	},
	{
		ID:                  "grok-4.20-heavy",
		Object:              "model",
		Created:             1741478400,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok 4.20 Heavy",
		Capability:          CapabilityChat,
		RequiredTier:        TierHeavy,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyResponses,
		QuotaWindow:         QuotaWindowHeavy,
		SupportsStream:      true,
		SupportsTools:       true,
	},
	{
		ID:                  "grok-4-voice",
		Object:              "model",
		Created:             1738368000,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok 4 Voice",
		Capability:          CapabilityVoice,
		RequiredTier:        TierSuper,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyResponses,
		QuotaWindow:         QuotaWindowAuto,
		SupportsStream:      true,
	},
	{
		ID:                  "grok-imagine-image-lite",
		Aliases:             []string{"grok-2-image"},
		Object:              "model",
		Created:             1741478400,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok Imagine Image Lite",
		Capability:          CapabilityImage,
		RequiredTier:        TierBasic,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyResponses,
		QuotaWindow:         QuotaWindowFast,
		SupportsStream:      false,
	},
	{
		ID:                  "grok-imagine-image",
		Object:              "model",
		Created:             1741478400,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok Imagine Image",
		Capability:          CapabilityImage,
		RequiredTier:        TierSuper,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyResponses,
		QuotaWindow:         QuotaWindowAuto,
		SupportsStream:      false,
	},
	{
		ID:                  "grok-imagine-image-pro",
		Object:              "model",
		Created:             1741478400,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok Imagine Image Pro",
		Capability:          CapabilityImage,
		RequiredTier:        TierSuper,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyResponses,
		QuotaWindow:         QuotaWindowAuto,
		SupportsStream:      false,
	},
	{
		ID:                  "grok-imagine-image-edit",
		Object:              "model",
		Created:             1741478400,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok Imagine Image Edit",
		Capability:          CapabilityImageEdit,
		RequiredTier:        TierSuper,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyResponses,
		QuotaWindow:         QuotaWindowAuto,
		SupportsStream:      false,
	},
	{
		ID:                  "grok-imagine-video",
		Object:              "model",
		Created:             1741478400,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok Imagine Video",
		Capability:          CapabilityVideo,
		RequiredTier:        TierSuper,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyMediaJob,
		QuotaWindow:         QuotaWindowAuto,
		SupportsStream:      false,
	},
}

var (
	specsByID    map[string]ModelSpec
	aliasesByKey map[string]string
)

func init() {
	specsByID = make(map[string]ModelSpec, len(modelSpecs))
	aliasesByKey = make(map[string]string)
	for _, spec := range modelSpecs {
		specsByID[normalizeModelKey(spec.ID)] = spec
		for _, alias := range spec.Aliases {
			aliasesByKey[normalizeModelKey(alias)] = spec.ID
		}
	}
}

func normalizeModelKey(model string) string {
	return strings.ToLower(strings.TrimSpace(model))
}

func Specs() []ModelSpec {
	return slices.Clone(modelSpecs)
}

func LookupModelSpec(model string) (ModelSpec, bool) {
	key := normalizeModelKey(model)
	if key == "" {
		return ModelSpec{}, false
	}
	if spec, ok := specsByID[key]; ok {
		return spec, true
	}
	if canonicalID, ok := aliasesByKey[key]; ok {
		spec, ok := specsByID[normalizeModelKey(canonicalID)]
		return spec, ok
	}
	return ModelSpec{}, false
}

func ResolveCanonicalModelID(model string) string {
	spec, ok := LookupModelSpec(model)
	if !ok {
		return strings.TrimSpace(model)
	}
	return spec.ID
}

func EnabledModels() []Model {
	models := make([]Model, 0, len(modelSpecs))
	for _, spec := range modelSpecs {
		models = append(models, Model{
			ID:          spec.ID,
			Object:      spec.Object,
			Created:     spec.Created,
			OwnedBy:     spec.OwnedBy,
			Type:        spec.Type,
			DisplayName: spec.DisplayName,
		})
	}
	return models
}

func EnabledModelIDs() []string {
	ids := make([]string, 0, len(modelSpecs))
	for _, spec := range modelSpecs {
		ids = append(ids, spec.ID)
	}
	return ids
}

func DefaultFallbackModelID() string {
	for _, spec := range modelSpecs {
		if spec.Capability == CapabilityChat && spec.RequiredTier == TierBasic {
			return spec.ID
		}
	}
	for _, spec := range modelSpecs {
		if spec.Capability == CapabilityChat {
			return spec.ID
		}
	}
	if len(modelSpecs) == 0 {
		return ""
	}
	return modelSpecs[0].ID
}
