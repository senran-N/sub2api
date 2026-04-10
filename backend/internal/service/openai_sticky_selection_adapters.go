package service

import (
	"context"
	"time"

	"github.com/senran-N/sub2api/internal/config"
)

func newOpenAIAcquiredSelection(account *Account, acquired *AcquireResult) *AccountSelectionResult {
	if acquired == nil {
		return newAcquiredAccountSelection(account, nil)
	}
	return newAcquiredAccountSelection(account, acquired.ReleaseFunc)
}

func (s *OpenAIGatewayService) buildOpenAIStickyWaitPlanAdapter(
	cfg config.GatewaySchedulingConfig,
) func(account *Account) (*AccountSelectionResult, bool) {
	return func(account *Account) (*AccountSelectionResult, bool) {
		return buildStickySessionWaitPlanIfConcurrencyEnabled(account, cfg, s.concurrencyService)
	}
}

func (s *OpenAIGatewayService) buildOpenAIStickyTTLSelectionAdapter(
	ctx context.Context,
	groupID *int64,
	sessionHash string,
	ttl time.Duration,
) func(account *Account, acquired *AcquireResult) *AccountSelectionResult {
	return func(account *Account, acquired *AcquireResult) *AccountSelectionResult {
		_ = s.refreshStickySessionTTL(ctx, groupID, sessionHash, ttl)
		return newOpenAIAcquiredSelection(account, acquired)
	}
}

func (s *OpenAIGatewayService) buildOpenAIResponseBindingSelectionAdapter(
	ctx context.Context,
	groupID *int64,
	responseID string,
	accountID int64,
	store OpenAIWSStateStore,
) func(account *Account, acquired *AcquireResult) *AccountSelectionResult {
	return func(account *Account, acquired *AcquireResult) *AccountSelectionResult {
		logOpenAIWSBindResponseAccountWarn(
			derefGroupID(groupID),
			accountID,
			responseID,
			store.BindResponseAccount(ctx, derefGroupID(groupID), responseID, accountID, s.openAIWSResponseStickyTTL()),
		)
		return newOpenAIAcquiredSelection(account, acquired)
	}
}
