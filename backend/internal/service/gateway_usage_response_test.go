package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyCacheTTLOverride_PromotesAggregateTokensToDefault5MWithoutReportingRewrite(t *testing.T) {
	usage := &ClaudeUsage{CacheCreationInputTokens: 42}

	changed := applyCacheTTLOverride(usage, "5m")

	require.False(t, changed)
	require.Equal(t, 42, usage.CacheCreation5mTokens)
	require.Equal(t, 0, usage.CacheCreation1hTokens)
}

func TestRewriteCacheCreationJSON_RewritesToOneHourBucket(t *testing.T) {
	usageObject := map[string]any{
		"cache_creation": map[string]any{
			"ephemeral_5m_input_tokens": float64(20),
			"ephemeral_1h_input_tokens": float64(5),
		},
	}

	changed := rewriteCacheCreationJSON(usageObject, "1h")

	require.True(t, changed)
	cacheCreation, ok := usageObject["cache_creation"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(0), cacheCreation["ephemeral_5m_input_tokens"])
	require.Equal(t, float64(25), cacheCreation["ephemeral_1h_input_tokens"])
}
