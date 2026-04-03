package ports

import (
	"context"
	"time"

	"github.com/senran-N/sub2api/internal/domain"
)

// ScheduledTestPlanRepository stores and queries scheduled test plans.
type ScheduledTestPlanRepository interface {
	Create(ctx context.Context, plan *domain.ScheduledTestPlan) (*domain.ScheduledTestPlan, error)
	GetByID(ctx context.Context, id int64) (*domain.ScheduledTestPlan, error)
	ListByAccountID(ctx context.Context, accountID int64) ([]*domain.ScheduledTestPlan, error)
	ListDue(ctx context.Context, now time.Time) ([]*domain.ScheduledTestPlan, error)
	Update(ctx context.Context, plan *domain.ScheduledTestPlan) (*domain.ScheduledTestPlan, error)
	Delete(ctx context.Context, id int64) error
	UpdateAfterRun(ctx context.Context, id int64, lastRunAt time.Time, nextRunAt time.Time) error
}

// ScheduledTestResultRepository stores and queries scheduled test results.
type ScheduledTestResultRepository interface {
	Create(ctx context.Context, result *domain.ScheduledTestResult) (*domain.ScheduledTestResult, error)
	ListByPlanID(ctx context.Context, planID int64, limit int) ([]*domain.ScheduledTestResult, error)
	PruneOldResults(ctx context.Context, planID int64, keepCount int) error
}
