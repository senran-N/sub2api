package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/senran-N/sub2api/internal/pkg/ip"
)

func (s *APIKeyService) compileAPIKeyIPRules(apiKey *APIKey) {
	if apiKey == nil {
		return
	}
	apiKey.CompiledIPWhitelist = ip.CompileIPRules(apiKey.IPWhitelist)
	apiKey.CompiledIPBlacklist = ip.CompileIPRules(apiKey.IPBlacklist)
}

func (s *APIKeyService) GenerateKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}

	prefix := s.cfg.Default.APIKeyPrefix
	if prefix == "" {
		prefix = "sk-"
	}

	return prefix + hex.EncodeToString(bytes), nil
}

func (s *APIKeyService) ValidateCustomKey(key string) error {
	if len(key) < 16 {
		return ErrAPIKeyTooShort
	}

	for _, char := range key {
		if (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-' {
			continue
		}
		return ErrAPIKeyInvalidChars
	}

	return nil
}

func (s *APIKeyService) checkAPIKeyRateLimit(ctx context.Context, userID int64) error {
	if s.cache == nil {
		return nil
	}

	count, err := s.cache.GetCreateAttemptCount(ctx, userID)
	if err != nil {
		return nil
	}
	if count >= apiKeyMaxErrorsPerHour {
		return ErrAPIKeyRateLimited
	}
	return nil
}

func (s *APIKeyService) incrementAPIKeyErrorCount(ctx context.Context, userID int64) {
	if s.cache == nil {
		return
	}
	_ = s.cache.IncrementCreateAttemptCount(ctx, userID)
}

func (s *APIKeyService) canUserBindGroup(ctx context.Context, user *User, group *Group) bool {
	if group.IsSubscriptionType() {
		_, err := s.userSubRepo.GetActiveByUserIDAndGroupID(ctx, user.ID, group.ID)
		return err == nil
	}
	return user.CanBindGroup(group.ID, group.IsExclusive)
}

func (s *APIKeyService) canUserBindGroupInternal(user *User, group *Group, subscribedGroupIDs map[int64]bool) bool {
	if group.IsSubscriptionType() {
		return subscribedGroupIDs[group.ID]
	}
	return user.CanBindGroup(group.ID, group.IsExclusive)
}

func (s *APIKeyService) validateIPAccessRules(ipWhitelist, ipBlacklist []string) error {
	if len(ipWhitelist) > 0 {
		if invalid := ip.ValidateIPPatterns(ipWhitelist); len(invalid) > 0 {
			return fmt.Errorf("%w: %v", ErrInvalidIPPattern, invalid)
		}
	}
	if len(ipBlacklist) > 0 {
		if invalid := ip.ValidateIPPatterns(ipBlacklist); len(invalid) > 0 {
			return fmt.Errorf("%w: %v", ErrInvalidIPPattern, invalid)
		}
	}
	return nil
}

func (s *APIKeyService) ensureGroupBindingAllowed(ctx context.Context, user *User, groupID *int64) error {
	if groupID == nil {
		return nil
	}

	group, err := s.groupRepo.GetByID(ctx, *groupID)
	if err != nil {
		return fmt.Errorf("get group: %w", err)
	}
	if !s.canUserBindGroup(ctx, user, group) {
		return ErrGroupNotAllowed
	}
	return nil
}
