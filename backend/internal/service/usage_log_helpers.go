package service

import "strings"

func optionalTrimmedStringPtr(raw string) *string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func optionalInt64Ptr(value int64) *int64 {
	if value <= 0 {
		return nil
	}
	return &value
}

// optionalNonEqualStringPtr returns a pointer to value if it is non-empty and
// differs from compare; otherwise nil. Used to store upstream_model only when
// it differs from the requested model.
func optionalNonEqualStringPtr(value, compare string) *string {
	if value == "" || value == compare {
		return nil
	}
	return &value
}

func forwardResultBillingModel(requestedModel, upstreamModel string) string {
	if trimmed := strings.TrimSpace(requestedModel); trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(upstreamModel)
}

func resolveChannelBillingModel(fields ChannelUsageFields, fallbackBillingModel string) string {
	switch fields.BillingModelSource {
	case BillingModelSourceRequested:
		if model := strings.TrimSpace(fields.OriginalModel); model != "" {
			return model
		}
	case BillingModelSourceChannelMapped:
		if model := strings.TrimSpace(fields.ChannelMappedModel); model != "" {
			return model
		}
	case BillingModelSourceUpstream:
		if model := strings.TrimSpace(fallbackBillingModel); model != "" {
			return model
		}
	}
	return strings.TrimSpace(fallbackBillingModel)
}
