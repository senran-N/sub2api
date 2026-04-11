//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type refreshTokenCleanupCacheStub struct {
	data                      *RefreshTokenData
	deleteRefreshTokenErr     error
	deleteRefreshTokenCtxErr  error
	deleteTokenFamilyCtxErr   error
	deleteUserTokensCtxErr    error
	deletedRefreshTokenHashes []string
	deletedFamilies           []string
	deletedUserIDs            []int64
	storedRefreshTokens       int
}

func (s *refreshTokenCleanupCacheStub) StoreRefreshToken(context.Context, string, *RefreshTokenData, time.Duration) error {
	s.storedRefreshTokens++
	return nil
}

func (s *refreshTokenCleanupCacheStub) GetRefreshToken(context.Context, string) (*RefreshTokenData, error) {
	return s.data, nil
}

func (s *refreshTokenCleanupCacheStub) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	s.deleteRefreshTokenCtxErr = ctx.Err()
	s.deletedRefreshTokenHashes = append(s.deletedRefreshTokenHashes, tokenHash)
	return s.deleteRefreshTokenErr
}

func (s *refreshTokenCleanupCacheStub) DeleteUserRefreshTokens(ctx context.Context, userID int64) error {
	s.deleteUserTokensCtxErr = ctx.Err()
	s.deletedUserIDs = append(s.deletedUserIDs, userID)
	return nil
}

func (s *refreshTokenCleanupCacheStub) DeleteTokenFamily(ctx context.Context, familyID string) error {
	s.deleteTokenFamilyCtxErr = ctx.Err()
	s.deletedFamilies = append(s.deletedFamilies, familyID)
	return nil
}

func (s *refreshTokenCleanupCacheStub) AddToUserTokenSet(context.Context, int64, string, time.Duration) error {
	return nil
}

func (s *refreshTokenCleanupCacheStub) AddToFamilyTokenSet(context.Context, string, string, time.Duration) error {
	return nil
}

func (s *refreshTokenCleanupCacheStub) GetUserTokenHashes(context.Context, int64) ([]string, error) {
	return nil, nil
}

func (s *refreshTokenCleanupCacheStub) GetFamilyTokenHashes(context.Context, string) ([]string, error) {
	return nil, nil
}

func (s *refreshTokenCleanupCacheStub) IsTokenInFamily(context.Context, string, string) (bool, error) {
	return false, nil
}

func TestAuthService_RefreshTokenPair_ExpiredTokenDeletesWithDetachedContext(t *testing.T) {
	refreshToken := refreshTokenPrefix + "expired-token"
	cache := &refreshTokenCleanupCacheStub{
		data: &RefreshTokenData{
			UserID:       1,
			FamilyID:     "family-1",
			TokenVersion: 1,
			ExpiresAt:    time.Now().Add(-time.Minute),
		},
	}
	svc := &AuthService{refreshTokenCache: cache}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	pair, err := svc.RefreshTokenPair(ctx, refreshToken)

	require.Nil(t, pair)
	require.ErrorIs(t, err, ErrRefreshTokenExpired)
	require.Equal(t, []string{hashToken(refreshToken)}, cache.deletedRefreshTokenHashes)
	require.NoError(t, cache.deleteRefreshTokenCtxErr)
}

func TestAuthService_RevokeAllUserSessions_UsesDetachedContext(t *testing.T) {
	cache := &refreshTokenCleanupCacheStub{}
	svc := &AuthService{refreshTokenCache: cache}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := svc.RevokeAllUserSessions(ctx, 42)

	require.NoError(t, err)
	require.Equal(t, []int64{42}, cache.deletedUserIDs)
	require.NoError(t, cache.deleteUserTokensCtxErr)
}

func TestAuthService_RefreshTokenPair_DeleteOldTokenFailureFailsClosed(t *testing.T) {
	refreshToken := refreshTokenPrefix + "rotate-token"
	cache := &refreshTokenCleanupCacheStub{
		data: &RefreshTokenData{
			UserID:       1,
			FamilyID:     "family-rotate",
			TokenVersion: 2,
			ExpiresAt:    time.Now().Add(time.Hour),
		},
		deleteRefreshTokenErr: context.DeadlineExceeded,
	}
	userRepo := &userRepoStub{
		user: &User{
			ID:           1,
			Email:        "rotate@test.com",
			Role:         RoleUser,
			Status:       StatusActive,
			TokenVersion: 2,
		},
	}
	svc := NewAuthService(
		nil,
		userRepo,
		nil,
		cache,
		&config.Config{JWT: config.JWTConfig{Secret: "test-secret", ExpireHour: 1, RefreshTokenExpireDays: 7}},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	pair, err := svc.RefreshTokenPair(context.Background(), refreshToken)

	require.Nil(t, pair)
	require.ErrorIs(t, err, ErrServiceUnavailable)
	require.Equal(t, []string{hashToken(refreshToken)}, cache.deletedRefreshTokenHashes)
	require.Equal(t, 0, cache.storedRefreshTokens)
}
