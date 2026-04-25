package service

import (
	"context"
	"strings"
)

type RuntimeSessionPrepareRequest struct {
	Parsed               *ParsedRequest
	Body                 []byte
	ParseProtocol        string
	SessionHash          string
	Model                string
	Stream               bool
	ClientIP             string
	UserAgent            string
	APIKeyID             int64
	GroupID              *int64
	SessionKeyPrefix     string
	BridgeLegacyMetadata bool
}

type RuntimeSessionPrepareResult struct {
	Context        context.Context
	Parsed         *ParsedRequest
	SessionHash    string
	SessionKey     string
	BoundAccountID int64
	Prefetched     RuntimeStickySessionPrefetchResult
}

type RuntimeStickySessionPrefetchRequest struct {
	GroupID              *int64
	SessionKey           string
	BridgeLegacyMetadata bool
}

type RuntimeStickySessionPrefetchResult struct {
	Context         context.Context
	AccountID       int64
	GroupID         int64
	StickyLookupErr error
}

func (s *GatewayService) PrepareRuntimeSession(ctx context.Context, req RuntimeSessionPrepareRequest) RuntimeSessionPrepareResult {
	if ctx == nil {
		ctx = context.Background()
	}

	parsed := req.Parsed
	if parsed == nil {
		parsed, _ = ParseGatewayRequest(req.Body, req.ParseProtocol)
	}
	if parsed == nil {
		parsed = &ParsedRequest{Model: req.Model, Stream: req.Stream, Body: req.Body}
	}
	if parsed.Body == nil {
		parsed.Body = req.Body
	}
	parsed.SessionContext = &SessionContext{
		ClientIP:  req.ClientIP,
		UserAgent: req.UserAgent,
		APIKeyID:  req.APIKeyID,
	}

	gatewayService := s
	if gatewayService == nil {
		gatewayService = &GatewayService{}
	}
	sessionHash := strings.TrimSpace(req.SessionHash)
	if sessionHash == "" {
		sessionHash = gatewayService.GenerateSessionHash(parsed)
	}
	sessionKey := sessionHash
	if sessionHash != "" && req.SessionKeyPrefix != "" {
		sessionKey = req.SessionKeyPrefix + sessionHash
	}

	prefetched := gatewayService.PrefetchRuntimeStickySession(ctx, RuntimeStickySessionPrefetchRequest{
		GroupID:              req.GroupID,
		SessionKey:           sessionKey,
		BridgeLegacyMetadata: req.BridgeLegacyMetadata,
	})

	return RuntimeSessionPrepareResult{
		Context:        prefetched.Context,
		Parsed:         parsed,
		SessionHash:    sessionHash,
		SessionKey:     sessionKey,
		BoundAccountID: prefetched.AccountID,
		Prefetched:     prefetched,
	}
}

func (s *GatewayService) PrefetchRuntimeStickySession(ctx context.Context, req RuntimeStickySessionPrefetchRequest) RuntimeStickySessionPrefetchResult {
	if ctx == nil {
		ctx = context.Background()
	}

	result := RuntimeStickySessionPrefetchResult{
		Context: ctx,
		GroupID: derefGroupID(req.GroupID),
	}
	if s == nil || req.SessionKey == "" {
		return result
	}

	accountID, err := s.GetCachedSessionAccountID(ctx, req.GroupID, req.SessionKey)
	result.StickyLookupErr = err
	if err != nil || accountID <= 0 {
		return result
	}

	result.AccountID = accountID
	result.Context = WithPrefetchedStickySession(ctx, accountID, result.GroupID, req.BridgeLegacyMetadata)
	return result
}
