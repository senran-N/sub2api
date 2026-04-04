package service

import (
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	BudgetRectifyBudgetTokens = 32000
	BudgetRectifyMaxTokens    = 64000
	BudgetRectifyMinMaxTokens = 32001
)

func isThinkingBudgetConstraintError(errMsg string) bool {
	message := strings.ToLower(errMsg)

	hasBudget := strings.Contains(message, "budget_tokens") || strings.Contains(message, "budget tokens")
	if !hasBudget {
		return false
	}
	if !strings.Contains(message, "thinking") {
		return false
	}
	if strings.Contains(message, ">= 1024") || strings.Contains(message, "greater than or equal to 1024") {
		return true
	}
	if strings.Contains(message, "1024") && strings.Contains(message, "input should be") {
		return true
	}
	return false
}

func RectifyThinkingBudget(body []byte) ([]byte, bool) {
	thinkingType := gjson.GetBytes(body, "thinking.type").String()
	if thinkingType == "adaptive" {
		return body, false
	}

	modified := body
	changed := false

	if thinkingType != "enabled" {
		if result, err := sjson.SetBytes(modified, "thinking.type", "enabled"); err == nil {
			modified = result
			changed = true
		}
	}

	currentBudget := gjson.GetBytes(modified, "thinking.budget_tokens").Int()
	if currentBudget != BudgetRectifyBudgetTokens {
		if result, err := sjson.SetBytes(modified, "thinking.budget_tokens", BudgetRectifyBudgetTokens); err == nil {
			modified = result
			changed = true
		}
	}

	maxTokens := gjson.GetBytes(modified, "max_tokens").Int()
	if maxTokens < int64(BudgetRectifyMinMaxTokens) {
		if result, err := sjson.SetBytes(modified, "max_tokens", BudgetRectifyMaxTokens); err == nil {
			modified = result
			changed = true
		}
	}

	return modified, changed
}
