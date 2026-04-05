package service

import (
	"context"

	"github.com/senran-N/sub2api/internal/pkg/logger"
)

func usageTokensFromClaudeUsage(usage ClaudeUsage) UsageTokens {
	return UsageTokens{
		InputTokens:           usage.InputTokens,
		OutputTokens:          usage.OutputTokens,
		CacheCreationTokens:   usage.CacheCreationInputTokens,
		CacheReadTokens:       usage.CacheReadInputTokens,
		CacheCreation5mTokens: usage.CacheCreation5mTokens,
		CacheCreation1hTokens: usage.CacheCreation1hTokens,
	}
}

func soraPriceConfigFromGroup(group *Group) *SoraPriceConfig {
	if group == nil {
		return nil
	}
	return &SoraPriceConfig{
		ImagePrice360:          group.SoraImagePrice360,
		ImagePrice540:          group.SoraImagePrice540,
		VideoPricePerRequest:   group.SoraVideoPricePerRequest,
		VideoPricePerRequestHD: group.SoraVideoPricePerRequestHD,
	}
}

func imagePriceConfigFromGroup(group *Group) *ImagePriceConfig {
	if group == nil {
		return nil
	}
	return &ImagePriceConfig{
		Price1K: group.ImagePrice1K,
		Price2K: group.ImagePrice2K,
		Price4K: group.ImagePrice4K,
	}
}

func (s *GatewayService) calculateGatewayUsageCost(ctx context.Context, result *ForwardResult, apiKey *APIKey, billingModel string, multiplier float64) *CostBreakdown {
	if billingModel == "" {
		billingModel = forwardResultBillingModel(result.Model, result.UpstreamModel)
	}

	switch {
	case result.MediaType == "image":
		return s.billingService.CalculateSoraImageCost(result.ImageSize, result.ImageCount, soraPriceConfigFromGroup(apiKey.Group), multiplier)
	case result.MediaType == "video":
		return s.billingService.CalculateSoraVideoCost(billingModel, soraPriceConfigFromGroup(apiKey.Group), multiplier)
	case result.MediaType == "prompt":
		return &CostBreakdown{}
	case result.ImageCount > 0:
		return s.billingService.CalculateImageCost(billingModel, result.ImageSize, result.ImageCount, imagePriceConfigFromGroup(apiKey.Group), multiplier)
	default:
		if resolved := resolveChannelPricing(ctx, s.resolver, apiKey, billingModel); resolved != nil {
			groupID := apiKey.Group.ID
			cost, err := s.billingService.CalculateCostUnified(CostInput{
				Ctx:            ctx,
				Model:          billingModel,
				GroupID:        &groupID,
				Tokens:         usageTokensFromClaudeUsage(result.Usage),
				RequestCount:   1,
				RateMultiplier: multiplier,
				Resolver:       s.resolver,
				Resolved:       resolved,
			})
			if err == nil {
				return cost
			}
			logger.LegacyPrintf("service.gateway", "Calculate unified cost failed: %v", err)
		}
		cost, err := s.billingService.CalculateCost(billingModel, usageTokensFromClaudeUsage(result.Usage), multiplier)
		if err != nil {
			logger.LegacyPrintf("service.gateway", "Calculate cost failed: %v", err)
			return &CostBreakdown{ActualCost: 0}
		}
		return cost
	}
}

func (s *GatewayService) calculateGatewayLongContextUsageCost(ctx context.Context, result *ForwardResult, apiKey *APIKey, billingModel string, multiplier float64, threshold int, longContextMultiplier float64) *CostBreakdown {
	if billingModel == "" {
		billingModel = forwardResultBillingModel(result.Model, result.UpstreamModel)
	}

	if result.ImageCount > 0 {
		return s.billingService.CalculateImageCost(billingModel, result.ImageSize, result.ImageCount, imagePriceConfigFromGroup(apiKey.Group), multiplier)
	}

	if resolved := resolveChannelPricing(ctx, s.resolver, apiKey, billingModel); resolved != nil {
		groupID := apiKey.Group.ID
		cost, err := s.billingService.CalculateCostUnified(CostInput{
			Ctx:            ctx,
			Model:          billingModel,
			GroupID:        &groupID,
			Tokens:         usageTokensFromClaudeUsage(result.Usage),
			RequestCount:   1,
			RateMultiplier: multiplier,
			Resolver:       s.resolver,
			Resolved:       resolved,
		})
		if err == nil {
			return cost
		}
		logger.LegacyPrintf("service.gateway", "Calculate unified long-context cost failed: %v", err)
	}

	cost, err := s.billingService.CalculateCostWithLongContext(
		billingModel,
		usageTokensFromClaudeUsage(result.Usage),
		multiplier,
		threshold,
		longContextMultiplier,
	)
	if err != nil {
		logger.LegacyPrintf("service.gateway", "Calculate cost failed: %v", err)
		return &CostBreakdown{ActualCost: 0}
	}
	return cost
}
