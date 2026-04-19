package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/util/urlvalidator"
	"golang.org/x/sync/singleflight"
)

var ErrCompatibleModelDiscoveryUnsupported = errors.New("compatible upstream model discovery is not supported for this account")

type CompatibleUpstreamModel struct {
	ID          string
	Object      string
	Type        string
	DisplayName string
	Created     int64
	CreatedAt   string
	OwnedBy     string
}

type CompatibleUpstreamModelsService struct {
	accountRepo         AccountRepository
	httpUpstream        HTTPUpstream
	cfg                 *config.Config
	tlsFPProfileService *TLSFingerprintProfileService
	cache               *gocache.Cache
	cacheTTL            time.Duration
	singleflightGroup   singleflight.Group
}

func NewCompatibleUpstreamModelsService(
	accountRepo AccountRepository,
	httpUpstream HTTPUpstream,
	cfg *config.Config,
	tlsFPProfileService *TLSFingerprintProfileService,
) *CompatibleUpstreamModelsService {
	ttl := resolveModelsListCacheTTL(cfg)
	return &CompatibleUpstreamModelsService{
		accountRepo:         accountRepo,
		httpUpstream:        httpUpstream,
		cfg:                 cfg,
		tlsFPProfileService: tlsFPProfileService,
		cache:               gocache.New(ttl, time.Minute),
		cacheTTL:            ttl,
	}
}

func cloneCompatibleUpstreamModels(src []CompatibleUpstreamModel) []CompatibleUpstreamModel {
	if len(src) == 0 {
		return nil
	}
	dst := make([]CompatibleUpstreamModel, len(src))
	copy(dst, src)
	return dst
}

func compatibleModelsCacheKey(account *Account) string {
	if account == nil {
		return ""
	}
	return strings.Join([]string{
		fmt.Sprintf("%d", account.ID),
		account.UpdatedAt.UTC().Format(time.RFC3339Nano),
		strings.TrimSpace(account.Platform),
		strings.TrimSpace(account.Type),
		strings.TrimSpace(account.GetCompatibleBaseURL()),
		strings.TrimSpace(account.GetCompatibleAuthMode("")),
		strings.TrimSpace(account.GetCompatibleEndpointOverride("models")),
	}, "|")
}

func (s *CompatibleUpstreamModelsService) validateUpstreamBaseURL(raw string) (string, error) {
	if s.cfg == nil {
		return "", errors.New("config is not available")
	}
	if !s.cfg.Security.URLAllowlist.Enabled {
		return urlvalidator.ValidateURLFormat(raw, s.cfg.Security.URLAllowlist.AllowInsecureHTTP)
	}
	return urlvalidator.ValidateHTTPSURL(raw, urlvalidator.ValidationOptions{
		AllowedHosts:     s.cfg.Security.URLAllowlist.UpstreamHosts,
		RequireAllowlist: true,
		AllowPrivate:     s.cfg.Security.URLAllowlist.AllowPrivateHosts,
	})
}

func (s *CompatibleUpstreamModelsService) DiscoverAccountModels(ctx context.Context, account *Account) ([]CompatibleUpstreamModel, error) {
	if account == nil || !account.SupportsCompatibleModelDiscovery() {
		return nil, ErrCompatibleModelDiscoveryUnsupported
	}
	cacheKey := compatibleModelsCacheKey(account)
	if cacheKey != "" && s.cache != nil {
		if cached, found := s.cache.Get(cacheKey); found {
			if models, ok := cached.([]CompatibleUpstreamModel); ok {
				return cloneCompatibleUpstreamModels(models), nil
			}
		}
	}

	result, err, _ := s.singleflightGroup.Do(cacheKey, func() (any, error) {
		models, discoverErr := s.fetchAccountModels(ctx, account)
		if discoverErr != nil {
			return nil, discoverErr
		}
		if cacheKey != "" && s.cache != nil {
			s.cache.Set(cacheKey, cloneCompatibleUpstreamModels(models), s.cacheTTL)
		}
		return models, nil
	})
	if err != nil {
		return nil, err
	}
	models, _ := result.([]CompatibleUpstreamModel)
	return cloneCompatibleUpstreamModels(models), nil
}

func (s *CompatibleUpstreamModelsService) DiscoverGroupModels(ctx context.Context, groupID *int64, platform string) ([]CompatibleUpstreamModel, error) {
	if s == nil || s.accountRepo == nil {
		return nil, ErrCompatibleModelDiscoveryUnsupported
	}

	var (
		accounts []Account
		err      error
	)
	if groupID != nil {
		accounts, err = s.accountRepo.ListSchedulableByGroupID(ctx, *groupID)
	} else {
		accounts, err = s.accountRepo.ListSchedulable(ctx)
	}
	if err != nil {
		return nil, err
	}

	modelSet := make(map[string]CompatibleUpstreamModel)
	attempted := 0
	succeeded := 0
	var lastErr error

	for i := range accounts {
		account := &accounts[i]
		if platform != "" && account.Platform != platform {
			continue
		}
		if !account.SupportsCompatibleModelDiscovery() {
			continue
		}

		attempted++
		models, discoverErr := s.DiscoverAccountModels(ctx, account)
		if discoverErr != nil {
			lastErr = discoverErr
			continue
		}
		succeeded++
		for _, model := range models {
			if strings.TrimSpace(model.ID) == "" {
				continue
			}
			if _, exists := modelSet[model.ID]; !exists {
				modelSet[model.ID] = model
			}
		}
	}

	if succeeded == 0 {
		if attempted == 0 {
			return nil, ErrCompatibleModelDiscoveryUnsupported
		}
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, nil
	}

	models := make([]CompatibleUpstreamModel, 0, len(modelSet))
	for _, model := range modelSet {
		models = append(models, model)
	}
	sort.Slice(models, func(i, j int) bool { return models[i].ID < models[j].ID })
	return models, nil
}

func (s *CompatibleUpstreamModelsService) fetchAccountModels(ctx context.Context, account *Account) ([]CompatibleUpstreamModel, error) {
	baseURL, err := s.validateUpstreamBaseURL(account.GetCompatibleBaseURL())
	if err != nil {
		return nil, err
	}

	modelsURL := compatibleModelsEndpointURL(account, baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, modelsURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if account.Platform == PlatformAnthropic && req.Header.Get("anthropic-version") == "" {
		req.Header.Set("anthropic-version", "2023-06-01")
	}
	applyCompatibleAuthHeaders(req.Header, account.GetCompatibleAPIKey(), account.GetCompatibleAuthMode(""))

	resp, err := s.httpUpstream.DoWithTLS(
		req,
		accountTestProxyURL(account),
		account.ID,
		account.Concurrency,
		s.tlsFPProfileService.ResolveTLSProfile(account),
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("upstream models request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	models, err := parseCompatibleUpstreamModels(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse upstream models payload: %w", err)
	}
	return models, nil
}

func compatibleModelsEndpointURL(account *Account, baseURL string) string {
	if account != nil {
		if override := account.GetCompatibleEndpointOverride("models"); override != "" {
			return resolveCompatibleEndpointURL(baseURL, "/v1/models", override)
		}
	}
	if account != nil && account.Platform == PlatformOpenAI {
		return buildCompatibleModelsURL(baseURL)
	}
	return resolveCompatibleEndpointURL(baseURL, "/v1/models", "")
}

func parseCompatibleUpstreamModels(body []byte) ([]CompatibleUpstreamModel, error) {
	type payload struct {
		Data   []map[string]any `json:"data"`
		Models []map[string]any `json:"models"`
	}

	trimmedBody := bytes.TrimSpace(body)
	if len(trimmedBody) == 0 {
		return nil, errors.New("empty models payload")
	}

	var envelope payload
	if err := json.Unmarshal(trimmedBody, &envelope); err == nil {
		if envelope.Data != nil {
			return normalizeCompatibleUpstreamModels(envelope.Data), nil
		}
		if envelope.Models != nil {
			return normalizeCompatibleUpstreamModels(envelope.Models), nil
		}
	}

	var bare []map[string]any
	if err := json.Unmarshal(trimmedBody, &bare); err == nil && bare != nil {
		return normalizeCompatibleUpstreamModels(bare), nil
	}

	return nil, errors.New("unsupported models payload shape")
}

func normalizeCompatibleUpstreamModels(items []map[string]any) []CompatibleUpstreamModel {
	if len(items) == 0 {
		return nil
	}
	models := make([]CompatibleUpstreamModel, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		model := CompatibleUpstreamModel{
			ID:          strings.TrimSpace(firstStringField(item, "id", "name")),
			Object:      strings.TrimSpace(firstStringField(item, "object")),
			Type:        strings.TrimSpace(firstStringField(item, "type")),
			DisplayName: strings.TrimSpace(firstStringField(item, "display_name", "displayName")),
			CreatedAt:   strings.TrimSpace(firstStringField(item, "created_at", "createdAt")),
			OwnedBy:     strings.TrimSpace(firstStringField(item, "owned_by", "ownedBy")),
		}
		model.ID = strings.TrimPrefix(model.ID, "models/")
		if model.ID == "" {
			continue
		}
		if model.DisplayName == "" {
			model.DisplayName = model.ID
		}
		if model.Object == "" {
			model.Object = "model"
		}
		if model.Type == "" {
			model.Type = "model"
		}
		model.Created = firstInt64Field(item, "created")
		if _, exists := seen[model.ID]; exists {
			continue
		}
		seen[model.ID] = struct{}{}
		models = append(models, model)
	}
	sort.Slice(models, func(i, j int) bool { return models[i].ID < models[j].ID })
	return models
}

func firstStringField(item map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := item[key]
		if !ok || value == nil {
			continue
		}
		if text, ok := value.(string); ok {
			return text
		}
	}
	return ""
}

func firstInt64Field(item map[string]any, keys ...string) int64 {
	for _, key := range keys {
		value, ok := item[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case float64:
			return int64(typed)
		case int64:
			return typed
		case int:
			return int64(typed)
		}
	}
	return 0
}
