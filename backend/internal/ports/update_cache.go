package ports

import (
	"context"
	"time"
)

// UpdateCache stores serialized update-check results.
type UpdateCache interface {
	GetUpdateInfo(ctx context.Context) (string, error)
	SetUpdateInfo(ctx context.Context, data string, ttl time.Duration) error
}
