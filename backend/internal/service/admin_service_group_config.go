package service

import "context"

func normalizeCreateGroupPlatform(platform string) string {
	if platform == "" {
		return PlatformAnthropic
	}
	return platform
}

func normalizeCreateGroupSubscriptionType(subscriptionType string) string {
	if subscriptionType == "" {
		return SubscriptionTypeStandard
	}
	return subscriptionType
}

func normalizeCreateGroupMCPXMLInject(mcpXMLInject *bool) bool {
	if mcpXMLInject == nil {
		return true
	}
	return *mcpXMLInject
}

func normalizePositiveOptionalID(value *int64) *int64 {
	if value == nil || *value <= 0 {
		return nil
	}
	return value
}

func (s *adminServiceImpl) buildGroupForCreate(ctx context.Context, input *CreateGroupInput) (*Group, string, error) {
	platform := normalizeCreateGroupPlatform(input.Platform)
	subscriptionType := normalizeCreateGroupSubscriptionType(input.SubscriptionType)

	if input.FallbackGroupID != nil {
		if err := s.validateFallbackGroup(ctx, 0, *input.FallbackGroupID); err != nil {
			return nil, "", err
		}
	}

	fallbackOnInvalidRequest := normalizePositiveOptionalID(input.FallbackGroupIDOnInvalidRequest)
	if fallbackOnInvalidRequest != nil {
		if err := s.validateFallbackGroupOnInvalidRequest(ctx, 0, platform, subscriptionType, *fallbackOnInvalidRequest); err != nil {
			return nil, "", err
		}
	}

	group := &Group{
		Name:                            input.Name,
		Description:                     input.Description,
		Platform:                        platform,
		RateMultiplier:                  input.RateMultiplier,
		IsExclusive:                     input.IsExclusive,
		Status:                          StatusActive,
		SubscriptionType:                subscriptionType,
		DailyLimitUSD:                   normalizeLimit(input.DailyLimitUSD),
		WeeklyLimitUSD:                  normalizeLimit(input.WeeklyLimitUSD),
		MonthlyLimitUSD:                 normalizeLimit(input.MonthlyLimitUSD),
		ImagePrice1K:                    normalizePrice(input.ImagePrice1K),
		ImagePrice2K:                    normalizePrice(input.ImagePrice2K),
		ImagePrice4K:                    normalizePrice(input.ImagePrice4K),
		ClaudeCodeOnly:                  input.ClaudeCodeOnly,
		FallbackGroupID:                 input.FallbackGroupID,
		FallbackGroupIDOnInvalidRequest: fallbackOnInvalidRequest,
		ModelRouting:                    input.ModelRouting,
		MCPXMLInject:                    normalizeCreateGroupMCPXMLInject(input.MCPXMLInject),
		SupportedModelScopes:            input.SupportedModelScopes,
		AllowMessagesDispatch:           input.AllowMessagesDispatch,
		RequireOAuthOnly:                input.RequireOAuthOnly,
		RequirePrivacySet:               input.RequirePrivacySet,
		DefaultMappedModel:              input.DefaultMappedModel,
	}
	return group, platform, nil
}

func (s *adminServiceImpl) applyUpdateGroupInput(ctx context.Context, id int64, group *Group, input *UpdateGroupInput) error {
	if input.Name != "" {
		group.Name = input.Name
	}
	if input.Description != "" {
		group.Description = input.Description
	}
	if input.Platform != "" {
		group.Platform = input.Platform
	}
	if input.RateMultiplier != nil {
		group.RateMultiplier = *input.RateMultiplier
	}
	if input.IsExclusive != nil {
		group.IsExclusive = *input.IsExclusive
	}
	if input.Status != "" {
		group.Status = input.Status
	}
	if input.SubscriptionType != "" {
		group.SubscriptionType = input.SubscriptionType
	}

	group.DailyLimitUSD = normalizeLimit(input.DailyLimitUSD)
	group.WeeklyLimitUSD = normalizeLimit(input.WeeklyLimitUSD)
	group.MonthlyLimitUSD = normalizeLimit(input.MonthlyLimitUSD)

	if input.ImagePrice1K != nil {
		group.ImagePrice1K = normalizePrice(input.ImagePrice1K)
	}
	if input.ImagePrice2K != nil {
		group.ImagePrice2K = normalizePrice(input.ImagePrice2K)
	}
	if input.ImagePrice4K != nil {
		group.ImagePrice4K = normalizePrice(input.ImagePrice4K)
	}

	if input.ClaudeCodeOnly != nil {
		group.ClaudeCodeOnly = *input.ClaudeCodeOnly
	}
	if input.FallbackGroupID != nil {
		if *input.FallbackGroupID > 0 {
			if err := s.validateFallbackGroup(ctx, id, *input.FallbackGroupID); err != nil {
				return err
			}
			group.FallbackGroupID = input.FallbackGroupID
		} else {
			group.FallbackGroupID = nil
		}
	}

	fallbackOnInvalidRequest := group.FallbackGroupIDOnInvalidRequest
	if input.FallbackGroupIDOnInvalidRequest != nil {
		fallbackOnInvalidRequest = normalizePositiveOptionalID(input.FallbackGroupIDOnInvalidRequest)
	}
	if fallbackOnInvalidRequest != nil {
		if err := s.validateFallbackGroupOnInvalidRequest(ctx, id, group.Platform, group.SubscriptionType, *fallbackOnInvalidRequest); err != nil {
			return err
		}
	}
	group.FallbackGroupIDOnInvalidRequest = fallbackOnInvalidRequest

	if input.ModelRouting != nil {
		group.ModelRouting = input.ModelRouting
	}
	if input.ModelRoutingEnabled != nil {
		group.ModelRoutingEnabled = *input.ModelRoutingEnabled
	}
	if input.MCPXMLInject != nil {
		group.MCPXMLInject = *input.MCPXMLInject
	}
	if input.SupportedModelScopes != nil {
		group.SupportedModelScopes = *input.SupportedModelScopes
	}
	if input.AllowMessagesDispatch != nil {
		group.AllowMessagesDispatch = *input.AllowMessagesDispatch
	}
	if input.RequireOAuthOnly != nil {
		group.RequireOAuthOnly = *input.RequireOAuthOnly
	}
	if input.RequirePrivacySet != nil {
		group.RequirePrivacySet = *input.RequirePrivacySet
	}
	if input.DefaultMappedModel != nil {
		group.DefaultMappedModel = *input.DefaultMappedModel
	}

	return nil
}
