package service

import "strings"

// mapAntigravityModel 获取映射后的模型名
// 完全依赖映射配置：账户映射（通配符）→ 默认映射兜底（DefaultAntigravityModelMapping）
// 注意：返回空字符串表示模型不被支持，调度时会过滤掉该账号
func mapAntigravityModel(account *Account, requestedModel string) string {
	if account == nil {
		return ""
	}

	mapping := account.GetModelMapping()
	if len(mapping) == 0 {
		return ""
	}

	mapped := account.GetMappedModel(requestedModel)
	if mapped != requestedModel {
		return mapped
	}

	if account.IsModelSupported(requestedModel) {
		return requestedModel
	}

	return ""
}

// getMappedModel 获取映射后的模型名
// 完全依赖映射配置：账户映射（通配符）→ 默认映射兜底
func (s *AntigravityGatewayService) getMappedModel(account *Account, requestedModel string) string {
	return mapAntigravityModel(account, requestedModel)
}

// applyThinkingModelSuffix 根据 thinking 配置调整模型名
// 当映射结果是 claude-sonnet-4-5 且请求开启了 thinking 时，改为 claude-sonnet-4-5-thinking
func applyThinkingModelSuffix(mappedModel string, thinkingEnabled bool) string {
	if !thinkingEnabled {
		return mappedModel
	}
	if mappedModel == "claude-sonnet-4-5" {
		return "claude-sonnet-4-5-thinking"
	}
	return mappedModel
}

// IsModelSupported 检查模型是否被支持
// 所有 claude- 和 gemini- 前缀的模型都能通过映射或透传支持
func (s *AntigravityGatewayService) IsModelSupported(requestedModel string) bool {
	return strings.HasPrefix(requestedModel, "claude-") ||
		strings.HasPrefix(requestedModel, "gemini-")
}
