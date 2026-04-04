package service

import (
	"context"
	"fmt"
	"strings"
)

func (s *APIKeyService) CheckAPIKeyQuotaAndExpiry(apiKey *APIKey) error {
	if apiKey.IsExpired() {
		return ErrAPIKeyExpired
	}
	if apiKey.IsQuotaExhausted() {
		return ErrAPIKeyQuotaExhausted
	}
	return nil
}

func (s *APIKeyService) UpdateQuotaUsed(ctx context.Context, apiKeyID int64, cost float64) error {
	if cost <= 0 {
		return nil
	}

	type quotaStateReader interface {
		IncrementQuotaUsedAndGetState(ctx context.Context, id int64, amount float64) (*APIKeyQuotaUsageState, error)
	}

	if repo, ok := s.apiKeyRepo.(quotaStateReader); ok {
		state, err := repo.IncrementQuotaUsedAndGetState(ctx, apiKeyID, cost)
		if err != nil {
			return fmt.Errorf("increment quota used: %w", err)
		}
		if state != nil && state.Status == StatusAPIKeyQuotaExhausted && strings.TrimSpace(state.Key) != "" {
			s.InvalidateAuthCacheByKey(ctx, state.Key)
		}
		return nil
	}

	newQuotaUsed, err := s.apiKeyRepo.IncrementQuotaUsed(ctx, apiKeyID, cost)
	if err != nil {
		return fmt.Errorf("increment quota used: %w", err)
	}

	apiKey, err := s.apiKeyRepo.GetByID(ctx, apiKeyID)
	if err != nil {
		return nil
	}
	if apiKey.Quota > 0 && newQuotaUsed >= apiKey.Quota {
		apiKey.Status = StatusAPIKeyQuotaExhausted
		if err := s.apiKeyRepo.Update(ctx, apiKey); err != nil {
			return nil
		}
		s.InvalidateAuthCacheByKey(ctx, apiKey.Key)
	}

	return nil
}

func (s *APIKeyService) GetRateLimitData(ctx context.Context, id int64) (*APIKeyRateLimitData, error) {
	return s.apiKeyRepo.GetRateLimitData(ctx, id)
}

func (s *APIKeyService) UpdateRateLimitUsage(ctx context.Context, apiKeyID int64, cost float64) error {
	if cost <= 0 {
		return nil
	}
	return s.apiKeyRepo.IncrementRateLimitUsage(ctx, apiKeyID, cost)
}
