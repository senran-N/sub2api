package service

import (
	"context"
	"net/http"
	"time"
)

const grokAccountStatePersistTimeout = 5 * time.Second

type grokAccountStateExtraWriter interface {
	UpdateExtra(ctx context.Context, id int64, updates map[string]any) error
}

type GrokAccountStateService struct {
	accountRepo grokAccountStateExtraWriter
	now         func() time.Time
}

func NewGrokAccountStateService(accountRepo grokAccountStateExtraWriter) *GrokAccountStateService {
	return &GrokAccountStateService{
		accountRepo: accountRepo,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s *GrokAccountStateService) PersistProbeResult(
	ctx context.Context,
	account *Account,
	modelID string,
	resp *http.Response,
	probeErr error,
) {
	if s == nil || s.accountRepo == nil || account == nil {
		return
	}
	if NormalizeCompatibleGatewayPlatform(account.Platform) != PlatformGrok {
		return
	}

	updates := buildGrokProbeStateExtraUpdates(account, modelID, resp, probeErr, s.now())
	s.persistExtraUpdates(ctx, account, updates)
}

func (s *GrokAccountStateService) PersistSyncSnapshot(ctx context.Context, account *Account, snapshot grokStateSyncSnapshot) {
	if s == nil || s.accountRepo == nil || account == nil {
		return
	}
	if NormalizeCompatibleGatewayPlatform(account.Platform) != PlatformGrok {
		return
	}

	updates := buildGrokSyncStateExtraUpdates(account, snapshot)
	s.persistExtraUpdates(ctx, account, updates)
}

func (s *GrokAccountStateService) persistExtraUpdates(ctx context.Context, account *Account, updates map[string]any) {
	if s == nil || s.accountRepo == nil || account == nil || len(updates) == 0 {
		return
	}

	updateCtx, cancel := newGrokAccountStateContext(ctx)
	defer cancel()

	if err := s.accountRepo.UpdateExtra(updateCtx, account.ID, updates); err != nil {
		return
	}
	mergeAccountExtra(account, updates)
}

func newGrokAccountStateContext(ctx context.Context) (context.Context, context.CancelFunc) {
	base := context.Background()
	if ctx != nil {
		base = context.WithoutCancel(ctx)
	}
	return context.WithTimeout(base, grokAccountStatePersistTimeout)
}
