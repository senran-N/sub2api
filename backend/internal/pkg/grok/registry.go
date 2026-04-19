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
	SupportsStream      bool
	SupportsTools       bool
}

var modelSpecs = []ModelSpec{
	{
		ID:                  "grok-3",
		Aliases:             []string{"grok-4.20-auto"},
		Object:              "model",
		Created:             1736294400,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok 3",
		Capability:          CapabilityChat,
		RequiredTier:        TierBasic,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyResponses,
		SupportsStream:      true,
		SupportsTools:       true,
	},
	{
		ID:                  "grok-3-fast",
		Aliases:             []string{"grok-4.20-fast"},
		Object:              "model",
		Created:             1736294400,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok 3 Fast",
		Capability:          CapabilityChat,
		RequiredTier:        TierBasic,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyChatCompletions,
		SupportsStream:      true,
		SupportsTools:       true,
	},
	{
		ID:                  "grok-4-fast-reasoning",
		Aliases:             []string{"grok-4.20-expert", "grok-4.20-heavy"},
		Object:              "model",
		Created:             1738368000,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok 4 Fast Reasoning",
		Capability:          CapabilityChat,
		RequiredTier:        TierHeavy,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyResponses,
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
		SupportsStream:      true,
	},
	{
		ID:                  "grok-2-image",
		Aliases:             []string{"grok-imagine-image-lite"},
		Object:              "model",
		Created:             1727740800,
		OwnedBy:             "xai",
		Type:                "model",
		DisplayName:         "Grok 2 Image",
		Capability:          CapabilityImage,
		RequiredTier:        TierBasic,
		AllowedAccountTypes: []string{"apikey", "upstream", "session"},
		ProtocolFamily:      ProtocolFamilyResponses,
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
