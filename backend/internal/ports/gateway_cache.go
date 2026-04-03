package ports

import (
	"context"
	"time"
)

// GatewayCache stores sticky session bindings for gateway account selection.
type GatewayCache interface {
	GetSessionAccountID(ctx context.Context, groupID int64, sessionHash string) (int64, error)
	SetSessionAccountID(ctx context.Context, groupID int64, sessionHash string, accountID int64, ttl time.Duration) error
	RefreshSessionTTL(ctx context.Context, groupID int64, sessionHash string, ttl time.Duration) error
	DeleteSessionAccountID(ctx context.Context, groupID int64, sessionHash string) error
}
