package ports

import (
	"context"
	"time"

	"github.com/senran-N/sub2api/internal/domain"
)

// RefreshTokenCache stores and invalidates JWT refresh token state.
type RefreshTokenCache interface {
	StoreRefreshToken(ctx context.Context, tokenHash string, data *domain.RefreshTokenData, ttl time.Duration) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*domain.RefreshTokenData, error)
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
	DeleteUserRefreshTokens(ctx context.Context, userID int64) error
	DeleteTokenFamily(ctx context.Context, familyID string) error
	AddToUserTokenSet(ctx context.Context, userID int64, tokenHash string, ttl time.Duration) error
	AddToFamilyTokenSet(ctx context.Context, familyID string, tokenHash string, ttl time.Duration) error
	GetUserTokenHashes(ctx context.Context, userID int64) ([]string, error)
	GetFamilyTokenHashes(ctx context.Context, familyID string) ([]string, error)
	IsTokenInFamily(ctx context.Context, familyID string, tokenHash string) (bool, error)
}
