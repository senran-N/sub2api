package ports

import (
	"context"

	"github.com/senran-N/sub2api/internal/domain"
)

// IdentityCache stores per-account identity fingerprints and masked session ids.
type IdentityCache interface {
	GetFingerprint(ctx context.Context, accountID int64) (*domain.Fingerprint, error)
	SetFingerprint(ctx context.Context, accountID int64, fp *domain.Fingerprint) error
	GetMaskedSessionID(ctx context.Context, accountID int64) (string, error)
	SetMaskedSessionID(ctx context.Context, accountID int64, sessionID string) error
}
