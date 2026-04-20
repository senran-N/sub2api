package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/senran-N/sub2api/internal/pkg/logger"
	"golang.org/x/sync/singleflight"
)

const (
	grokMediaAssetRoutePrefix       = "/grok/media/assets"
	grokMediaAssetStatusReady       = "ready"
	defaultGrokMediaAssetRetention  = 72 * time.Hour
	defaultGrokMediaAssetGCInterval = 30 * time.Minute
	defaultGrokMediaAssetGCBatch    = 128
)

type GrokMediaAssetService struct {
	gatewayService *GatewayService
	repo           GrokMediaAssetRepository
	cacheRoot      string
	downloads      singleflight.Group
	retention      time.Duration
	gcInterval     time.Duration
	gcBatch        int
	now            func() time.Time
	lastCleanupAt  atomic.Int64
}

func NewGrokMediaAssetService(gatewayService *GatewayService, repo GrokMediaAssetRepository) *GrokMediaAssetService {
	if repo == nil {
		return nil
	}
	return &GrokMediaAssetService{
		gatewayService: gatewayService,
		repo:           repo,
		cacheRoot:      defaultGrokMediaCacheRoot(gatewayService),
		retention:      defaultGrokMediaAssetRetention,
		gcInterval:     defaultGrokMediaAssetGCInterval,
		gcBatch:        defaultGrokMediaAssetGCBatch,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s *GrokMediaAssetService) RewriteResponse(
	c *gin.Context,
	account *Account,
	body []byte,
	assetType string,
	requestedModel string,
	canonicalModel string,
	jobID string,
) ([]byte, string, error) {
	if s == nil || s.repo == nil || c == nil || account == nil || len(body) == 0 || !json.Valid(body) {
		return body, "", nil
	}
	s.maybeCleanupExpired(c.Request.Context())

	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return body, "", nil
	}

	rewriter := grokMediaResponseRewriter{
		service:        s,
		ctx:            c.Request.Context(),
		requestContext: c,
		account:        account,
		assetType:      strings.TrimSpace(assetType),
		requestedModel: strings.TrimSpace(requestedModel),
		canonicalModel: strings.TrimSpace(canonicalModel),
		jobID:          strings.TrimSpace(jobID),
	}
	rewriter.outputFormat, rewriter.proxyEnabled = s.resolveRewritePolicy(c.Request.Context(), rewriter.assetType)
	rewriter.walk(&payload)
	if !rewriter.changed {
		return body, "", nil
	}

	rewritten, err := json.Marshal(payload)
	if err != nil {
		return body, "", err
	}
	return rewritten, rewriter.primaryAssetID, nil
}

func (s *GrokMediaAssetService) Serve(c *gin.Context, assetID string) bool {
	if s == nil || s.repo == nil || c == nil {
		return false
	}
	s.maybeCleanupExpired(c.Request.Context())

	record, err := s.repo.GetByAssetID(c.Request.Context(), assetID)
	if err != nil {
		if errors.Is(err, ErrGrokMediaAssetNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "not_found_error",
					"message": "Grok media asset is not known to this gateway",
				},
			})
			return true
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "api_error",
				"message": "Failed to load Grok media asset",
			},
		})
		return true
	}
	if record == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "not_found_error",
				"message": "Grok media asset is not known to this gateway",
			},
		})
		return true
	}

	localPath, mimeType, err := s.ensureLocalAsset(c.Request.Context(), record)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"code":    "api_error",
				"message": "Failed to load Grok media asset content",
			},
		})
		return true
	}

	accessAt := s.now()
	_ = s.repo.MarkAccessed(c.Request.Context(), record.AssetID, accessAt, s.expiryAt(c.Request.Context(), accessAt))

	if mimeType != "" {
		c.Header("Content-Type", mimeType)
	}
	http.ServeFile(c.Writer, c.Request, localPath)
	return true
}

func (s *GrokMediaAssetService) createProxyAssetRecord(
	ctx context.Context,
	account *Account,
	assetType string,
	requestedModel string,
	canonicalModel string,
	jobID string,
	upstreamURL string,
) (*GrokMediaAssetRecord, error) {
	return s.upsertRemoteAssetRecord(ctx, account, assetType, requestedModel, canonicalModel, jobID, "", upstreamURL)
}

func (s *GrokMediaAssetService) UpsertRemoteAssetRecord(
	ctx context.Context,
	account *Account,
	assetType string,
	requestedModel string,
	canonicalModel string,
	jobID string,
	assetID string,
	upstreamURL string,
) (*GrokMediaAssetRecord, error) {
	return s.upsertRemoteAssetRecord(ctx, account, assetType, requestedModel, canonicalModel, jobID, assetID, upstreamURL)
}

func (s *GrokMediaAssetService) upsertRemoteAssetRecord(
	ctx context.Context,
	account *Account,
	assetType string,
	requestedModel string,
	canonicalModel string,
	jobID string,
	assetID string,
	upstreamURL string,
) (*GrokMediaAssetRecord, error) {
	if s == nil || s.repo == nil || account == nil {
		return nil, errors.New("grok media asset service is not configured")
	}

	upstreamURL = strings.TrimSpace(upstreamURL)
	if upstreamURL == "" {
		return nil, errors.New("upstream url is empty")
	}

	assetID = firstNonEmpty(strings.TrimSpace(assetID), uuid.NewString())
	record := GrokMediaAssetRecord{
		AssetID:        assetID,
		AccountID:      account.ID,
		JobID:          strings.TrimSpace(jobID),
		RequestedModel: strings.TrimSpace(requestedModel),
		CanonicalModel: strings.TrimSpace(canonicalModel),
		AssetType:      strings.TrimSpace(assetType),
		UpstreamURL:    upstreamURL,
		Status:         "remote",
		ExpiresAt:      s.expiryAt(ctx, s.now()),
	}
	if err := s.repo.Upsert(ctx, record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *GrokMediaAssetService) RenderExistingAssetValue(c *gin.Context, assetID string, assetType string) (string, string, error) {
	if s == nil || s.repo == nil {
		return "", "", errors.New("grok media asset service is not configured")
	}
	ctx := context.Background()
	if c != nil && c.Request != nil && c.Request.Context() != nil {
		ctx = c.Request.Context()
	}

	record, err := s.repo.GetByAssetID(ctx, strings.TrimSpace(assetID))
	if err != nil {
		return "", "", err
	}
	if record == nil {
		return "", "", ErrGrokMediaAssetNotFound
	}

	resolvedAssetType := firstNonEmpty(strings.TrimSpace(assetType), strings.TrimSpace(record.AssetType), "image")
	outputFormat, proxyEnabled := s.resolveRewritePolicy(ctx, resolvedAssetType)
	upstreamURL := strings.TrimSpace(record.UpstreamURL)
	if upstreamURL == "" {
		return "", "", errors.New("grok media asset upstream url is empty")
	}

	switch outputFormat {
	case GrokMediaOutputFormatUpstreamURL:
		return upstreamURL, upstreamURL, nil
	case GrokMediaOutputFormatMarkdown:
		renderURL := upstreamURL
		if proxyEnabled {
			renderURL = s.BuildLocalURL(c, record.AssetID)
		}
		return fmt.Sprintf("![grok-image](%s)", renderURL), upstreamURL, nil
	case GrokMediaOutputFormatBase64:
		localPath, mimeType, err := s.ensureLocalAsset(ctx, record)
		if err != nil {
			return "", "", err
		}
		payload, err := os.ReadFile(localPath)
		if err != nil {
			return "", "", err
		}
		mimeType = firstNonEmpty(strings.TrimSpace(mimeType), http.DetectContentType(payload))
		return "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(payload), upstreamURL, nil
	case GrokMediaOutputFormatHTML:
		renderURL := upstreamURL
		if proxyEnabled {
			renderURL = s.BuildLocalURL(c, record.AssetID)
		}
		return fmt.Sprintf(`<video controls src="%s"></video>`, html.EscapeString(renderURL)), upstreamURL, nil
	default:
		if !proxyEnabled {
			return upstreamURL, upstreamURL, nil
		}
		return s.BuildLocalURL(c, record.AssetID), upstreamURL, nil
	}
}

func (s *GrokMediaAssetService) BuildLocalURL(c *gin.Context, assetID string) string {
	path := strings.TrimRight(grokMediaAssetRoutePrefix, "/") + "/" + strings.TrimSpace(assetID)
	if c == nil || c.Request == nil {
		return path
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if forwarded := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")); forwarded != "" {
		scheme = strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	host := strings.TrimSpace(c.Request.Host)
	if host == "" {
		return path
	}
	return scheme + "://" + host + path
}

func (s *GrokMediaAssetService) ensureLocalAsset(ctx context.Context, record *GrokMediaAssetRecord) (string, string, error) {
	if record == nil {
		return "", "", ErrGrokMediaAssetNotFound
	}
	if localPath := strings.TrimSpace(record.LocalPath); localPath != "" {
		if stat, err := os.Stat(localPath); err == nil && !stat.IsDir() {
			return localPath, firstNonEmpty(strings.TrimSpace(record.MimeType), mime.TypeByExtension(filepath.Ext(localPath))), nil
		}
	}

	result, err, _ := s.downloads.Do(strings.TrimSpace(record.AssetID), func() (any, error) {
		freshRecord, getErr := s.repo.GetByAssetID(ctx, record.AssetID)
		if getErr != nil {
			return nil, getErr
		}
		if freshRecord != nil {
			if localPath := strings.TrimSpace(freshRecord.LocalPath); localPath != "" {
				if stat, err := os.Stat(localPath); err == nil && !stat.IsDir() {
					return []string{localPath, firstNonEmpty(strings.TrimSpace(freshRecord.MimeType), mime.TypeByExtension(filepath.Ext(localPath)))}, nil
				}
			}
			record = freshRecord
		}
		return s.downloadAndCache(ctx, record)
	})
	if err != nil {
		return "", "", err
	}
	pair, ok := result.([]string)
	if !ok || len(pair) != 2 {
		return "", "", errors.New("unexpected grok media asset cache result")
	}
	return pair[0], pair[1], nil
}

func (s *GrokMediaAssetService) downloadAndCache(ctx context.Context, record *GrokMediaAssetRecord) (any, error) {
	if record == nil || s.gatewayService == nil || s.gatewayService.httpUpstream == nil {
		return nil, errors.New("grok media asset download is not configured")
	}

	upstreamURL := strings.TrimSpace(record.UpstreamURL)
	parsed, err := url.Parse(upstreamURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("invalid grok media upstream url")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, upstreamURL, nil)
	if err != nil {
		return nil, err
	}

	account, _ := s.gatewayService.accountRepo.GetByID(ctx, record.AccountID)
	if account != nil {
		runtimeSettings := DefaultGrokRuntimeSettings()
		if s.gatewayService != nil && s.gatewayService.settingService != nil {
			runtimeSettings = s.gatewayService.settingService.GetGrokRuntimeSettings(ctx)
		}
		if target, targetErr := resolveGrokTransportTargetWithSettings(
			account,
			s.gatewayService.validateUpstreamBaseURL,
			runtimeSettings,
		); targetErr == nil {
			target.Apply(req)
			if target.Kind == grokTransportKindSession {
				applyGrokSessionBrowserHeaders(req.Header, target, "")
			}
		}
	}

	resp, err := s.gatewayService.httpUpstream.DoWithTLS(
		req,
		resolveGrokMediaAssetProxyURL(account),
		record.AccountID,
		resolveAccountConcurrency(account),
		resolveGrokGatewayTLSProfile(s.gatewayService, account),
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("unexpected grok media asset status: %d", resp.StatusCode)
	}

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	cachePath, mimeType, hashValue, err := s.writeCachedFile(ctx, parsed, resp.Header.Get("Content-Type"), payload)
	if err != nil {
		return nil, err
	}

	accessAt := s.now()
	if err := s.repo.UpdateCacheState(ctx, GrokMediaAssetCachePatch{
		AssetID:      record.AssetID,
		LocalPath:    cachePath,
		ContentHash:  hashValue,
		MimeType:     mimeType,
		SizeBytes:    int64(len(payload)),
		Status:       grokMediaAssetStatusReady,
		ExpiresAt:    s.expiryAt(ctx, accessAt),
		LastAccessAt: &accessAt,
	}); err != nil {
		return nil, err
	}

	return []string{cachePath, mimeType}, nil
}

func (s *GrokMediaAssetService) writeCachedFile(
	ctx context.Context,
	parsedURL *url.URL,
	contentType string,
	payload []byte,
) (string, string, string, error) {
	if err := os.MkdirAll(s.cacheRoot, 0o755); err != nil {
		return "", "", "", err
	}

	mimeType := firstNonEmpty(strings.TrimSpace(strings.Split(contentType, ";")[0]), mime.TypeByExtension(filepath.Ext(parsedURL.Path)), http.DetectContentType(payload))
	ext := grokMediaAssetExtension(mimeType, parsedURL.Path)

	sum := sha256.Sum256(payload)
	hashValue := hex.EncodeToString(sum[:])
	if reusedPath, reusedMime, ok := s.reuseCachedFile(ctx, hashValue); ok {
		return reusedPath, firstNonEmpty(reusedMime, mimeType), hashValue, nil
	}
	cachePath := filepath.Join(s.cacheRoot, hashValue+ext)

	if stat, err := os.Stat(cachePath); err == nil && !stat.IsDir() {
		return cachePath, mimeType, hashValue, nil
	}
	if err := os.WriteFile(cachePath, payload, 0o644); err != nil {
		return "", "", "", err
	}
	return cachePath, mimeType, hashValue, nil
}

func (s *GrokMediaAssetService) reuseCachedFile(ctx context.Context, hashValue string) (string, string, bool) {
	if s == nil || s.repo == nil || strings.TrimSpace(hashValue) == "" {
		return "", "", false
	}

	record, err := s.repo.FindCachedByHash(ctx, hashValue)
	if err != nil || record == nil {
		return "", "", false
	}
	localPath := strings.TrimSpace(record.LocalPath)
	if localPath == "" {
		return "", "", false
	}
	if stat, err := os.Stat(localPath); err == nil && !stat.IsDir() {
		return localPath, strings.TrimSpace(record.MimeType), true
	}
	return "", "", false
}

func (s *GrokMediaAssetService) maybeCleanupExpired(ctx context.Context) {
	if s == nil || s.repo == nil || s.gcInterval <= 0 {
		return
	}

	now := s.now().UTC()
	lastUnix := s.lastCleanupAt.Load()
	if lastUnix > 0 && now.Sub(time.Unix(0, lastUnix).UTC()) < s.gcInterval {
		return
	}
	if !s.lastCleanupAt.CompareAndSwap(lastUnix, now.UnixNano()) {
		return
	}
	if err := s.CleanupExpiredNow(ctx); err != nil {
		logger.LegacyPrintf("service.grok_media_asset", "Warning: cleanup expired assets failed: %v", err)
	}
}

func (s *GrokMediaAssetService) CleanupExpiredNow(ctx context.Context) error {
	if s == nil || s.repo == nil || s.gcBatch <= 0 {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	cutoff := s.now().UTC()
	for {
		expired, err := s.repo.DeleteExpired(ctx, cutoff, s.gcBatch)
		if err != nil {
			return err
		}
		if len(expired) == 0 {
			return nil
		}
		if err := s.removeUnreferencedFiles(ctx, expired); err != nil {
			return err
		}
	}
}

func (s *GrokMediaAssetService) removeUnreferencedFiles(ctx context.Context, records []GrokMediaAssetRecord) error {
	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		localPath := strings.TrimSpace(record.LocalPath)
		if localPath == "" {
			continue
		}
		if _, ok := seen[localPath]; ok {
			continue
		}
		seen[localPath] = struct{}{}

		refCount, err := s.repo.CountByLocalPath(ctx, localPath)
		if err != nil {
			return err
		}
		if refCount > 0 {
			continue
		}
		if err := os.Remove(localPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

func (s *GrokMediaAssetService) expiryAt(ctx context.Context, base time.Time) *time.Time {
	if s == nil {
		return nil
	}
	retention := s.retention
	if s.gatewayService != nil && s.gatewayService.settingService != nil {
		if configured := s.gatewayService.settingService.GetGrokMediaCacheRetention(ctx); configured > 0 {
			retention = configured
		}
	}
	if retention <= 0 {
		return nil
	}
	expiresAt := base.UTC().Add(retention)
	return &expiresAt
}

func defaultGrokMediaCacheRoot(gatewayService *GatewayService) string {
	if gatewayService != nil && gatewayService.cfg != nil {
		if pricingDir := strings.TrimSpace(gatewayService.cfg.Pricing.DataDir); pricingDir != "" {
			return filepath.Join(pricingDir, "grok-media")
		}
	}
	return filepath.Join("data", "grok-media")
}

func grokMediaAssetExtension(mimeType string, sourcePath string) string {
	if exts, err := mime.ExtensionsByType(mimeType); err == nil && len(exts) > 0 {
		return exts[0]
	}
	if ext := filepath.Ext(strings.TrimSpace(sourcePath)); ext != "" {
		return ext
	}
	return ".bin"
}

func resolveGrokMediaAssetProxyURL(account *Account) string {
	if account != nil && account.Proxy != nil {
		return account.Proxy.URL()
	}
	return ""
}

func resolveAccountConcurrency(account *Account) int {
	if account != nil {
		return account.Concurrency
	}
	return 0
}

type grokMediaResponseRewriter struct {
	service        *GrokMediaAssetService
	ctx            context.Context
	requestContext *gin.Context
	account        *Account
	assetType      string
	requestedModel string
	canonicalModel string
	jobID          string
	outputFormat   string
	proxyEnabled   bool
	primaryAssetID string
	changed        bool
}

func (r *grokMediaResponseRewriter) walk(node *any) {
	switch typed := (*node).(type) {
	case map[string]any:
		r.rewriteMap(typed)
		for key, value := range typed {
			child := value
			r.walk(&child)
			typed[key] = child
		}
	case []any:
		for index, value := range typed {
			child := value
			r.walk(&child)
			typed[index] = child
		}
	}
}

func (r *grokMediaResponseRewriter) rewriteMap(node map[string]any) {
	for _, key := range []string{"url", "content_url"} {
		rawValue, ok := node[key]
		if !ok {
			continue
		}
		rawURL, ok := rawValue.(string)
		if !ok || !isProxyableMediaURL(rawURL) {
			continue
		}
		renderedValue, assetID, err := r.renderValue(rawURL)
		if err != nil {
			continue
		}
		if renderedValue == rawURL {
			continue
		}
		switch key {
		case "url":
			node["upstream_url"] = rawURL
		case "content_url":
			node["upstream_content_url"] = rawURL
		}
		node[key] = renderedValue
		if r.primaryAssetID == "" && assetID != "" {
			r.primaryAssetID = assetID
		}
		r.changed = true
	}
}

func (s *GrokMediaAssetService) resolveRewritePolicy(ctx context.Context, assetType string) (string, bool) {
	settings := DefaultGrokMediaSettings()
	if s != nil && s.gatewayService != nil && s.gatewayService.settingService != nil {
		settings = s.gatewayService.settingService.GetGrokMediaSettings(ctx)
	}

	switch strings.ToLower(strings.TrimSpace(assetType)) {
	case "video":
		return settings.VideoOutputFormat, settings.MediaProxyEnabled
	default:
		return settings.ImageOutputFormat, settings.MediaProxyEnabled
	}
}

func (r *grokMediaResponseRewriter) renderValue(rawURL string) (string, string, error) {
	switch r.outputFormat {
	case GrokMediaOutputFormatUpstreamURL:
		return rawURL, "", nil
	case GrokMediaOutputFormatMarkdown:
		renderURL, assetID, err := r.renderURL(rawURL)
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("![grok-image](%s)", renderURL), assetID, nil
	case GrokMediaOutputFormatBase64:
		return r.renderBase64(rawURL)
	case GrokMediaOutputFormatHTML:
		renderURL, assetID, err := r.renderURL(rawURL)
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf(`<video controls src="%s"></video>`, html.EscapeString(renderURL)), assetID, nil
	default:
		return r.renderURL(rawURL)
	}
}

func (r *grokMediaResponseRewriter) renderURL(rawURL string) (string, string, error) {
	if !r.proxyEnabled {
		return rawURL, "", nil
	}
	record, err := r.service.createProxyAssetRecord(
		r.ctx,
		r.account,
		r.assetType,
		r.requestedModel,
		r.canonicalModel,
		r.jobID,
		rawURL,
	)
	if err != nil {
		return "", "", err
	}
	return r.service.BuildLocalURL(r.requestContext, record.AssetID), record.AssetID, nil
}

func (r *grokMediaResponseRewriter) renderBase64(rawURL string) (string, string, error) {
	record, err := r.service.createProxyAssetRecord(
		r.ctx,
		r.account,
		r.assetType,
		r.requestedModel,
		r.canonicalModel,
		r.jobID,
		rawURL,
	)
	if err != nil {
		return "", "", err
	}
	localPath, mimeType, err := r.service.ensureLocalAsset(r.ctx, record)
	if err != nil {
		return "", "", err
	}
	payload, err := os.ReadFile(localPath)
	if err != nil {
		return "", "", err
	}
	mimeType = firstNonEmpty(strings.TrimSpace(mimeType), http.DetectContentType(payload))
	return "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(payload), record.AssetID, nil
}

func isProxyableMediaURL(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}
