package ports

import (
	"context"

	"github.com/senran-N/sub2api/internal/model"
)

// ErrorPassthroughRepository stores and queries error passthrough rules.
type ErrorPassthroughRepository interface {
	List(ctx context.Context) ([]*model.ErrorPassthroughRule, error)
	GetByID(ctx context.Context, id int64) (*model.ErrorPassthroughRule, error)
	Create(ctx context.Context, rule *model.ErrorPassthroughRule) (*model.ErrorPassthroughRule, error)
	Update(ctx context.Context, rule *model.ErrorPassthroughRule) (*model.ErrorPassthroughRule, error)
	Delete(ctx context.Context, id int64) error
}

// ErrorPassthroughCache caches error passthrough rules for fast matching.
type ErrorPassthroughCache interface {
	Get(ctx context.Context) ([]*model.ErrorPassthroughRule, bool)
	Set(ctx context.Context, rules []*model.ErrorPassthroughRule) error
	Invalidate(ctx context.Context) error
	NotifyUpdate(ctx context.Context) error
	SubscribeUpdates(ctx context.Context, handler func())
}
