package service

import (
	"strings"

	"github.com/senran-N/sub2api/internal/pkg/claude"
)

const (
	anthropicForwardModelSourceAccount = "account"
	anthropicForwardModelSourcePrefix  = "prefix"
)

// resolveAnthropicCompatForwardModel keeps Anthropic-compatible model
// resolution aligned across normal forwarding, count_tokens, compat bridges, and
// admin probes.
func resolveAnthropicCompatForwardModel(account *Account, requestedModel string) (resolvedModel string, source string) {
	resolvedModel = strings.TrimSpace(requestedModel)
	if resolvedModel == "" || account == nil || account.Platform != PlatformAnthropic {
		return resolvedModel, ""
	}

	if account.Type == AccountTypeAPIKey || account.Type == AccountTypeUpstream {
		if mappedModel, matched := resolveMappedModelWithOpenAIReasoningFallback(account, resolvedModel); matched {
			mappedModel = strings.TrimSpace(mappedModel)
			if mappedModel != "" && mappedModel != resolvedModel {
				return mappedModel, anthropicForwardModelSourceAccount
			}
		}
		return resolvedModel, ""
	}

	if account.IsBedrock() {
		return resolvedModel, ""
	}

	normalizedModel := claude.NormalizeModelID(resolvedModel)
	if normalizedModel != resolvedModel {
		return normalizedModel, anthropicForwardModelSourcePrefix
	}
	return resolvedModel, ""
}
