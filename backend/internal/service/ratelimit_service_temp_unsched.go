package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const tempUnschedBodyMaxBytes = 64 << 10
const tempUnschedMessageMaxBytes = 2048

func (s *RateLimitService) GetTempUnschedStatus(ctx context.Context, accountID int64) (*TempUnschedState, error) {
	nowUnix := time.Now().Unix()
	if cachedState, err := s.getActiveTempUnschedFromCache(ctx, accountID, nowUnix); err != nil {
		return nil, err
	} else if cachedState != nil {
		return cachedState, nil
	}

	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account.TempUnschedulableUntil == nil || account.TempUnschedulableUntil.Unix() <= nowUnix {
		return nil, nil
	}

	state := decodeTempUnschedState(account.TempUnschedulableUntil.Unix(), account.TempUnschedulableReason)
	s.cacheTempUnschedState(ctx, accountID, state)
	return state, nil
}

func (s *RateLimitService) HandleTempUnschedulable(ctx context.Context, account *Account, statusCode int, responseBody []byte) bool {
	if account == nil || !account.ShouldHandleErrorCode(statusCode) {
		return false
	}
	return s.tryTempUnschedulable(ctx, account, statusCode, responseBody)
}

func (s *RateLimitService) tryTempUnschedulable(ctx context.Context, account *Account, statusCode int, responseBody []byte) bool {
	if account == nil || !account.IsTempUnschedulableEnabled() {
		return false
	}
	if statusCode == http.StatusUnauthorized && account.Platform != PlatformAntigravity {
		if s.shouldEscalateTempUnsched401(ctx, account, statusCode) {
			slog.Info("401_escalated_to_error", "account_id", account.ID, "reason", "previous temp-unschedulable was also 401")
			return false
		}
	}

	rules := account.GetTempUnschedulableRules()
	if len(rules) == 0 || statusCode <= 0 || len(responseBody) == 0 {
		return false
	}

	body := responseBody
	if len(body) > tempUnschedBodyMaxBytes {
		body = body[:tempUnschedBodyMaxBytes]
	}
	bodyLower := strings.ToLower(string(body))

	for idx, rule := range rules {
		if rule.ErrorCode != statusCode || len(rule.Keywords) == 0 {
			continue
		}

		matchedKeyword := matchTempUnschedKeyword(bodyLower, rule.Keywords)
		if matchedKeyword == "" {
			continue
		}
		if s.triggerTempUnschedulable(ctx, account, rule, idx, statusCode, matchedKeyword, responseBody) {
			return true
		}
	}

	return false
}

func (s *RateLimitService) shouldEscalateTempUnsched401(ctx context.Context, account *Account, statusCode int) bool {
	reason := account.TempUnschedulableReason
	if reason == "" {
		dbAccount, err := s.accountRepo.GetByID(ctx, account.ID)
		if err == nil && dbAccount != nil {
			reason = dbAccount.TempUnschedulableReason
		}
	}
	return wasTempUnschedByStatusCode(reason, statusCode)
}

func (s *RateLimitService) triggerTempUnschedulable(ctx context.Context, account *Account, rule TempUnschedulableRule, ruleIndex int, statusCode int, matchedKeyword string, responseBody []byte) bool {
	if account == nil || rule.DurationMinutes <= 0 {
		return false
	}

	now := time.Now()
	until := now.Add(time.Duration(rule.DurationMinutes) * time.Minute)
	state := &TempUnschedState{
		UntilUnix:       until.Unix(),
		TriggeredAtUnix: now.Unix(),
		StatusCode:      statusCode,
		MatchedKeyword:  matchedKeyword,
		RuleIndex:       ruleIndex,
		ErrorMessage:    truncateTempUnschedMessage(responseBody, tempUnschedMessageMaxBytes),
	}
	if !s.persistTempUnschedState(ctx, account.ID, until, state, "temp_unsched_set_failed") {
		return false
	}

	slog.Info("account_temp_unschedulable", "account_id", account.ID, "until", until, "rule_index", ruleIndex, "status_code", statusCode)
	return true
}

func (s *RateLimitService) persistTempUnschedState(ctx context.Context, accountID int64, until time.Time, state *TempUnschedState, logKey string) bool {
	reason := marshalTempUnschedState(state)
	if err := s.accountRepo.SetTempUnschedulable(ctx, accountID, until, reason); err != nil {
		slog.Warn(logKey, "account_id", accountID, "error", err)
		return false
	}
	s.cacheTempUnschedState(ctx, accountID, state)
	return true
}

func (s *RateLimitService) cacheTempUnschedState(ctx context.Context, accountID int64, state *TempUnschedState) {
	if s.tempUnschedCache == nil || state == nil {
		return
	}
	if err := s.tempUnschedCache.SetTempUnsched(ctx, accountID, state); err != nil {
		slog.Warn("temp_unsched_cache_set_failed", "account_id", accountID, "error", err)
	}
}

func (s *RateLimitService) getActiveTempUnschedFromCache(ctx context.Context, accountID int64, nowUnix int64) (*TempUnschedState, error) {
	if s.tempUnschedCache == nil {
		return nil, nil
	}
	state, err := s.tempUnschedCache.GetTempUnsched(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if state == nil || state.UntilUnix <= nowUnix {
		return nil, nil
	}
	return state, nil
}

func decodeTempUnschedState(untilUnix int64, reason string) *TempUnschedState {
	state := &TempUnschedState{UntilUnix: untilUnix}
	if reason == "" {
		return state
	}

	var parsed TempUnschedState
	if err := json.Unmarshal([]byte(reason), &parsed); err == nil {
		if parsed.UntilUnix == 0 {
			parsed.UntilUnix = untilUnix
		}
		return &parsed
	}

	state.ErrorMessage = reason
	return state
}

func marshalTempUnschedState(state *TempUnschedState) string {
	if state == nil {
		return ""
	}
	raw, err := json.Marshal(state)
	if err == nil {
		return string(raw)
	}
	return strings.TrimSpace(state.ErrorMessage)
}

func wasTempUnschedByStatusCode(reason string, statusCode int) bool {
	if statusCode <= 0 {
		return false
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return false
	}

	var state TempUnschedState
	if err := json.Unmarshal([]byte(reason), &state); err != nil {
		return false
	}
	return state.StatusCode == statusCode
}

func matchTempUnschedKeyword(bodyLower string, keywords []string) string {
	if bodyLower == "" {
		return ""
	}
	for _, keyword := range keywords {
		normalizedKeyword := strings.TrimSpace(keyword)
		if normalizedKeyword == "" {
			continue
		}
		if strings.Contains(bodyLower, strings.ToLower(normalizedKeyword)) {
			return normalizedKeyword
		}
	}
	return ""
}

func truncateTempUnschedMessage(body []byte, maxBytes int) string {
	if maxBytes <= 0 || len(body) == 0 {
		return ""
	}
	if len(body) > maxBytes {
		body = body[:maxBytes]
	}
	return strings.TrimSpace(string(body))
}
