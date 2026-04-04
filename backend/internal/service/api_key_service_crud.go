package service

import (
	"context"
	"fmt"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/pagination"
)

func (s *APIKeyService) Create(ctx context.Context, userID int64, req CreateAPIKeyRequest) (*APIKey, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if err := s.validateIPAccessRules(req.IPWhitelist, req.IPBlacklist); err != nil {
		return nil, err
	}
	if err := s.ensureGroupBindingAllowed(ctx, user, req.GroupID); err != nil {
		return nil, err
	}

	key, err := s.resolveRequestedAPIKey(ctx, userID, req.CustomKey)
	if err != nil {
		return nil, err
	}

	apiKey := s.newAPIKeyForCreate(userID, key, req)
	if err := s.apiKeyRepo.Create(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("create api key: %w", err)
	}

	s.InvalidateAuthCacheByKey(ctx, apiKey.Key)
	s.compileAPIKeyIPRules(apiKey)
	return apiKey, nil
}

func (s *APIKeyService) List(ctx context.Context, userID int64, params pagination.PaginationParams, filters APIKeyListFilters) ([]APIKey, *pagination.PaginationResult, error) {
	keys, pageResult, err := s.apiKeyRepo.ListByUserID(ctx, userID, params, filters)
	if err != nil {
		return nil, nil, fmt.Errorf("list api keys: %w", err)
	}
	return keys, pageResult, nil
}

func (s *APIKeyService) VerifyOwnership(ctx context.Context, userID int64, apiKeyIDs []int64) ([]int64, error) {
	if len(apiKeyIDs) == 0 {
		return []int64{}, nil
	}

	validIDs, err := s.apiKeyRepo.VerifyOwnership(ctx, userID, apiKeyIDs)
	if err != nil {
		return nil, fmt.Errorf("verify api key ownership: %w", err)
	}
	return validIDs, nil
}

func (s *APIKeyService) GetByID(ctx context.Context, id int64) (*APIKey, error) {
	apiKey, err := s.apiKeyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get api key: %w", err)
	}
	s.compileAPIKeyIPRules(apiKey)
	return apiKey, nil
}

func (s *APIKeyService) Update(ctx context.Context, id int64, userID int64, req UpdateAPIKeyRequest) (*APIKey, error) {
	apiKey, err := s.apiKeyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get api key: %w", err)
	}
	if apiKey.UserID != userID {
		return nil, ErrInsufficientPerms
	}
	if err := s.validateIPAccessRules(req.IPWhitelist, req.IPBlacklist); err != nil {
		return nil, err
	}

	resetRateLimit, err := s.applyAPIKeyUpdate(ctx, apiKey, req)
	if err != nil {
		return nil, err
	}
	if err := s.apiKeyRepo.Update(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("update api key: %w", err)
	}

	s.InvalidateAuthCacheByKey(ctx, apiKey.Key)
	s.compileAPIKeyIPRules(apiKey)
	if resetRateLimit && s.rateLimitCacheInvalid != nil {
		_ = s.rateLimitCacheInvalid.InvalidateAPIKeyRateLimit(ctx, apiKey.ID)
	}

	return apiKey, nil
}

func (s *APIKeyService) Delete(ctx context.Context, id int64, userID int64) error {
	key, ownerID, err := s.apiKeyRepo.GetKeyAndOwnerID(ctx, id)
	if err != nil {
		return fmt.Errorf("get api key: %w", err)
	}
	if ownerID != userID {
		return ErrInsufficientPerms
	}

	if s.cache != nil {
		_ = s.cache.DeleteCreateAttemptCount(ctx, userID)
	}
	s.InvalidateAuthCacheByKey(ctx, key)

	if err := s.apiKeyRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete api key: %w", err)
	}
	s.lastUsedTouchL1.Delete(id)
	return nil
}

func (s *APIKeyService) GetAvailableGroups(ctx context.Context, userID int64) ([]Group, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	allGroups, err := s.groupRepo.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active groups: %w", err)
	}

	activeSubscriptions, err := s.userSubRepo.ListActiveByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list active subscriptions: %w", err)
	}

	subscribedGroupIDs := make(map[int64]bool, len(activeSubscriptions))
	for _, subscription := range activeSubscriptions {
		subscribedGroupIDs[subscription.GroupID] = true
	}

	availableGroups := make([]Group, 0, len(allGroups))
	for _, group := range allGroups {
		if s.canUserBindGroupInternal(user, &group, subscribedGroupIDs) {
			availableGroups = append(availableGroups, group)
		}
	}
	return availableGroups, nil
}

func (s *APIKeyService) SearchAPIKeys(ctx context.Context, userID int64, keyword string, limit int) ([]APIKey, error) {
	keys, err := s.apiKeyRepo.SearchAPIKeys(ctx, userID, keyword, limit)
	if err != nil {
		return nil, fmt.Errorf("search api keys: %w", err)
	}
	return keys, nil
}

func (s *APIKeyService) GetUserGroupRates(ctx context.Context, userID int64) (map[int64]float64, error) {
	if s.userGroupRateRepo == nil {
		return nil, nil
	}

	rates, err := s.userGroupRateRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user group rates: %w", err)
	}
	return rates, nil
}

func (s *APIKeyService) resolveRequestedAPIKey(ctx context.Context, userID int64, customKey *string) (string, error) {
	if customKey == nil || *customKey == "" {
		key, err := s.GenerateKey()
		if err != nil {
			return "", fmt.Errorf("generate key: %w", err)
		}
		return key, nil
	}

	if err := s.checkAPIKeyRateLimit(ctx, userID); err != nil {
		return "", err
	}
	if err := s.ValidateCustomKey(*customKey); err != nil {
		return "", err
	}

	exists, err := s.apiKeyRepo.ExistsByKey(ctx, *customKey)
	if err != nil {
		return "", fmt.Errorf("check key exists: %w", err)
	}
	if exists {
		s.incrementAPIKeyErrorCount(ctx, userID)
		return "", ErrAPIKeyExists
	}
	return *customKey, nil
}

func (s *APIKeyService) newAPIKeyForCreate(userID int64, key string, req CreateAPIKeyRequest) *APIKey {
	apiKey := &APIKey{
		UserID:      userID,
		Key:         key,
		Name:        req.Name,
		GroupID:     req.GroupID,
		Status:      StatusActive,
		IPWhitelist: req.IPWhitelist,
		IPBlacklist: req.IPBlacklist,
		Quota:       req.Quota,
		QuotaUsed:   0,
		RateLimit5h: req.RateLimit5h,
		RateLimit1d: req.RateLimit1d,
		RateLimit7d: req.RateLimit7d,
	}

	if req.ExpiresInDays != nil && *req.ExpiresInDays > 0 {
		expiresAt := time.Now().AddDate(0, 0, *req.ExpiresInDays)
		apiKey.ExpiresAt = &expiresAt
	}

	return apiKey
}

func (s *APIKeyService) applyAPIKeyUpdate(ctx context.Context, apiKey *APIKey, req UpdateAPIKeyRequest) (bool, error) {
	if req.Name != nil {
		apiKey.Name = *req.Name
	}
	if req.GroupID != nil {
		user, err := s.userRepo.GetByID(ctx, apiKey.UserID)
		if err != nil {
			return false, fmt.Errorf("get user: %w", err)
		}
		if err := s.ensureGroupBindingAllowed(ctx, user, req.GroupID); err != nil {
			return false, err
		}
		apiKey.GroupID = req.GroupID
	}
	if req.Status != nil {
		apiKey.Status = *req.Status
		if s.cache != nil {
			_ = s.cache.DeleteCreateAttemptCount(ctx, apiKey.UserID)
		}
	}

	s.applyAPIKeyQuotaUpdate(apiKey, req)
	s.applyAPIKeyExpirationUpdate(apiKey, req)
	apiKey.IPWhitelist = req.IPWhitelist
	apiKey.IPBlacklist = req.IPBlacklist
	return s.applyAPIKeyRateLimitUpdate(apiKey, req), nil
}

func (s *APIKeyService) applyAPIKeyQuotaUpdate(apiKey *APIKey, req UpdateAPIKeyRequest) {
	if req.Quota != nil {
		apiKey.Quota = *req.Quota
		if apiKey.Status == StatusAPIKeyQuotaExhausted && *req.Quota > apiKey.QuotaUsed {
			apiKey.Status = StatusActive
		}
	}
	if req.ResetQuota != nil && *req.ResetQuota {
		apiKey.QuotaUsed = 0
		if apiKey.Status == StatusAPIKeyQuotaExhausted {
			apiKey.Status = StatusActive
		}
	}
}

func (s *APIKeyService) applyAPIKeyExpirationUpdate(apiKey *APIKey, req UpdateAPIKeyRequest) {
	if req.ClearExpiration {
		apiKey.ExpiresAt = nil
		if apiKey.Status == StatusAPIKeyExpired {
			apiKey.Status = StatusActive
		}
		return
	}
	if req.ExpiresAt == nil {
		return
	}
	apiKey.ExpiresAt = req.ExpiresAt
	if apiKey.Status == StatusAPIKeyExpired && time.Now().Before(*req.ExpiresAt) {
		apiKey.Status = StatusActive
	}
}

func (s *APIKeyService) applyAPIKeyRateLimitUpdate(apiKey *APIKey, req UpdateAPIKeyRequest) bool {
	if req.RateLimit5h != nil {
		apiKey.RateLimit5h = *req.RateLimit5h
	}
	if req.RateLimit1d != nil {
		apiKey.RateLimit1d = *req.RateLimit1d
	}
	if req.RateLimit7d != nil {
		apiKey.RateLimit7d = *req.RateLimit7d
	}

	resetRateLimit := req.ResetRateLimitUsage != nil && *req.ResetRateLimitUsage
	if !resetRateLimit {
		return false
	}

	apiKey.Usage5h = 0
	apiKey.Usage1d = 0
	apiKey.Usage7d = 0
	apiKey.Window5hStart = nil
	apiKey.Window1dStart = nil
	apiKey.Window7dStart = nil
	return true
}
