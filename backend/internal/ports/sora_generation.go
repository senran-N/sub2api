package ports

import (
	"context"

	"github.com/senran-N/sub2api/internal/domain"
)

// SoraGenerationRepository stores and queries Sora generation records.
type SoraGenerationRepository interface {
	Create(ctx context.Context, gen *domain.SoraGeneration) error
	GetByID(ctx context.Context, id int64) (*domain.SoraGeneration, error)
	Update(ctx context.Context, gen *domain.SoraGeneration) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, params domain.SoraGenerationListParams) ([]*domain.SoraGeneration, int64, error)
	CountByUserAndStatus(ctx context.Context, userID int64, statuses []string) (int64, error)
}
