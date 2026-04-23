package service

import (
	"context"
	"errors"
	"strings"
	"time"

	dbent "github.com/senran-N/sub2api/ent"
	"github.com/senran-N/sub2api/ent/authidentity"
	infraerrors "github.com/senran-N/sub2api/internal/pkg/errors"
)

var ErrAuthIdentityConflict = infraerrors.Conflict("AUTH_IDENTITY_CONFLICT", "auth identity is already bound to another user")

func (s *AuthService) EntClient() *dbent.Client {
	if s == nil {
		return nil
	}
	return s.entClient
}

func (s *AuthService) BindOAuthIdentity(ctx context.Context, userID int64, providerType, providerKey, providerSubject string, metadata map[string]any) error {
	if s == nil || s.entClient == nil || userID <= 0 {
		return ErrServiceUnavailable
	}
	providerType = strings.ToLower(strings.TrimSpace(providerType))
	providerKey = strings.TrimSpace(providerKey)
	providerSubject = strings.TrimSpace(providerSubject)
	if providerType == "" || providerKey == "" || providerSubject == "" {
		return infraerrors.BadRequest("INVALID_AUTH_IDENTITY", "invalid auth identity")
	}
	if metadata == nil {
		metadata = map[string]any{}
	}
	now := time.Now().UTC()
	if err := s.entClient.AuthIdentity.Create().
		SetUserID(userID).
		SetProviderType(providerType).
		SetProviderKey(providerKey).
		SetProviderSubject(providerSubject).
		SetVerifiedAt(now).
		SetMetadata(metadata).
		OnConflictColumns(authidentity.FieldProviderType, authidentity.FieldProviderKey, authidentity.FieldProviderSubject).
		Ignore().
		Exec(ctx); err != nil {
		return err
	}
	identity, err := s.entClient.AuthIdentity.Query().
		Where(authidentity.ProviderTypeEQ(providerType), authidentity.ProviderKeyEQ(providerKey), authidentity.ProviderSubjectEQ(providerSubject)).
		Only(ctx)
	if err != nil {
		return err
	}
	if identity.UserID != userID {
		return ErrAuthIdentityConflict
	}
	return nil
}

func (s *AuthService) ValidatePasswordCredentials(ctx context.Context, email, password string) (*User, error) {
	if s == nil {
		return nil, ErrServiceUnavailable
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, ErrServiceUnavailable
	}
	if !s.CheckPassword(password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive() {
		return nil, ErrUserNotActive
	}

	s.BackfillEmailIdentityOnSuccessfulLogin(ctx, user)
	return user, nil
}

func (s *AuthService) revokeAuthSessionsDetached(userID int64) {
	if s == nil || userID <= 0 {
		return
	}
	go func() {
		_ = s.RevokeAllUserSessions(context.Background(), userID)
	}()
}
