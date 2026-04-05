package service

import "context"

func enrichChannelUsageFields(
	ctx context.Context,
	channelService *ChannelService,
	groupID *int64,
	requestedModel string,
	upstreamModel string,
	fields ChannelUsageFields,
) ChannelUsageFields {
	if fields.OriginalModel == "" {
		fields.OriginalModel = requestedModel
	}
	if fields.ChannelMappedModel == "" {
		fields.ChannelMappedModel = requestedModel
	}
	if groupID == nil || *groupID <= 0 || channelService == nil || requestedModel == "" {
		return fields
	}

	if fields.ChannelID > 0 && fields.BillingModelSource != "" && fields.ModelMappingChain != "" {
		return fields
	}

	mapping := channelService.ResolveChannelMapping(ctx, *groupID, requestedModel)
	resolved := mapping.ToUsageFields(requestedModel, upstreamModel)

	if fields.ChannelID == 0 {
		fields.ChannelID = resolved.ChannelID
	}
	if fields.ChannelMappedModel == "" || fields.ChannelMappedModel == requestedModel {
		fields.ChannelMappedModel = resolved.ChannelMappedModel
	}
	if fields.BillingModelSource == "" {
		fields.BillingModelSource = resolved.BillingModelSource
	}
	if fields.ModelMappingChain == "" {
		fields.ModelMappingChain = resolved.ModelMappingChain
	}
	return fields
}

func resolveChannelPricing(
	ctx context.Context,
	resolver *ModelPricingResolver,
	apiKey *APIKey,
	billingModel string,
) *ResolvedPricing {
	if resolver == nil || apiKey == nil || apiKey.Group == nil || billingModel == "" {
		return nil
	}

	groupID := apiKey.Group.ID
	resolved := resolver.Resolve(ctx, PricingInput{
		Model:   billingModel,
		GroupID: &groupID,
	})
	if resolved == nil || resolved.Source != PricingSourceChannel {
		return nil
	}
	return resolved
}
