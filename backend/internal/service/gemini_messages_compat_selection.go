package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/pkg/ctxkey"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

func (s *GeminiMessagesCompatService) SelectAccountForModel(ctx context.Context, groupID *int64, sessionHash string, requestedModel string) (*Account, error) {
	return s.SelectAccountForModelWithExclusions(ctx, groupID, sessionHash, requestedModel, nil)
}

func (s *GeminiMessagesCompatService) SelectAccountForModelWithExclusions(ctx context.Context, groupID *int64, sessionHash string, requestedModel string, excludedIDs map[int64]struct{}) (*Account, error) {
	platform, useMixedScheduling, hasForcePlatform, err := s.resolvePlatformAndSchedulingMode(ctx, groupID)
	if err != nil {
		return nil, err
	}

	cacheKey := "gemini:" + sessionHash
	if account := s.tryStickySessionHit(ctx, groupID, sessionHash, cacheKey, requestedModel, excludedIDs, platform, useMixedScheduling); account != nil {
		return account, nil
	}

	if s.schedulerSnapshot != nil {
		selected, supported, err := s.selectBestGeminiAccountFromIndexedSnapshot(
			ctx,
			groupID,
			requestedModel,
			excludedIDs,
			platform,
			useMixedScheduling,
			hasForcePlatform,
		)
		if err != nil {
			return nil, fmt.Errorf("query accounts failed: %w", err)
		}
		if selected == nil && groupID != nil && hasForcePlatform {
			selected, supported, err = s.selectBestGeminiAccountFromIndexedSnapshot(
				ctx,
				nil,
				requestedModel,
				excludedIDs,
				platform,
				useMixedScheduling,
				hasForcePlatform,
			)
			if err != nil {
				return nil, fmt.Errorf("query accounts failed: %w", err)
			}
		}
		if selected != nil {
			if sessionHash != "" {
				_ = s.cache.SetSessionAccountID(ctx, derefGroupID(groupID), cacheKey, selected.ID, geminiStickySessionTTL)
			}
			return s.hydrateSelectedAccount(ctx, selected)
		}
		if requestedModel != "" && !supported {
			return nil, fmt.Errorf("no available Gemini accounts supporting model: %s", requestedModel)
		}
		if requestedModel == "" {
			return nil, errors.New("no available Gemini accounts")
		}
		return nil, fmt.Errorf("no available Gemini accounts supporting model: %s", requestedModel)
	}

	accounts, err := s.listSchedulableAccountsOnce(ctx, groupID, platform, hasForcePlatform)
	if err != nil {
		return nil, fmt.Errorf("query accounts failed: %w", err)
	}
	if len(accounts) == 0 && groupID != nil && hasForcePlatform {
		accounts, err = s.listSchedulableAccountsOnce(ctx, nil, platform, hasForcePlatform)
		if err != nil {
			return nil, fmt.Errorf("query accounts failed: %w", err)
		}
	}

	selected := s.selectBestGeminiAccountFromBatch(ctx, accounts, requestedModel, excludedIDs, platform, useMixedScheduling)
	if selected == nil {
		if requestedModel != "" {
			return nil, fmt.Errorf("no available Gemini accounts supporting model: %s", requestedModel)
		}
		return nil, errors.New("no available Gemini accounts")
	}

	if sessionHash != "" {
		_ = s.cache.SetSessionAccountID(ctx, derefGroupID(groupID), cacheKey, selected.ID, geminiStickySessionTTL)
	}
	return s.hydrateSelectedAccount(ctx, selected)
}

func (s *GeminiMessagesCompatService) selectBestGeminiAccountFromIndexedSnapshot(
	ctx context.Context,
	groupID *int64,
	requestedModel string,
	excludedIDs map[int64]struct{},
	platform string,
	useMixedScheduling bool,
	hasForcePlatform bool,
) (*Account, bool, error) {
	if s == nil || s.schedulerSnapshot == nil {
		return nil, false, nil
	}

	sources, err := buildRequestedModelCapabilitySources(ctx, s.schedulerSnapshot, groupID, platform, hasForcePlatform, requestedModel)
	if err != nil {
		return nil, false, err
	}
	if len(sources) == 0 {
		sources = []SchedulerCapabilityIndex{{Kind: SchedulerCapabilityIndexAll}}
	}

	pager := newSchedulerIndexedAccountPager(s.schedulerSnapshot, groupID, platform, hasForcePlatform, sources)
	if pager == nil {
		return nil, false, nil
	}

	supported := strings.TrimSpace(requestedModel) == ""
	selected, supported, err := selectBestAccountFromIndexedSnapshotPager(
		ctx,
		pager,
		snapshotPageSizeOrDefault(s.cfg),
		supported,
		func(batch []*Account) (*Account, error) {
			return s.selectBestGeminiAccountFromPointerBatch(ctx, batch, requestedModel, excludedIDs, platform, useMixedScheduling), nil
		},
		s.isBetterGeminiAccount,
	)
	if err != nil {
		return nil, supported, err
	}
	return selected, supported, nil
}

func (s *GeminiMessagesCompatService) selectBestGeminiAIStudioAccountFromIndexedSnapshot(
	ctx context.Context,
	groupID *int64,
) (*Account, error) {
	if s == nil || s.schedulerSnapshot == nil {
		return nil, nil
	}

	pager := newSchedulerIndexedAccountPager(s.schedulerSnapshot, groupID, PlatformGemini, true, []SchedulerCapabilityIndex{
		{Kind: SchedulerCapabilityIndexAll},
	})
	if pager == nil {
		return nil, nil
	}

	selected, _, err := selectBestAccountFromIndexedSnapshotPager(
		ctx,
		pager,
		snapshotPageSizeOrDefault(s.cfg),
		true,
		func(batch []*Account) (*Account, error) {
			return s.selectBestGeminiAIStudioAccountFromPointerBatch(batch), nil
		},
		s.isBetterGeminiAIStudioAccount,
	)
	if err != nil {
		return nil, err
	}
	return selected, nil
}

func (s *GeminiMessagesCompatService) resolvePlatformAndSchedulingMode(ctx context.Context, groupID *int64) (string, bool, bool, error) {
	forcePlatform, hasForcePlatform := ctx.Value(ctxkey.ForcePlatform).(string)
	if hasForcePlatform && forcePlatform != "" {
		return forcePlatform, false, true, nil
	}

	if groupID != nil {
		var group *Group
		if ctxGroup, ok := ctx.Value(ctxkey.Group).(*Group); ok && IsGroupContextValid(ctxGroup) && ctxGroup.ID == *groupID {
			group = ctxGroup
		} else {
			var err error
			group, err = s.groupRepo.GetByIDLite(ctx, *groupID)
			if err != nil {
				return "", false, false, fmt.Errorf("get group failed: %w", err)
			}
		}
		return group.Platform, group.Platform == PlatformGemini, false, nil
	}

	return PlatformGemini, true, false, nil
}

func (s *GeminiMessagesCompatService) tryStickySessionHit(
	ctx context.Context,
	groupID *int64,
	sessionHash, cacheKey, requestedModel string,
	excludedIDs map[int64]struct{},
	platform string,
	useMixedScheduling bool,
) *Account {
	if sessionHash == "" {
		return nil
	}

	accountID, err := s.cache.GetSessionAccountID(ctx, derefGroupID(groupID), cacheKey)
	if err != nil || accountID <= 0 {
		return nil
	}
	if _, excluded := excludedIDs[accountID]; excluded {
		return nil
	}

	account, err := s.getSchedulableAccount(ctx, accountID)
	if err != nil {
		return nil
	}
	if shouldClearStickySession(account, requestedModel) {
		_ = s.cache.DeleteSessionAccountID(ctx, derefGroupID(groupID), cacheKey)
		return nil
	}
	if !s.isAccountUsableForRequest(ctx, account, requestedModel, platform, useMixedScheduling) {
		return nil
	}

	_ = s.cache.RefreshSessionTTL(ctx, derefGroupID(groupID), cacheKey, geminiStickySessionTTL)
	return account
}

func (s *GeminiMessagesCompatService) isAccountUsableForRequest(
	ctx context.Context,
	account *Account,
	requestedModel, platform string,
	useMixedScheduling bool,
) bool {
	return s.isAccountUsableForRequestWithPrecheck(ctx, account, requestedModel, platform, useMixedScheduling, nil)
}

func (s *GeminiMessagesCompatService) isAccountUsableForRequestWithPrecheck(
	ctx context.Context,
	account *Account,
	requestedModel, platform string,
	useMixedScheduling bool,
	precheckResult map[int64]bool,
) bool {
	if !isGeminiSelectionAccountEligible(account) {
		return false
	}
	if !account.IsSchedulableForModelWithContext(ctx, requestedModel) {
		return false
	}
	if requestedModel != "" && !s.isModelSupportedByAccount(account, requestedModel) {
		return false
	}
	if !s.isAccountValidForPlatform(account, platform, useMixedScheduling) {
		return false
	}
	if !s.passesRateLimitPreCheckWithCache(ctx, account, requestedModel, precheckResult) {
		return false
	}
	return true
}

func (s *GeminiMessagesCompatService) isAccountValidForPlatform(account *Account, platform string, useMixedScheduling bool) bool {
	if account.Platform == platform {
		return true
	}
	return useMixedScheduling && account.Platform == PlatformAntigravity && account.IsMixedSchedulingEnabled()
}

func (s *GeminiMessagesCompatService) passesRateLimitPreCheckWithCache(ctx context.Context, account *Account, requestedModel string, precheckResult map[int64]bool) bool {
	if s.rateLimitService == nil || requestedModel == "" {
		return true
	}
	if precheckResult != nil {
		if ok, exists := precheckResult[account.ID]; exists {
			return ok
		}
	}

	ok, err := s.rateLimitService.PreCheckUsage(ctx, account, requestedModel)
	if err != nil {
		logger.LegacyPrintf("service.gemini_messages_compat", "[Gemini PreCheck] Account %d precheck error: %v", account.ID, err)
	}
	return ok
}

func (s *GeminiMessagesCompatService) selectBestGeminiAccountFromBatch(
	ctx context.Context,
	accounts []Account,
	requestedModel string,
	excludedIDs map[int64]struct{},
	platform string,
	useMixedScheduling bool,
) *Account {
	candidates := s.filterSchedulableGeminiCandidates(
		ctx,
		accounts,
		requestedModel,
		excludedIDs,
		platform,
		useMixedScheduling,
	)
	return selectBestByPriorityAndLastUsed(candidates, preferOAuthAccountTieBreaker)
}

func (s *GeminiMessagesCompatService) selectBestGeminiAccountFromPointerBatch(
	ctx context.Context,
	accounts []*Account,
	requestedModel string,
	excludedIDs map[int64]struct{},
	platform string,
	useMixedScheduling bool,
) *Account {
	candidates := s.filterSchedulableGeminiCandidatePointers(
		ctx,
		accounts,
		requestedModel,
		excludedIDs,
		platform,
		useMixedScheduling,
	)
	return selectBestByPriorityAndLastUsed(candidates, preferOAuthAccountTieBreaker)
}

func (s *GeminiMessagesCompatService) selectBestGeminiAIStudioAccountFromBatch(accounts []Account) *Account {
	var selected *Account
	for i := range accounts {
		acc := &accounts[i]
		if !isGeminiSelectionAccountEligible(acc) || acc.Platform != PlatformGemini {
			continue
		}
		if s.isBetterGeminiAIStudioAccount(acc, selected) {
			selected = acc
		}
	}
	return selected
}

func (s *GeminiMessagesCompatService) selectBestGeminiAIStudioAccountFromPointerBatch(accounts []*Account) *Account {
	var selected *Account
	for _, acc := range accounts {
		if !isGeminiSelectionAccountEligible(acc) || acc.Platform != PlatformGemini {
			continue
		}
		if s.isBetterGeminiAIStudioAccount(acc, selected) {
			selected = acc
		}
	}
	return selected
}

func (s *GeminiMessagesCompatService) buildPreCheckUsageResultMap(ctx context.Context, accounts []Account, requestedModel string) map[int64]bool {
	if s.rateLimitService == nil || requestedModel == "" || len(accounts) == 0 {
		return nil
	}

	candidates := make([]*Account, 0, len(accounts))
	for i := range accounts {
		candidates = append(candidates, &accounts[i])
	}

	result, err := s.rateLimitService.PreCheckUsageBatch(ctx, candidates, requestedModel)
	if err != nil {
		logger.LegacyPrintf("service.gemini_messages_compat", "[Gemini PreCheckBatch] failed: %v", err)
	}
	return result
}

func (s *GeminiMessagesCompatService) filterSchedulableGeminiCandidates(
	ctx context.Context,
	accounts []Account,
	requestedModel string,
	excludedIDs map[int64]struct{},
	platform string,
	useMixedScheduling bool,
) []*Account {
	precheckResult := s.buildPreCheckUsageResultMap(ctx, accounts, requestedModel)
	candidates := make([]*Account, 0, len(accounts))
	for i := range accounts {
		account := &accounts[i]
		if _, excluded := excludedIDs[account.ID]; excluded {
			continue
		}
		if !s.isAccountUsableForRequestWithPrecheck(ctx, account, requestedModel, platform, useMixedScheduling, precheckResult) {
			continue
		}
		candidates = append(candidates, account)
	}
	return candidates
}

func (s *GeminiMessagesCompatService) filterSchedulableGeminiCandidatePointers(
	ctx context.Context,
	accounts []*Account,
	requestedModel string,
	excludedIDs map[int64]struct{},
	platform string,
	useMixedScheduling bool,
) []*Account {
	precheckResult := s.buildPreCheckUsageResultMap(ctx, derefAccounts(accounts), requestedModel)
	candidates := make([]*Account, 0, len(accounts))
	for _, account := range accounts {
		if account == nil {
			continue
		}
		if _, excluded := excludedIDs[account.ID]; excluded {
			continue
		}
		if !s.isAccountUsableForRequestWithPrecheck(ctx, account, requestedModel, platform, useMixedScheduling, precheckResult) {
			continue
		}
		candidates = append(candidates, account)
	}
	return candidates
}

func preferOAuthAccountTieBreaker(candidate, current *Account) bool {
	return candidate != nil && current != nil && candidate.Type == AccountTypeOAuth && current.Type != AccountTypeOAuth
}

func (s *GeminiMessagesCompatService) isBetterGeminiAccount(candidate, current *Account) bool {
	if candidate == nil {
		return false
	}
	if current == nil {
		return true
	}
	if candidate.Priority < current.Priority {
		return true
	}
	if candidate.Priority > current.Priority {
		return false
	}

	switch {
	case candidate.LastUsedAt == nil && current.LastUsedAt != nil:
		return true
	case candidate.LastUsedAt != nil && current.LastUsedAt == nil:
		return false
	case candidate.LastUsedAt == nil && current.LastUsedAt == nil:
		return candidate.Type == AccountTypeOAuth && current.Type != AccountTypeOAuth
	default:
		return candidate.LastUsedAt.Before(*current.LastUsedAt)
	}
}

func isGeminiSelectionAccountEligible(account *Account) bool {
	if account == nil || !account.IsSchedulable() {
		return false
	}
	return oauthSelectionCredentialIssue(account) == ""
}

func (s *GeminiMessagesCompatService) isBetterGeminiAIStudioAccount(candidate, current *Account) bool {
	if candidate == nil {
		return false
	}
	if current == nil {
		return true
	}

	candidateRank := s.geminiAIStudioAccountRank(candidate)
	currentRank := s.geminiAIStudioAccountRank(current)
	if candidateRank != currentRank {
		return candidateRank < currentRank
	}
	return s.isBetterGeminiAccount(candidate, current)
}

func (s *GeminiMessagesCompatService) geminiAIStudioAccountRank(account *Account) int {
	if account == nil {
		return 999
	}
	switch account.Type {
	case AccountTypeAPIKey:
		if strings.TrimSpace(account.GetCredential("api_key")) != "" {
			return 0
		}
		return 9
	case AccountTypeOAuth:
		if strings.TrimSpace(account.GetCredential("project_id")) == "" {
			return 1
		}
		if strings.TrimSpace(account.GetCredential("oauth_type")) == "ai_studio" {
			return 2
		}
		return 3
	default:
		return 10
	}
}

func (s *GeminiMessagesCompatService) isModelSupportedByAccount(account *Account, requestedModel string) bool {
	if account == nil {
		return false
	}
	if account.Platform == PlatformAntigravity {
		if strings.TrimSpace(requestedModel) == "" {
			return true
		}
		return mapAntigravityModel(account, requestedModel) != ""
	}
	if _, matched := resolveMappedModelWithOpenAIReasoningFallback(account, requestedModel); matched {
		return true
	}
	return account.IsModelSupported(requestedModel)
}

func (s *GeminiMessagesCompatService) GetAntigravityGatewayService() *AntigravityGatewayService {
	return s.antigravityGatewayService
}

func (s *GeminiMessagesCompatService) getSchedulableAccount(ctx context.Context, accountID int64) (*Account, error) {
	if s.schedulerSnapshot != nil {
		return s.schedulerSnapshot.GetAccount(ctx, accountID)
	}
	return s.accountRepo.GetByID(ctx, accountID)
}

func (s *GeminiMessagesCompatService) hydrateSelectedAccount(ctx context.Context, account *Account) (*Account, error) {
	if account == nil || s.schedulerSnapshot == nil {
		return account, nil
	}
	hydrated, err := s.schedulerSnapshot.GetAccount(ctx, account.ID)
	if err != nil {
		return nil, err
	}
	if hydrated == nil {
		return nil, fmt.Errorf("selected gemini account %d not found during hydration", account.ID)
	}
	return hydrated, nil
}

func (s *GeminiMessagesCompatService) listSchedulableAccountsOnce(ctx context.Context, groupID *int64, platform string, hasForcePlatform bool) ([]Account, error) {
	if s.schedulerSnapshot != nil {
		accounts, _, err := s.schedulerSnapshot.ListSchedulableAccounts(ctx, groupID, platform, hasForcePlatform)
		return accounts, err
	}

	useMixedScheduling := platform == PlatformGemini && !hasForcePlatform
	queryPlatforms := []string{platform}
	if useMixedScheduling {
		queryPlatforms = []string{platform, PlatformAntigravity}
	}

	if groupID != nil {
		return s.accountRepo.ListSchedulableByGroupIDAndPlatforms(ctx, *groupID, queryPlatforms)
	}
	if s.cfg != nil && s.cfg.RunMode == config.RunModeSimple {
		return s.accountRepo.ListSchedulableByPlatforms(ctx, queryPlatforms)
	}
	return s.accountRepo.ListSchedulableUngroupedByPlatforms(ctx, queryPlatforms)
}

func (s *GeminiMessagesCompatService) HasAntigravityAccounts(ctx context.Context, groupID *int64) (bool, error) {
	accounts, err := s.listSchedulableAccountsOnce(ctx, groupID, PlatformAntigravity, false)
	if err != nil {
		return false, err
	}
	return len(accounts) > 0, nil
}

func (s *GeminiMessagesCompatService) SelectAccountForAIStudioEndpoints(ctx context.Context, groupID *int64) (*Account, error) {
	if s.schedulerSnapshot != nil {
		selected, err := s.selectBestGeminiAIStudioAccountFromIndexedSnapshot(ctx, groupID)
		if err != nil {
			return nil, fmt.Errorf("query accounts failed: %w", err)
		}
		if selected == nil {
			return nil, errors.New("no available Gemini accounts")
		}
		return s.hydrateSelectedAccount(ctx, selected)
	}

	accounts, err := s.listSchedulableAccountsOnce(ctx, groupID, PlatformGemini, true)
	if err != nil {
		return nil, fmt.Errorf("query accounts failed: %w", err)
	}
	if len(accounts) == 0 {
		return nil, errors.New("no available Gemini accounts")
	}
	selected := s.selectBestGeminiAIStudioAccountFromBatch(accounts)
	if selected == nil {
		return nil, errors.New("no available Gemini accounts")
	}
	return s.hydrateSelectedAccount(ctx, selected)
}
