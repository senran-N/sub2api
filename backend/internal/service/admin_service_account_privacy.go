package service

import (
	"context"

	"github.com/senran-N/sub2api/internal/pkg/logger"
)

func (s *adminServiceImpl) ResetAccountQuota(ctx context.Context, id int64) error {
	return s.accountRepo.ResetQuotaUsed(ctx, id)
}

func (s *adminServiceImpl) resolveAccountProxyURL(ctx context.Context, proxyID *int64) string {
	return resolveAccountProxyURL(ctx, s.proxyRepo, proxyID)
}

func (s *adminServiceImpl) persistAccountPrivacyMode(
	ctx context.Context,
	account *Account,
	mode string,
	updateFailureLogKey string,
	applyInMemory func(*Account, string),
) string {
	return persistAccountPrivacyMode(ctx, s.accountRepo, account, mode, applyInMemory, func(err error) {
		if err != nil && updateFailureLogKey != "" {
			logger.LegacyPrintf("service.admin", "%s: account_id=%d err=%v", updateFailureLogKey, account.ID, err)
		}
	})
}

func (s *adminServiceImpl) EnsureOpenAIPrivacy(ctx context.Context, account *Account) string {
	if account.Platform != PlatformOpenAI || account.Type != AccountTypeOAuth {
		return ""
	}
	if s.privacyClientFactory == nil {
		return ""
	}
	if shouldSkipOpenAIPrivacyEnsure(account.Extra) {
		return ""
	}

	token := readAccountCredential(account, "access_token")
	if token == "" {
		return ""
	}

	mode := disableOpenAITraining(ctx, s.privacyClientFactory, token, s.resolveAccountProxyURL(ctx, account.ProxyID))
	return s.persistAccountPrivacyMode(ctx, account, mode, "", nil)
}

func (s *adminServiceImpl) ForceOpenAIPrivacy(ctx context.Context, account *Account) string {
	if account.Platform != PlatformOpenAI || account.Type != AccountTypeOAuth {
		return ""
	}
	if s.privacyClientFactory == nil {
		return ""
	}

	token := readAccountCredential(account, "access_token")
	if token == "" {
		return ""
	}

	mode := disableOpenAITraining(ctx, s.privacyClientFactory, token, s.resolveAccountProxyURL(ctx, account.ProxyID))
	return s.persistAccountPrivacyMode(ctx, account, mode, "force_update_openai_privacy_mode_failed", setAccountPrivacyMode)
}

func (s *adminServiceImpl) EnsureAntigravityPrivacy(ctx context.Context, account *Account) string {
	if account.Platform != PlatformAntigravity || account.Type != AccountTypeOAuth {
		return ""
	}
	if account.Extra != nil {
		if existing, ok := account.Extra["privacy_mode"].(string); ok && existing != "" {
			return existing
		}
	}

	token := readAccountCredential(account, "access_token")
	if token == "" {
		return ""
	}

	mode := setAntigravityPrivacy(
		ctx,
		token,
		readAccountCredential(account, "project_id"),
		s.resolveAccountProxyURL(ctx, account.ProxyID),
	)
	return s.persistAccountPrivacyMode(ctx, account, mode, "update_antigravity_privacy_mode_failed", applyAntigravityPrivacyMode)
}

func (s *adminServiceImpl) ForceAntigravityPrivacy(ctx context.Context, account *Account) string {
	if account.Platform != PlatformAntigravity || account.Type != AccountTypeOAuth {
		return ""
	}

	token := readAccountCredential(account, "access_token")
	if token == "" {
		return ""
	}

	mode := setAntigravityPrivacy(
		ctx,
		token,
		readAccountCredential(account, "project_id"),
		s.resolveAccountProxyURL(ctx, account.ProxyID),
	)
	return s.persistAccountPrivacyMode(ctx, account, mode, "force_update_antigravity_privacy_mode_failed", applyAntigravityPrivacyMode)
}
