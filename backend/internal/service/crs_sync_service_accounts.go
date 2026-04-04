package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type crsSyncState struct {
	input       SyncFromCRSInput
	now         string
	result      *SyncFromCRSResult
	selectedSet map[string]struct{}
	proxies     []Proxy
}

type crsAccountSyncSpec struct {
	item         SyncFromCRSItemResult
	desired      *Account
	refreshOAuth bool
}

type crsSyncBuildError struct {
	action  string
	message string
}

func (e *crsSyncBuildError) Error() string {
	return e.message
}

func newCRSSyncBuildFailed(message string) error {
	return &crsSyncBuildError{action: "failed", message: message}
}

func newCRSSyncBuildSkipped(message string) error {
	return &crsSyncBuildError{action: "skipped", message: message}
}

func (s *CRSSyncService) SyncFromCRS(ctx context.Context, input SyncFromCRSInput) (*SyncFromCRSResult, error) {
	exported, err := s.fetchCRSExport(ctx, input.BaseURL, input.Username, input.Password)
	if err != nil {
		return nil, err
	}

	state := &crsSyncState{
		input:       input,
		now:         time.Now().UTC().Format(time.RFC3339),
		result:      newCRSSyncResult(exported),
		selectedSet: buildSelectedSet(input.SelectedAccountIDs),
	}
	if input.SyncProxies {
		state.proxies, _ = s.proxyRepo.ListActive(ctx)
	}

	s.syncClaudeAccounts(ctx, state, exported.Data.ClaudeAccounts)
	s.syncClaudeConsoleAccounts(ctx, state, exported.Data.ClaudeConsoleAccounts)
	s.syncOpenAIOAuthAccounts(ctx, state, exported.Data.OpenAIOAuthAccounts)
	s.syncOpenAIResponsesAccounts(ctx, state, exported.Data.OpenAIResponsesAccounts)
	s.syncGeminiOAuthAccounts(ctx, state, exported.Data.GeminiOAuthAccounts)
	s.syncGeminiAPIKeyAccounts(ctx, state, exported.Data.GeminiAPIKeyAccounts)

	return state.result, nil
}

func newCRSSyncResult(exported *crsExportResponse) *SyncFromCRSResult {
	return &SyncFromCRSResult{
		Items: make(
			[]SyncFromCRSItemResult,
			0,
			len(exported.Data.ClaudeAccounts)+len(exported.Data.ClaudeConsoleAccounts)+len(exported.Data.OpenAIOAuthAccounts)+len(exported.Data.OpenAIResponsesAccounts)+len(exported.Data.GeminiOAuthAccounts)+len(exported.Data.GeminiAPIKeyAccounts),
		),
	}
}

func (s *CRSSyncService) syncClaudeAccounts(ctx context.Context, state *crsSyncState, accounts []crsClaudeAccount) {
	for _, src := range accounts {
		spec, err := s.buildClaudeAccountSyncSpec(ctx, state, src)
		s.applyCRSAccountSyncSpec(ctx, state, spec, err)
	}
}

func (s *CRSSyncService) syncClaudeConsoleAccounts(ctx context.Context, state *crsSyncState, accounts []crsConsoleAccount) {
	for _, src := range accounts {
		spec, err := s.buildClaudeConsoleSyncSpec(ctx, state, src)
		s.applyCRSAccountSyncSpec(ctx, state, spec, err)
	}
}

func (s *CRSSyncService) syncOpenAIOAuthAccounts(ctx context.Context, state *crsSyncState, accounts []crsOpenAIOAuthAccount) {
	for _, src := range accounts {
		spec, err := s.buildOpenAIOAuthSyncSpec(ctx, state, src)
		s.applyCRSAccountSyncSpec(ctx, state, spec, err)
	}
}

func (s *CRSSyncService) syncOpenAIResponsesAccounts(ctx context.Context, state *crsSyncState, accounts []crsOpenAIResponsesAccount) {
	for _, src := range accounts {
		spec, err := s.buildOpenAIResponsesSyncSpec(ctx, state, src)
		s.applyCRSAccountSyncSpec(ctx, state, spec, err)
	}
}

func (s *CRSSyncService) syncGeminiOAuthAccounts(ctx context.Context, state *crsSyncState, accounts []crsGeminiOAuthAccount) {
	for _, src := range accounts {
		spec, err := s.buildGeminiOAuthSyncSpec(ctx, state, src)
		s.applyCRSAccountSyncSpec(ctx, state, spec, err)
	}
}

func (s *CRSSyncService) syncGeminiAPIKeyAccounts(ctx context.Context, state *crsSyncState, accounts []crsGeminiAPIKeyAccount) {
	for _, src := range accounts {
		spec, err := s.buildGeminiAPIKeySyncSpec(ctx, state, src)
		s.applyCRSAccountSyncSpec(ctx, state, spec, err)
	}
}

func (s *CRSSyncService) applyCRSAccountSyncSpec(ctx context.Context, state *crsSyncState, spec *crsAccountSyncSpec, buildErr error) {
	if spec == nil {
		return
	}
	if buildErr != nil {
		state.recordBuildError(spec.item, buildErr)
		return
	}

	existing, err := s.accountRepo.GetByCRSAccountID(ctx, spec.item.CRSAccountID)
	if err != nil {
		state.recordFailure(spec.item, "db lookup failed: "+err.Error())
		return
	}

	if existing == nil {
		if !shouldCreateAccount(spec.item.CRSAccountID, state.selectedSet) {
			state.recordSkipped(spec.item, "not selected")
			return
		}
		if err := s.accountRepo.Create(ctx, spec.desired); err != nil {
			state.recordFailure(spec.item, "create failed: "+err.Error())
			return
		}
		s.refreshSyncedOAuthCredentials(ctx, spec.desired, spec.refreshOAuth)
		state.recordCreated(spec.item)
		return
	}

	applyCRSDesiredAccount(existing, spec.desired)
	if err := s.accountRepo.Update(ctx, existing); err != nil {
		state.recordFailure(spec.item, "update failed: "+err.Error())
		return
	}
	s.refreshSyncedOAuthCredentials(ctx, existing, spec.refreshOAuth)
	state.recordUpdated(spec.item)
}

func applyCRSDesiredAccount(existing *Account, desired *Account) {
	existing.Extra = mergeMap(existing.Extra, desired.Extra)
	existing.Name = desired.Name
	existing.Platform = desired.Platform
	existing.Type = desired.Type
	existing.Credentials = mergeMap(existing.Credentials, desired.Credentials)
	if desired.ProxyID != nil {
		existing.ProxyID = desired.ProxyID
	}
	existing.Concurrency = desired.Concurrency
	existing.Priority = desired.Priority
	existing.Status = desired.Status
	existing.Schedulable = desired.Schedulable
}

func (s *CRSSyncService) refreshSyncedOAuthCredentials(ctx context.Context, account *Account, enabled bool) {
	if !enabled {
		return
	}
	if refreshedCreds := s.refreshOAuthToken(ctx, account); refreshedCreds != nil {
		_ = persistAccountCredentials(ctx, s.accountRepo, account, refreshedCreds)
	}
}

func (s *CRSSyncService) buildClaudeAccountSyncSpec(ctx context.Context, state *crsSyncState, src crsClaudeAccount) (*crsAccountSyncSpec, error) {
	spec := &crsAccountSyncSpec{
		item: newCRSSyncItem(src.ID, src.Kind, src.Name),
	}

	targetType := strings.TrimSpace(src.AuthType)
	if targetType == "" {
		targetType = AccountTypeOAuth
	}
	if targetType != AccountTypeOAuth && targetType != AccountTypeSetupToken {
		return spec, newCRSSyncBuildSkipped("unsupported authType: " + targetType)
	}

	accessToken, _ := src.Credentials["access_token"].(string)
	if strings.TrimSpace(accessToken) == "" {
		return spec, newCRSSyncBuildFailed("missing access_token")
	}

	proxyID, err := s.syncCRSProxy(ctx, state, src.Proxy, src.Name)
	if err != nil {
		return spec, newCRSSyncBuildFailed("proxy sync failed: " + err.Error())
	}

	credentials := sanitizeCredentialsMap(src.Credentials)
	cleanBaseURL(credentials, "/v1")
	convertRFC3339CredentialToUnixValue(credentials, "expires_at", false)
	if _, exists := credentials["intercept_warmup_requests"]; !exists {
		credentials["intercept_warmup_requests"] = false
	}

	extra := buildCRSSyncedExtra(src.Extra, src.ID, src.Kind, state.now)
	if orgUUID, ok := src.Credentials["org_uuid"]; ok {
		extra["org_uuid"] = orgUUID
	}
	if accountUUID, ok := src.Credentials["account_uuid"]; ok {
		extra["account_uuid"] = accountUUID
	}

	spec.desired = buildCRSAccount(defaultName(src.Name, src.ID), PlatformAnthropic, targetType, credentials, extra, proxyID, 3, clampPriority(src.Priority), mapCRSStatus(src.IsActive, src.Status), src.Schedulable)
	spec.refreshOAuth = targetType == AccountTypeOAuth
	return spec, nil
}

func (s *CRSSyncService) buildClaudeConsoleSyncSpec(ctx context.Context, state *crsSyncState, src crsConsoleAccount) (*crsAccountSyncSpec, error) {
	spec := &crsAccountSyncSpec{
		item: newCRSSyncItem(src.ID, src.Kind, src.Name),
	}

	apiKey, _ := src.Credentials["api_key"].(string)
	if strings.TrimSpace(apiKey) == "" {
		return spec, newCRSSyncBuildFailed("missing api_key")
	}

	proxyID, err := s.syncCRSProxy(ctx, state, src.Proxy, src.Name)
	if err != nil {
		return spec, newCRSSyncBuildFailed("proxy sync failed: " + err.Error())
	}

	concurrency := 3
	if src.MaxConcurrentTasks > 0 {
		concurrency = src.MaxConcurrentTasks
	}
	spec.desired = buildCRSAccount(
		defaultName(src.Name, src.ID),
		PlatformAnthropic,
		AccountTypeAPIKey,
		sanitizeCredentialsMap(src.Credentials),
		buildCRSSyncMeta(src.ID, src.Kind, state.now),
		proxyID,
		concurrency,
		clampPriority(src.Priority),
		mapCRSStatus(src.IsActive, src.Status),
		src.Schedulable,
	)
	return spec, nil
}

func (s *CRSSyncService) buildOpenAIOAuthSyncSpec(ctx context.Context, state *crsSyncState, src crsOpenAIOAuthAccount) (*crsAccountSyncSpec, error) {
	spec := &crsAccountSyncSpec{
		item: newCRSSyncItem(src.ID, src.Kind, src.Name),
	}

	accessToken, _ := src.Credentials["access_token"].(string)
	if strings.TrimSpace(accessToken) == "" {
		return spec, newCRSSyncBuildFailed("missing access_token")
	}

	proxyID, err := s.syncCRSProxy(ctx, state, src.Proxy, src.Name)
	if err != nil {
		return spec, newCRSSyncBuildFailed("proxy sync failed: " + err.Error())
	}

	credentials := sanitizeCredentialsMap(src.Credentials)
	if tokenType, ok := credentials["token_type"].(string); !ok || strings.TrimSpace(tokenType) == "" {
		credentials["token_type"] = "Bearer"
	}
	convertRFC3339CredentialToUnixValue(credentials, "expires_at", false)

	extra := buildCRSSyncedExtra(src.Extra, src.ID, src.Kind, state.now)
	if crsEmail, ok := src.Extra["crs_email"]; ok {
		extra["email"] = crsEmail
	}

	spec.desired = buildCRSAccount(defaultName(src.Name, src.ID), PlatformOpenAI, AccountTypeOAuth, credentials, extra, proxyID, 3, clampPriority(src.Priority), mapCRSStatus(src.IsActive, src.Status), src.Schedulable)
	spec.refreshOAuth = true
	return spec, nil
}

func (s *CRSSyncService) buildOpenAIResponsesSyncSpec(ctx context.Context, state *crsSyncState, src crsOpenAIResponsesAccount) (*crsAccountSyncSpec, error) {
	spec := &crsAccountSyncSpec{
		item: newCRSSyncItem(src.ID, src.Kind, src.Name),
	}

	apiKey, _ := src.Credentials["api_key"].(string)
	if strings.TrimSpace(apiKey) == "" {
		return spec, newCRSSyncBuildFailed("missing api_key")
	}

	credentials := sanitizeCredentialsMap(src.Credentials)
	if baseURL, ok := credentials["base_url"].(string); !ok || strings.TrimSpace(baseURL) == "" {
		credentials["base_url"] = "https://api.openai.com"
	}
	cleanBaseURL(credentials, "/v1")

	proxyID, err := s.syncCRSProxy(ctx, state, src.Proxy, src.Name)
	if err != nil {
		return spec, newCRSSyncBuildFailed("proxy sync failed: " + err.Error())
	}

	spec.desired = buildCRSAccount(
		defaultName(src.Name, src.ID),
		PlatformOpenAI,
		AccountTypeAPIKey,
		credentials,
		buildCRSSyncMeta(src.ID, src.Kind, state.now),
		proxyID,
		3,
		clampPriority(src.Priority),
		mapCRSStatus(src.IsActive, src.Status),
		src.Schedulable,
	)
	return spec, nil
}

func (s *CRSSyncService) buildGeminiOAuthSyncSpec(ctx context.Context, state *crsSyncState, src crsGeminiOAuthAccount) (*crsAccountSyncSpec, error) {
	spec := &crsAccountSyncSpec{
		item: newCRSSyncItem(src.ID, src.Kind, src.Name),
	}

	refreshToken, _ := src.Credentials["refresh_token"].(string)
	if strings.TrimSpace(refreshToken) == "" {
		return spec, newCRSSyncBuildFailed("missing refresh_token")
	}

	proxyID, err := s.syncCRSProxy(ctx, state, src.Proxy, src.Name)
	if err != nil {
		return spec, newCRSSyncBuildFailed("proxy sync failed: " + err.Error())
	}

	credentials := sanitizeCredentialsMap(src.Credentials)
	if tokenType, ok := credentials["token_type"].(string); !ok || strings.TrimSpace(tokenType) == "" {
		credentials["token_type"] = "Bearer"
	}
	convertRFC3339CredentialToUnixValue(credentials, "expires_at", true)

	spec.desired = buildCRSAccount(
		defaultName(src.Name, src.ID),
		PlatformGemini,
		AccountTypeOAuth,
		credentials,
		buildCRSSyncedExtra(src.Extra, src.ID, src.Kind, state.now),
		proxyID,
		3,
		clampPriority(src.Priority),
		mapCRSStatus(src.IsActive, src.Status),
		src.Schedulable,
	)
	spec.refreshOAuth = true
	return spec, nil
}

func (s *CRSSyncService) buildGeminiAPIKeySyncSpec(ctx context.Context, state *crsSyncState, src crsGeminiAPIKeyAccount) (*crsAccountSyncSpec, error) {
	spec := &crsAccountSyncSpec{
		item: newCRSSyncItem(src.ID, src.Kind, src.Name),
	}

	apiKey, _ := src.Credentials["api_key"].(string)
	if strings.TrimSpace(apiKey) == "" {
		return spec, newCRSSyncBuildFailed("missing api_key")
	}

	proxyID, err := s.syncCRSProxy(ctx, state, src.Proxy, src.Name)
	if err != nil {
		return spec, newCRSSyncBuildFailed("proxy sync failed: " + err.Error())
	}

	credentials := sanitizeCredentialsMap(src.Credentials)
	if baseURL, ok := credentials["base_url"].(string); !ok || strings.TrimSpace(baseURL) == "" {
		credentials["base_url"] = "https://generativelanguage.googleapis.com"
	}

	spec.desired = buildCRSAccount(
		defaultName(src.Name, src.ID),
		PlatformGemini,
		AccountTypeAPIKey,
		credentials,
		buildCRSSyncedExtra(src.Extra, src.ID, src.Kind, state.now),
		proxyID,
		3,
		clampPriority(src.Priority),
		mapCRSStatus(src.IsActive, src.Status),
		src.Schedulable,
	)
	return spec, nil
}

func (s *CRSSyncService) syncCRSProxy(ctx context.Context, state *crsSyncState, src *crsProxy, accountName string) (*int64, error) {
	return s.mapOrCreateProxy(ctx, state.input.SyncProxies, &state.proxies, src, fmt.Sprintf("crs-%s", accountName))
}

func newCRSSyncItem(crsAccountID, kind, name string) SyncFromCRSItemResult {
	return SyncFromCRSItemResult{
		CRSAccountID: crsAccountID,
		Kind:         kind,
		Name:         name,
	}
}

func buildCRSAccount(name, platform, accountType string, credentials, extra map[string]any, proxyID *int64, concurrency, priority int, status string, schedulable bool) *Account {
	return &Account{
		Name:        name,
		Platform:    platform,
		Type:        accountType,
		Credentials: credentials,
		Extra:       extra,
		ProxyID:     proxyID,
		Concurrency: concurrency,
		Priority:    priority,
		Status:      status,
		Schedulable: schedulable,
	}
}

func buildCRSSyncMeta(crsAccountID, kind, syncedAt string) map[string]any {
	return map[string]any{
		"crs_account_id": crsAccountID,
		"crs_kind":       kind,
		"crs_synced_at":  syncedAt,
	}
}

func buildCRSSyncedExtra(existingExtra map[string]any, crsAccountID, kind, syncedAt string) map[string]any {
	return mergeMap(existingExtra, buildCRSSyncMeta(crsAccountID, kind, syncedAt))
}

func convertRFC3339CredentialToUnixValue(credentials map[string]any, key string, asString bool) {
	value, ok := credentials[key].(string)
	if !ok || strings.TrimSpace(value) == "" {
		return
	}
	parsedAt, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return
	}
	if asString {
		credentials[key] = fmt.Sprintf("%d", parsedAt.Unix())
		return
	}
	credentials[key] = parsedAt.Unix()
}

func (s *crsSyncState) recordBuildError(item SyncFromCRSItemResult, err error) {
	var buildErr *crsSyncBuildError
	if errors.As(err, &buildErr) {
		s.recordItem(item, buildErr.action, buildErr.message)
		return
	}
	s.recordFailure(item, err.Error())
}

func (s *crsSyncState) recordCreated(item SyncFromCRSItemResult) {
	s.recordItem(item, "created", "")
}

func (s *crsSyncState) recordUpdated(item SyncFromCRSItemResult) {
	s.recordItem(item, "updated", "")
}

func (s *crsSyncState) recordSkipped(item SyncFromCRSItemResult, message string) {
	s.recordItem(item, "skipped", message)
}

func (s *crsSyncState) recordFailure(item SyncFromCRSItemResult, message string) {
	s.recordItem(item, "failed", message)
}

func (s *crsSyncState) recordItem(item SyncFromCRSItemResult, action, message string) {
	item.Action = action
	item.Error = message
	switch action {
	case "created":
		s.result.Created++
	case "updated":
		s.result.Updated++
	case "skipped":
		s.result.Skipped++
	case "failed":
		s.result.Failed++
	}
	s.result.Items = append(s.result.Items, item)
}
