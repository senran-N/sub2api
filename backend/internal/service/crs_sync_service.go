package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/domain"
	"github.com/senran-N/sub2api/internal/pkg/httpclient"
	"github.com/senran-N/sub2api/internal/util/urlvalidator"
)

type CRSSyncService struct {
	accountRepo        AccountRepository
	proxyRepo          ProxyRepository
	oauthService       *OAuthService
	openaiOAuthService *OpenAIOAuthService
	geminiOAuthService *GeminiOAuthService
	cfg                *config.Config
}

func NewCRSSyncService(
	accountRepo AccountRepository,
	proxyRepo ProxyRepository,
	oauthService *OAuthService,
	openaiOAuthService *OpenAIOAuthService,
	geminiOAuthService *GeminiOAuthService,
	cfg *config.Config,
) *CRSSyncService {
	return &CRSSyncService{
		accountRepo:        accountRepo,
		proxyRepo:          proxyRepo,
		oauthService:       oauthService,
		openaiOAuthService: openaiOAuthService,
		geminiOAuthService: geminiOAuthService,
		cfg:                cfg,
	}
}

type SyncFromCRSInput = domain.SyncFromCRSInput
type SyncFromCRSItemResult = domain.SyncFromCRSItemResult
type SyncFromCRSResult = domain.SyncFromCRSResult

type crsLoginResponse struct {
	Success  bool   `json:"success"`
	Token    string `json:"token"`
	Message  string `json:"message"`
	Error    string `json:"error"`
	Username string `json:"username"`
}

type crsExportResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Message string `json:"message"`
	Data    struct {
		ExportedAt              string                      `json:"exportedAt"`
		ClaudeAccounts          []crsClaudeAccount          `json:"claudeAccounts"`
		ClaudeConsoleAccounts   []crsConsoleAccount         `json:"claudeConsoleAccounts"`
		OpenAIOAuthAccounts     []crsOpenAIOAuthAccount     `json:"openaiOAuthAccounts"`
		OpenAIResponsesAccounts []crsOpenAIResponsesAccount `json:"openaiResponsesAccounts"`
		GeminiOAuthAccounts     []crsGeminiOAuthAccount     `json:"geminiOAuthAccounts"`
		GeminiAPIKeyAccounts    []crsGeminiAPIKeyAccount    `json:"geminiApiKeyAccounts"`
	} `json:"data"`
}

type crsProxy struct {
	Protocol string `json:"protocol"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type crsClaudeAccount struct {
	Kind        string         `json:"kind"`
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Platform    string         `json:"platform"`
	AuthType    string         `json:"authType"` // oauth/setup-token
	IsActive    bool           `json:"isActive"`
	Schedulable bool           `json:"schedulable"`
	Priority    int            `json:"priority"`
	Status      string         `json:"status"`
	Proxy       *crsProxy      `json:"proxy"`
	Credentials map[string]any `json:"credentials"`
	Extra       map[string]any `json:"extra"`
}

type crsConsoleAccount struct {
	Kind               string         `json:"kind"`
	ID                 string         `json:"id"`
	Name               string         `json:"name"`
	Description        string         `json:"description"`
	Platform           string         `json:"platform"`
	IsActive           bool           `json:"isActive"`
	Schedulable        bool           `json:"schedulable"`
	Priority           int            `json:"priority"`
	Status             string         `json:"status"`
	MaxConcurrentTasks int            `json:"maxConcurrentTasks"`
	Proxy              *crsProxy      `json:"proxy"`
	Credentials        map[string]any `json:"credentials"`
}

type crsOpenAIResponsesAccount struct {
	Kind        string         `json:"kind"`
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Platform    string         `json:"platform"`
	IsActive    bool           `json:"isActive"`
	Schedulable bool           `json:"schedulable"`
	Priority    int            `json:"priority"`
	Status      string         `json:"status"`
	Proxy       *crsProxy      `json:"proxy"`
	Credentials map[string]any `json:"credentials"`
}

type crsOpenAIOAuthAccount struct {
	Kind        string         `json:"kind"`
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Platform    string         `json:"platform"`
	AuthType    string         `json:"authType"` // oauth
	IsActive    bool           `json:"isActive"`
	Schedulable bool           `json:"schedulable"`
	Priority    int            `json:"priority"`
	Status      string         `json:"status"`
	Proxy       *crsProxy      `json:"proxy"`
	Credentials map[string]any `json:"credentials"`
	Extra       map[string]any `json:"extra"`
}

type crsGeminiOAuthAccount struct {
	Kind        string         `json:"kind"`
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Platform    string         `json:"platform"`
	AuthType    string         `json:"authType"` // oauth
	IsActive    bool           `json:"isActive"`
	Schedulable bool           `json:"schedulable"`
	Priority    int            `json:"priority"`
	Status      string         `json:"status"`
	Proxy       *crsProxy      `json:"proxy"`
	Credentials map[string]any `json:"credentials"`
	Extra       map[string]any `json:"extra"`
}

type crsGeminiAPIKeyAccount struct {
	Kind        string         `json:"kind"`
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Platform    string         `json:"platform"`
	IsActive    bool           `json:"isActive"`
	Schedulable bool           `json:"schedulable"`
	Priority    int            `json:"priority"`
	Status      string         `json:"status"`
	Proxy       *crsProxy      `json:"proxy"`
	Credentials map[string]any `json:"credentials"`
	Extra       map[string]any `json:"extra"`
}

// fetchCRSExport validates the connection parameters, authenticates with CRS,
// and returns the exported accounts. Shared by SyncFromCRS and PreviewFromCRS.
func (s *CRSSyncService) fetchCRSExport(ctx context.Context, baseURL, username, password string) (*crsExportResponse, error) {
	if s.cfg == nil {
		return nil, errors.New("config is not available")
	}
	normalizedURL := strings.TrimSpace(baseURL)
	if s.cfg.Security.URLAllowlist.Enabled {
		normalized, err := normalizeBaseURL(normalizedURL, s.cfg.Security.URLAllowlist.CRSHosts, s.cfg.Security.URLAllowlist.AllowPrivateHosts)
		if err != nil {
			return nil, err
		}
		normalizedURL = normalized
	} else {
		normalized, err := urlvalidator.ValidateURLFormat(normalizedURL, s.cfg.Security.URLAllowlist.AllowInsecureHTTP)
		if err != nil {
			return nil, fmt.Errorf("invalid base_url: %w", err)
		}
		normalizedURL = normalized
	}
	if strings.TrimSpace(username) == "" || strings.TrimSpace(password) == "" {
		return nil, errors.New("username and password are required")
	}

	client, err := httpclient.GetClient(httpclient.Options{
		Timeout:            20 * time.Second,
		ValidateResolvedIP: s.cfg.Security.URLAllowlist.Enabled,
		AllowPrivateHosts:  s.cfg.Security.URLAllowlist.AllowPrivateHosts,
	})
	if err != nil {
		return nil, fmt.Errorf("create http client failed: %w", err)
	}

	adminToken, err := crsLogin(ctx, client, normalizedURL, username, password)
	if err != nil {
		return nil, err
	}

	return crsExportAccounts(ctx, client, normalizedURL, adminToken)
}

func mergeMap(existing map[string]any, updates map[string]any) map[string]any {
	out := make(map[string]any, len(existing)+len(updates))
	for k, v := range existing {
		out[k] = v
	}
	for k, v := range updates {
		out[k] = v
	}
	return out
}

func (s *CRSSyncService) mapOrCreateProxy(ctx context.Context, enabled bool, cached *[]Proxy, src *crsProxy, defaultName string) (*int64, error) {
	if !enabled || src == nil {
		return nil, nil
	}
	protocol := strings.ToLower(strings.TrimSpace(src.Protocol))
	switch protocol {
	case "socks":
		protocol = "socks5"
	case "socks5h":
		protocol = "socks5"
	}
	host := strings.TrimSpace(src.Host)
	port := src.Port
	username := strings.TrimSpace(src.Username)
	password := strings.TrimSpace(src.Password)

	if protocol == "" || host == "" || port <= 0 {
		return nil, nil
	}
	if protocol != "http" && protocol != "https" && protocol != "socks5" {
		return nil, nil
	}

	// Find existing proxy (active only).
	for _, p := range *cached {
		if strings.EqualFold(p.Protocol, protocol) &&
			p.Host == host &&
			p.Port == port &&
			p.Username == username &&
			p.Password == password {
			id := p.ID
			return &id, nil
		}
	}

	// Create new proxy
	proxy := &Proxy{
		Name:     defaultProxyName(defaultName, protocol, host, port),
		Protocol: protocol,
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Status:   StatusActive,
	}
	if err := s.proxyRepo.Create(ctx, proxy); err != nil {
		return nil, err
	}

	*cached = append(*cached, *proxy)
	id := proxy.ID
	return &id, nil
}

func defaultProxyName(base, protocol, host string, port int) string {
	base = strings.TrimSpace(base)
	if base == "" {
		base = "crs"
	}
	return fmt.Sprintf("%s (%s://%s:%d)", base, protocol, host, port)
}

func defaultName(name, id string) string {
	if strings.TrimSpace(name) != "" {
		return strings.TrimSpace(name)
	}
	return "CRS " + id
}

func clampPriority(priority int) int {
	if priority < 1 || priority > 100 {
		return 50
	}
	return priority
}

func sanitizeCredentialsMap(input map[string]any) map[string]any {
	if input == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(input))
	for k, v := range input {
		// Avoid nil values to keep JSONB cleaner
		if v != nil {
			out[k] = v
		}
	}
	return out
}

func mapCRSStatus(isActive bool, status string) string {
	if !isActive {
		return "inactive"
	}
	if strings.EqualFold(strings.TrimSpace(status), "error") {
		return "error"
	}
	return "active"
}

func normalizeBaseURL(raw string, allowlist []string, allowPrivate bool) (string, error) {
	// 当 allowlist 为空时，不强制要求白名单（只进行基本的 URL 和 SSRF 验证）
	requireAllowlist := len(allowlist) > 0
	normalized, err := urlvalidator.ValidateHTTPSURL(raw, urlvalidator.ValidationOptions{
		AllowedHosts:     allowlist,
		RequireAllowlist: requireAllowlist,
		AllowPrivate:     allowPrivate,
	})
	if err != nil {
		return "", fmt.Errorf("invalid base_url: %w", err)
	}
	return normalized, nil
}

// cleanBaseURL removes trailing suffix from base_url in credentials
// Used for both Claude and OpenAI accounts to remove /v1
func cleanBaseURL(credentials map[string]any, suffixToRemove string) {
	if baseURL, ok := credentials["base_url"].(string); ok && baseURL != "" {
		trimmed := strings.TrimSpace(baseURL)
		if strings.HasSuffix(trimmed, suffixToRemove) {
			credentials["base_url"] = strings.TrimSuffix(trimmed, suffixToRemove)
		}
	}
}

func crsLogin(ctx context.Context, client *http.Client, baseURL, username, password string) (string, error) {
	payload := map[string]any{
		"username": username,
		"password": password,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/web/auth/login", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("crs login failed: status=%d body=%s", resp.StatusCode, string(raw))
	}

	var parsed crsLoginResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("crs login parse failed: %w", err)
	}
	if !parsed.Success || strings.TrimSpace(parsed.Token) == "" {
		msg := parsed.Message
		if msg == "" {
			msg = parsed.Error
		}
		if msg == "" {
			msg = "unknown error"
		}
		return "", errors.New("crs login failed: " + msg)
	}
	return parsed.Token, nil
}

func crsExportAccounts(ctx context.Context, client *http.Client, baseURL, adminToken string) (*crsExportResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/admin/sync/export-accounts?include_secrets=true", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("crs export failed: status=%d body=%s", resp.StatusCode, string(raw))
	}

	var parsed crsExportResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("crs export parse failed: %w", err)
	}
	if !parsed.Success {
		msg := parsed.Message
		if msg == "" {
			msg = parsed.Error
		}
		if msg == "" {
			msg = "unknown error"
		}
		return nil, errors.New("crs export failed: " + msg)
	}
	return &parsed, nil
}

// refreshOAuthToken attempts to refresh OAuth token for a synced account
// Returns updated credentials or nil if refresh failed/not applicable
func (s *CRSSyncService) refreshOAuthToken(ctx context.Context, account *Account) map[string]any {
	if account.Type != AccountTypeOAuth {
		return nil
	}

	var newCredentials map[string]any
	var err error

	switch account.Platform {
	case PlatformAnthropic:
		if s.oauthService == nil {
			return nil
		}
		tokenInfo, refreshErr := s.oauthService.RefreshAccountToken(ctx, account)
		if refreshErr != nil {
			err = refreshErr
		} else {
			// Preserve existing credentials
			newCredentials = make(map[string]any)
			for k, v := range account.Credentials {
				newCredentials[k] = v
			}
			// Update token fields
			newCredentials["access_token"] = tokenInfo.AccessToken
			newCredentials["token_type"] = tokenInfo.TokenType
			newCredentials["expires_in"] = tokenInfo.ExpiresIn
			newCredentials["expires_at"] = tokenInfo.ExpiresAt
			if tokenInfo.RefreshToken != "" {
				newCredentials["refresh_token"] = tokenInfo.RefreshToken
			}
			if tokenInfo.Scope != "" {
				newCredentials["scope"] = tokenInfo.Scope
			}
		}
	case PlatformOpenAI:
		if s.openaiOAuthService == nil {
			return nil
		}
		tokenInfo, refreshErr := s.openaiOAuthService.RefreshAccountToken(ctx, account)
		if refreshErr != nil {
			err = refreshErr
		} else {
			newCredentials = s.openaiOAuthService.BuildAccountCredentials(tokenInfo)
			// Preserve non-token settings from existing credentials
			for k, v := range account.Credentials {
				if _, exists := newCredentials[k]; !exists {
					newCredentials[k] = v
				}
			}
		}
	case PlatformGemini:
		if s.geminiOAuthService == nil {
			return nil
		}
		tokenInfo, refreshErr := s.geminiOAuthService.RefreshAccountToken(ctx, account)
		if refreshErr != nil {
			err = refreshErr
		} else {
			newCredentials = s.geminiOAuthService.BuildAccountCredentials(tokenInfo)
			for k, v := range account.Credentials {
				if _, exists := newCredentials[k]; !exists {
					newCredentials[k] = v
				}
			}
		}
	default:
		return nil
	}

	if err != nil {
		// Log but don't fail the sync - token might still be valid or refreshable later
		return nil
	}

	return newCredentials
}

// buildSelectedSet converts a slice of selected CRS account IDs to a set for O(1) lookup.
// Returns nil if ids is nil (field not sent → backward compatible: create all).
// Returns an empty map if ids is non-nil but empty (user selected none → create none).
func buildSelectedSet(ids []string) map[string]struct{} {
	if ids == nil {
		return nil
	}
	set := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return set
}

// shouldCreateAccount checks if a new CRS account should be created based on user selection.
// Returns true if selectedSet is nil (backward compatible: create all) or if crsID is in the set.
func shouldCreateAccount(crsID string, selectedSet map[string]struct{}) bool {
	if selectedSet == nil {
		return true
	}
	_, ok := selectedSet[crsID]
	return ok
}

type PreviewFromCRSResult = domain.PreviewFromCRSResult
type CRSPreviewAccount = domain.CRSPreviewAccount

// PreviewFromCRS connects to CRS, fetches all accounts, and classifies them
// as new or existing by batch-querying local crs_account_id mappings.
func (s *CRSSyncService) PreviewFromCRS(ctx context.Context, input SyncFromCRSInput) (*PreviewFromCRSResult, error) {
	exported, err := s.fetchCRSExport(ctx, input.BaseURL, input.Username, input.Password)
	if err != nil {
		return nil, err
	}

	// Batch query all existing CRS account IDs
	existingCRSIDs, err := s.accountRepo.ListCRSAccountIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list existing CRS accounts: %w", err)
	}

	result := &PreviewFromCRSResult{
		NewAccounts:      make([]CRSPreviewAccount, 0),
		ExistingAccounts: make([]CRSPreviewAccount, 0),
	}

	classify := func(crsID, kind, name, platform, accountType string) {
		preview := CRSPreviewAccount{
			CRSAccountID: crsID,
			Kind:         kind,
			Name:         defaultName(name, crsID),
			Platform:     platform,
			Type:         accountType,
		}
		if _, exists := existingCRSIDs[crsID]; exists {
			result.ExistingAccounts = append(result.ExistingAccounts, preview)
		} else {
			result.NewAccounts = append(result.NewAccounts, preview)
		}
	}

	for _, src := range exported.Data.ClaudeAccounts {
		authType := strings.TrimSpace(src.AuthType)
		if authType == "" {
			authType = AccountTypeOAuth
		}
		classify(src.ID, src.Kind, src.Name, PlatformAnthropic, authType)
	}
	for _, src := range exported.Data.ClaudeConsoleAccounts {
		classify(src.ID, src.Kind, src.Name, PlatformAnthropic, AccountTypeAPIKey)
	}
	for _, src := range exported.Data.OpenAIOAuthAccounts {
		classify(src.ID, src.Kind, src.Name, PlatformOpenAI, AccountTypeOAuth)
	}
	for _, src := range exported.Data.OpenAIResponsesAccounts {
		classify(src.ID, src.Kind, src.Name, PlatformOpenAI, AccountTypeAPIKey)
	}
	for _, src := range exported.Data.GeminiOAuthAccounts {
		classify(src.ID, src.Kind, src.Name, PlatformGemini, AccountTypeOAuth)
	}
	for _, src := range exported.Data.GeminiAPIKeyAccounts {
		classify(src.ID, src.Kind, src.Name, PlatformGemini, AccountTypeAPIKey)
	}

	return result, nil
}
