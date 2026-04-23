package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	dbent "github.com/senran-N/sub2api/ent"
	"github.com/senran-N/sub2api/ent/pendingauthsession"
	infraerrors "github.com/senran-N/sub2api/internal/pkg/errors"
)

var (
	ErrPendingAuthSessionNotFound = infraerrors.NotFound("PENDING_AUTH_SESSION_NOT_FOUND", "pending auth session not found")
	ErrPendingAuthSessionExpired  = infraerrors.Unauthorized("PENDING_AUTH_SESSION_EXPIRED", "pending auth session has expired")
	ErrPendingAuthSessionConsumed = infraerrors.Unauthorized("PENDING_AUTH_SESSION_CONSUMED", "pending auth session has already been used")
	ErrPendingAuthBrowserMismatch = infraerrors.Unauthorized("PENDING_AUTH_BROWSER_MISMATCH", "pending auth session does not match this browser session")
)

const defaultPendingAuthTTL = 15 * time.Minute

type PendingAuthIdentityKey struct {
	ProviderType    string
	ProviderKey     string
	ProviderSubject string
}

type CreatePendingAuthSessionInput struct {
	SessionToken             string
	Intent                   string
	Identity                 PendingAuthIdentityKey
	TargetUserID             *int64
	RedirectTo               string
	ResolvedEmail            string
	RegistrationPasswordHash string
	BrowserSessionKey        string
	UpstreamIdentityClaims   map[string]any
	LocalFlowState           map[string]any
	ExpiresAt                time.Time
}

type AuthPendingIdentityService struct {
	entClient *dbent.Client
}

func NewAuthPendingIdentityService(entClient *dbent.Client) *AuthPendingIdentityService {
	return &AuthPendingIdentityService{entClient: entClient}
}

func (s *AuthPendingIdentityService) CreatePendingSession(ctx context.Context, input CreatePendingAuthSessionInput) (*dbent.PendingAuthSession, error) {
	if s == nil || s.entClient == nil {
		return nil, fmt.Errorf("pending auth ent client is not configured")
	}

	sessionToken := strings.TrimSpace(input.SessionToken)
	if sessionToken == "" {
		generated, err := randomOpaqueToken(24)
		if err != nil {
			return nil, err
		}
		sessionToken = generated
	}

	expiresAt := input.ExpiresAt.UTC()
	if expiresAt.IsZero() {
		expiresAt = time.Now().UTC().Add(defaultPendingAuthTTL)
	}

	create := s.entClient.PendingAuthSession.Create().
		SetSessionToken(sessionToken).
		SetIntent(strings.TrimSpace(input.Intent)).
		SetProviderType(strings.TrimSpace(input.Identity.ProviderType)).
		SetProviderKey(strings.TrimSpace(input.Identity.ProviderKey)).
		SetProviderSubject(strings.TrimSpace(input.Identity.ProviderSubject)).
		SetRedirectTo(strings.TrimSpace(input.RedirectTo)).
		SetResolvedEmail(strings.TrimSpace(input.ResolvedEmail)).
		SetRegistrationPasswordHash(strings.TrimSpace(input.RegistrationPasswordHash)).
		SetBrowserSessionKey(strings.TrimSpace(input.BrowserSessionKey)).
		SetUpstreamIdentityClaims(copyPendingMap(input.UpstreamIdentityClaims)).
		SetLocalFlowState(copyPendingMap(input.LocalFlowState)).
		SetExpiresAt(expiresAt)
	if input.TargetUserID != nil && *input.TargetUserID > 0 {
		create = create.SetTargetUserID(*input.TargetUserID)
	}
	return create.Save(ctx)
}

func (s *AuthPendingIdentityService) GetBrowserSession(ctx context.Context, sessionToken, browserSessionKey string) (*dbent.PendingAuthSession, error) {
	if s == nil || s.entClient == nil {
		return nil, fmt.Errorf("pending auth ent client is not configured")
	}

	session, err := s.getBrowserSession(ctx, sessionToken)
	if err != nil {
		return nil, err
	}
	if err := validatePendingSessionState(session, browserSessionKey); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *AuthPendingIdentityService) ConsumeBrowserSession(ctx context.Context, sessionToken, browserSessionKey string) (*dbent.PendingAuthSession, error) {
	if s == nil || s.entClient == nil {
		return nil, fmt.Errorf("pending auth ent client is not configured")
	}

	session, err := s.getBrowserSession(ctx, sessionToken)
	if err != nil {
		return nil, err
	}
	if err := validatePendingSessionState(session, browserSessionKey); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	updated, err := s.entClient.PendingAuthSession.UpdateOneID(session.ID).
		Where(
			pendingauthsession.ConsumedAtIsNil(),
			pendingauthsession.ExpiresAtGTE(now),
		).
		SetConsumedAt(now).
		Save(ctx)
	if err == nil {
		return updated, nil
	}
	if !dbent.IsNotFound(err) {
		return nil, err
	}

	current, currentErr := s.entClient.PendingAuthSession.Get(ctx, session.ID)
	if currentErr != nil {
		if dbent.IsNotFound(currentErr) {
			return nil, ErrPendingAuthSessionNotFound
		}
		return nil, currentErr
	}
	if err := validatePendingSessionState(current, browserSessionKey); err != nil {
		return nil, err
	}
	return nil, ErrPendingAuthSessionConsumed
}

func (s *AuthPendingIdentityService) getBrowserSession(ctx context.Context, sessionToken string) (*dbent.PendingAuthSession, error) {
	sessionToken = strings.TrimSpace(sessionToken)
	if sessionToken == "" {
		return nil, ErrPendingAuthSessionNotFound
	}

	session, err := s.entClient.PendingAuthSession.Query().
		Where(pendingauthsession.SessionTokenEQ(sessionToken)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrPendingAuthSessionNotFound
		}
		return nil, err
	}
	return session, nil
}

func validatePendingSessionState(session *dbent.PendingAuthSession, browserSessionKey string) error {
	if session == nil {
		return ErrPendingAuthSessionNotFound
	}

	now := time.Now().UTC()
	if session.ConsumedAt != nil {
		return ErrPendingAuthSessionConsumed
	}
	if !session.ExpiresAt.IsZero() && now.After(session.ExpiresAt) {
		return ErrPendingAuthSessionExpired
	}
	if strings.TrimSpace(session.BrowserSessionKey) != "" && strings.TrimSpace(browserSessionKey) != strings.TrimSpace(session.BrowserSessionKey) {
		return ErrPendingAuthBrowserMismatch
	}
	return nil
}

func copyPendingMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func randomOpaqueToken(byteLen int) (string, error) {
	if byteLen <= 0 {
		byteLen = 16
	}
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
