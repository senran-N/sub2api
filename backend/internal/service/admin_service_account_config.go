package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const maxAccountLoadFactor = 10000

type accountMixedChannelChecker func(ctx context.Context, currentAccountID int64, currentAccountPlatform string, groupIDs []int64) error

type accountMutationBuilder struct {
	groupRepo           GroupRepository
	proxyRepo           ProxyRepository
	mixedChannelChecker accountMixedChannelChecker
	validateProxy       bool
}

func newAdminAccountMutationBuilder(s *adminServiceImpl) accountMutationBuilder {
	return accountMutationBuilder{
		groupRepo:           s.groupRepo,
		proxyRepo:           s.proxyRepo,
		mixedChannelChecker: s.checkMixedChannelRisk,
		validateProxy:       true,
	}
}

func (b accountMutationBuilder) resolveCreateAccountGroupIDs(ctx context.Context, input *CreateAccountInput) ([]int64, error) {
	groupIDs := input.GroupIDs
	if len(groupIDs) == 0 && !input.SkipDefaultGroupBind {
		defaultGroupName := input.Platform + "-default"
		groups, err := b.groupRepo.ListActiveByPlatform(ctx, input.Platform)
		if err == nil {
			for _, group := range groups {
				if group.Name == defaultGroupName {
					groupIDs = []int64{group.ID}
					break
				}
			}
		}
	}

	if len(groupIDs) > 0 && !input.SkipMixedChannelCheck && b.mixedChannelChecker != nil {
		if err := b.mixedChannelChecker(ctx, 0, input.Platform, groupIDs); err != nil {
			return nil, err
		}
	}

	return groupIDs, nil
}

func normalizeCreateAccountAutoPauseOnExpired(value *bool) bool {
	if value == nil {
		return true
	}
	return *value
}

func validateAccountProxyID(ctx context.Context, proxyRepo ProxyRepository, proxyID *int64) (*int64, error) {
	if proxyID == nil || *proxyID == 0 {
		return nil, nil
	}
	if proxyRepo == nil {
		return nil, errors.New("proxy repository not configured")
	}
	if _, err := proxyRepo.GetByID(ctx, *proxyID); err != nil {
		return nil, err
	}
	return proxyID, nil
}

func normalizeAccountLoadFactor(value *int) (*int, error) {
	if value == nil || *value <= 0 {
		return nil, nil
	}
	if *value > maxAccountLoadFactor {
		return nil, errors.New("load_factor must be <= 10000")
	}
	return value, nil
}

func normalizeAccountExpiresAt(unixSeconds *int64) *time.Time {
	if unixSeconds == nil || *unixSeconds <= 0 {
		return nil
	}

	expiresAt := time.Unix(*unixSeconds, 0)
	return &expiresAt
}

func validateAccountRateMultiplier(value *float64) error {
	if value != nil && *value < 0 {
		return errors.New("rate_multiplier must be >= 0")
	}
	return nil
}

func normalizeGrokSessionCredentialsForMutation(platform string, accountType string, credentials map[string]any) (map[string]any, error) {
	if NormalizeCompatibleGatewayPlatform(platform) != PlatformGrok || accountType != AccountTypeSession || len(credentials) == 0 {
		return credentials, nil
	}

	normalized := cloneAnyMap(credentials)
	rawSessionToken := strings.TrimSpace(getStringFromMaps(normalized, nil, "session_token"))
	if rawSessionToken == "" {
		return nil, errors.New("session_token is required for grok session accounts")
	}

	cookieHeader, err := ValidateGrokSessionImportToken(rawSessionToken)
	if err != nil {
		return nil, fmt.Errorf("invalid session_token: %w", err)
	}
	normalized["session_token"] = cookieHeader
	return normalized, nil
}

func (b accountMutationBuilder) buildAccountForCreate(input *CreateAccountInput) (*Account, error) {
	if err := validateAccountRateMultiplier(input.RateMultiplier); err != nil {
		return nil, err
	}

	loadFactor, err := normalizeAccountLoadFactor(input.LoadFactor)
	if err != nil {
		return nil, err
	}

	account := &Account{
		Name:               input.Name,
		Notes:              normalizeAccountNotes(input.Notes),
		Platform:           input.Platform,
		Type:               input.Type,
		Credentials:        input.Credentials,
		Extra:              input.Extra,
		ProxyID:            input.ProxyID,
		Concurrency:        input.Concurrency,
		Priority:           input.Priority,
		Status:             StatusActive,
		Schedulable:        true,
		AutoPauseOnExpired: normalizeCreateAccountAutoPauseOnExpired(input.AutoPauseOnExpired),
		RateMultiplier:     input.RateMultiplier,
		LoadFactor:         loadFactor,
	}

	if err := normalizeAccountMutationPayload(account, nil); err != nil {
		return nil, err
	}
	account.ExpiresAt = normalizeAccountExpiresAt(input.ExpiresAt)

	return account, nil
}

func normalizeAccountMutationPayload(account *Account, existingExtra map[string]any) error {
	if account == nil {
		return errors.New("account cannot be nil")
	}

	credentials, err := normalizeGrokSessionCredentialsForMutation(account.Platform, account.Type, account.Credentials)
	if err != nil {
		return err
	}
	account.Credentials = credentials
	account.Extra = normalizePlatformAccountExtra(existingExtra, account.Extra, account.Platform, account.Type)

	if account.Extra != nil {
		if err := ValidateQuotaResetConfig(account.Extra); err != nil {
			return err
		}
		ComputeQuotaResetAt(account.Extra)
	}
	return nil
}

func mergeBulkAccountCredentials(existing map[string]any, incoming map[string]any) map[string]any {
	if len(incoming) == 0 {
		return existing
	}
	merged := cloneAnyMap(existing)
	for key, value := range incoming {
		merged[key] = cloneAnyValue(value)
	}
	return merged
}

func applyAccountProxyID(account *Account, proxyID *int64) {
	if proxyID == nil {
		return
	}
	if *proxyID == 0 {
		account.ProxyID = nil
	} else {
		account.ProxyID = proxyID
	}
	account.Proxy = nil
}

func applyMutableAccountExtra(account *Account, inputExtra map[string]any, wasOveragesEnabled bool) error {
	if inputExtra == nil {
		return nil
	}

	normalizedExtra := cloneAnyMap(inputExtra)
	for _, key := range []string{"quota_used", "quota_daily_used", "quota_daily_start", "quota_weekly_used", "quota_weekly_start"} {
		if v, ok := account.Extra[key]; ok {
			normalizedExtra[key] = v
		}
	}
	account.Extra = normalizePlatformAccountExtra(account.Extra, normalizedExtra, account.Platform, account.Type)

	if account.Platform == PlatformAntigravity && wasOveragesEnabled && !account.IsOveragesEnabled() {
		delete(account.Extra, "antigravity_credits_overages")
		if rawLimits, ok := account.Extra[modelRateLimitsKey].(map[string]any); ok {
			delete(rawLimits, creditsExhaustedKey)
		}
	}
	if account.Platform == PlatformAntigravity && !wasOveragesEnabled && account.IsOveragesEnabled() {
		delete(account.Extra, modelRateLimitsKey)
		delete(account.Extra, "antigravity_credits_overages")
	}
	if err := ValidateQuotaResetConfig(account.Extra); err != nil {
		return err
	}
	ComputeQuotaResetAt(account.Extra)
	return nil
}

func (b accountMutationBuilder) applyAccountUpdateInput(ctx context.Context, account *Account, input *UpdateAccountInput, wasOveragesEnabled bool) error {
	if input.Name != "" {
		account.Name = input.Name
	}
	if input.Type != "" {
		account.Type = input.Type
	}
	if input.Notes != nil {
		account.Notes = normalizeAccountNotes(input.Notes)
	}
	if len(input.Credentials) > 0 {
		credentials, err := normalizeGrokSessionCredentialsForMutation(account.Platform, account.Type, input.Credentials)
		if err != nil {
			return err
		}
		account.Credentials = credentials
	}
	if err := applyMutableAccountExtra(account, input.Extra, wasOveragesEnabled); err != nil {
		return err
	}
	if input.ProxyID != nil {
		var proxyID *int64
		if b.validateProxy {
			validatedProxyID, err := validateAccountProxyID(ctx, b.proxyRepo, input.ProxyID)
			if err != nil {
				return err
			}
			proxyID = validatedProxyID
		} else if *input.ProxyID != 0 {
			proxyID = input.ProxyID
		}
		if proxyID == nil {
			clearProxyID := int64(0)
			applyAccountProxyID(account, &clearProxyID)
		} else {
			applyAccountProxyID(account, proxyID)
		}
	}
	if input.Concurrency != nil {
		account.Concurrency = *input.Concurrency
	}
	if input.Priority != nil {
		account.Priority = *input.Priority
	}
	if err := validateAccountRateMultiplier(input.RateMultiplier); err != nil {
		return err
	}
	if input.RateMultiplier != nil {
		account.RateMultiplier = input.RateMultiplier
	}
	if input.LoadFactor != nil {
		loadFactor, err := normalizeAccountLoadFactor(input.LoadFactor)
		if err != nil {
			return err
		}
		account.LoadFactor = loadFactor
	}
	if input.Status != "" {
		account.Status = input.Status
	}
	if input.ExpiresAt != nil {
		account.ExpiresAt = normalizeAccountExpiresAt(input.ExpiresAt)
	}
	if input.AutoPauseOnExpired != nil {
		account.AutoPauseOnExpired = *input.AutoPauseOnExpired
	}

	if input.GroupIDs != nil {
		if err := validateAccountGroupBindings(ctx, b.groupRepo, account.Type, *input.GroupIDs); err != nil {
			return err
		}
		if !input.SkipMixedChannelCheck && b.mixedChannelChecker != nil {
			if err := b.mixedChannelChecker(ctx, account.ID, account.Platform, *input.GroupIDs); err != nil {
				return err
			}
		}
	}

	return nil
}

func buildAccountBulkUpdate(input *BulkUpdateAccountsInput) (AccountBulkUpdate, error) {
	if err := validateAccountRateMultiplier(input.RateMultiplier); err != nil {
		return AccountBulkUpdate{}, err
	}

	repoUpdates := AccountBulkUpdate{
		Credentials: input.Credentials,
		Extra:       input.Extra,
	}
	if input.Name != "" {
		repoUpdates.Name = &input.Name
	}
	if input.ProxyID != nil {
		repoUpdates.ProxyID = input.ProxyID
	}
	if input.Concurrency != nil {
		repoUpdates.Concurrency = input.Concurrency
	}
	if input.Priority != nil {
		repoUpdates.Priority = input.Priority
	}
	if input.RateMultiplier != nil {
		repoUpdates.RateMultiplier = input.RateMultiplier
	}
	if input.LoadFactor != nil {
		loadFactor, err := normalizeAccountLoadFactor(input.LoadFactor)
		if err != nil {
			return AccountBulkUpdate{}, err
		}
		repoUpdates.LoadFactor = loadFactor
	}
	if input.Status != "" {
		repoUpdates.Status = &input.Status
	}
	if input.Schedulable != nil {
		repoUpdates.Schedulable = input.Schedulable
	}
	return repoUpdates, nil
}

func (b accountMutationBuilder) validateBulkAccountGroupChange(ctx context.Context, accountRepo AccountRepository, input *BulkUpdateAccountsInput) error {
	if input.GroupIDs != nil {
		if err := validateGroupIDsExist(ctx, b.groupRepo, *input.GroupIDs); err != nil {
			return err
		}
	}

	needGroupValidation := input.GroupIDs != nil
	needMixedChannelCheck := needGroupValidation && !input.SkipMixedChannelCheck
	if !needGroupValidation {
		return nil
	}

	platformByID := map[int64]string{}
	accounts, err := accountRepo.GetByIDs(ctx, input.AccountIDs)
	if err != nil {
		return err
	}
	for _, account := range accounts {
		if account != nil {
			platformByID[account.ID] = account.Platform
		}
	}
	for _, account := range accounts {
		if account != nil && account.Type == AccountTypeAPIKey {
			if err := validateAPIKeyGroupCompatibility(ctx, b.groupRepo, account.Type, *input.GroupIDs); err != nil {
				return err
			}
			break
		}
	}
	if !needMixedChannelCheck {
		return nil
	}

	for _, accountID := range input.AccountIDs {
		platform := platformByID[accountID]
		if platform == "" {
			continue
		}
		if b.mixedChannelChecker != nil {
			if err := b.mixedChannelChecker(ctx, accountID, platform, *input.GroupIDs); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateAccountIDList(accountIDs []int64) error {
	if len(accountIDs) == 0 {
		return nil
	}
	for _, accountID := range accountIDs {
		if accountID <= 0 {
			return fmt.Errorf("invalid account id: %d", accountID)
		}
	}
	return nil
}
