package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	infraerrors "github.com/senran-N/sub2api/internal/pkg/errors"
	"github.com/senran-N/sub2api/internal/pkg/timezone"
)

func (s *APIKeyService) GetByKey(ctx context.Context, key string) (*APIKey, error) {
	cacheKey := s.authCacheKey(key)

	if entry, ok := s.getAuthCacheEntry(ctx, cacheKey); ok {
		if apiKey, used, err := s.applyAuthCacheEntry(key, entry); used {
			if err != nil {
				return nil, fmt.Errorf("get api key: %w", err)
			}
			s.compileAPIKeyIPRules(apiKey)
			return apiKey, nil
		}
	}

	if s.authCfg.singleflight {
		value, err, _ := s.authGroup.Do(cacheKey, func() (any, error) {
			return s.loadAuthCacheEntry(ctx, key, cacheKey)
		})
		if err != nil {
			return nil, err
		}
		entry, _ := value.(*APIKeyAuthCacheEntry)
		if apiKey, used, err := s.applyAuthCacheEntry(key, entry); used {
			if err != nil {
				return nil, fmt.Errorf("get api key: %w", err)
			}
			s.compileAPIKeyIPRules(apiKey)
			return apiKey, nil
		}
	} else {
		entry, err := s.loadAuthCacheEntry(ctx, key, cacheKey)
		if err != nil {
			return nil, err
		}
		if apiKey, used, err := s.applyAuthCacheEntry(key, entry); used {
			if err != nil {
				return nil, fmt.Errorf("get api key: %w", err)
			}
			s.compileAPIKeyIPRules(apiKey)
			return apiKey, nil
		}
	}

	apiKey, err := s.apiKeyRepo.GetByKeyForAuth(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("get api key: %w", err)
	}
	apiKey.Key = key
	s.compileAPIKeyIPRules(apiKey)
	return apiKey, nil
}

func (s *APIKeyService) ValidateKey(ctx context.Context, key string) (*APIKey, *User, error) {
	apiKey, err := s.GetByKey(ctx, key)
	if err != nil {
		return nil, nil, err
	}
	if !apiKey.IsActive() {
		return nil, nil, infraerrors.Unauthorized("API_KEY_INACTIVE", "api key is not active")
	}

	user, err := s.userRepo.GetByID(ctx, apiKey.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("get user: %w", err)
	}
	if !user.IsActive() {
		return nil, nil, ErrUserNotActive
	}

	return apiKey, user, nil
}

func (s *APIKeyService) TouchLastUsed(ctx context.Context, keyID int64) error {
	if keyID <= 0 {
		return nil
	}

	now := time.Now()
	if value, ok := s.lastUsedTouchL1.Load(keyID); ok {
		if nextAllowedAt, ok := value.(time.Time); ok && now.Before(nextAllowedAt) {
			return nil
		}
	}

	_, err, _ := s.lastUsedTouchSF.Do(strconv.FormatInt(keyID, 10), func() (any, error) {
		latest := time.Now()
		if value, ok := s.lastUsedTouchL1.Load(keyID); ok {
			if nextAllowedAt, ok := value.(time.Time); ok && latest.Before(nextAllowedAt) {
				return nil, nil
			}
		}

		if err := s.apiKeyRepo.UpdateLastUsed(ctx, keyID, latest); err != nil {
			s.lastUsedTouchL1.Store(keyID, latest.Add(apiKeyLastUsedFailBackoff))
			return nil, fmt.Errorf("touch api key last used: %w", err)
		}

		s.lastUsedTouchL1.Store(keyID, latest.Add(apiKeyLastUsedMinTouch))
		return nil, nil
	})
	return err
}

func (s *APIKeyService) IncrementUsage(ctx context.Context, keyID int64) error {
	if s.cache == nil {
		return nil
	}

	cacheKey := fmt.Sprintf("apikey:usage:%d:%s", keyID, timezone.Now().Format("2006-01-02"))
	if err := s.cache.IncrementDailyUsage(ctx, cacheKey); err != nil {
		return fmt.Errorf("increment usage: %w", err)
	}
	_ = s.cache.SetDailyUsageExpiry(ctx, cacheKey, 24*time.Hour)
	return nil
}
