package service

import (
	"context"
	"strings"

	"github.com/senran-N/sub2api/internal/pkg/claude"
)

// isModelSupportedByAccountWithContext 根据账户平台检查模型支持（带 context）
// 对于 Antigravity 平台，会先获取映射后的最终模型名（包括 thinking 后缀）再检查支持
func (s *GatewayService) isModelSupportedByAccountWithContext(ctx context.Context, account *Account, requestedModel string) bool {
	if account == nil {
		return false
	}
	if account.Platform == PlatformAntigravity {
		if strings.TrimSpace(requestedModel) == "" {
			return true
		}

		mapped := mapAntigravityModel(account, requestedModel)
		if mapped == "" {
			return false
		}

		if enabled, ok := ThinkingEnabledFromContext(ctx); ok {
			finalModel := applyThinkingModelSuffix(mapped, enabled)
			if finalModel == mapped {
				return true
			}
			return account.IsModelSupported(finalModel)
		}
		return true
	}
	if account.Platform == PlatformOpenAI || account.Platform == PlatformGrok {
		return isCompatibleGatewayAccountModelEligible(account, requestedModel, account.Platform)
	}
	return s.isModelSupportedByAccount(account, requestedModel)
}

// isModelSupportedByAccount 根据账户平台检查模型支持（无 context，用于非 Antigravity 平台）
func (s *GatewayService) isModelSupportedByAccount(account *Account, requestedModel string) bool {
	if account == nil {
		return false
	}
	if account.Platform == PlatformAntigravity {
		if strings.TrimSpace(requestedModel) == "" {
			return true
		}
		return mapAntigravityModel(account, requestedModel) != ""
	}
	if account.IsBedrock() {
		_, ok := ResolveBedrockModelID(account, requestedModel)
		return ok
	}
	if account.Platform == PlatformOpenAI || account.Platform == PlatformGrok {
		return isCompatibleGatewayAccountModelEligible(account, requestedModel, account.Platform)
	}
	if account.Platform == PlatformAnthropic {
		if resolvedModel, source := resolveAnthropicCompatForwardModel(account, requestedModel); source != "" {
			if source == anthropicForwardModelSourceAccount {
				return strings.TrimSpace(resolvedModel) != ""
			}
			requestedModel = resolvedModel
		} else if account.Type != AccountTypeAPIKey && account.Type != AccountTypeUpstream {
			requestedModel = claude.NormalizeModelID(requestedModel)
		}
	}
	return account.IsModelSupported(requestedModel)
}
