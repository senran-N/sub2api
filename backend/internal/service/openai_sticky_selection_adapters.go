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
	store OpenAIWSStateStore,
) func(account *Account, acquired *AcquireResult) *AccountSelectionResult {
	return func(account *Account, acquired *AcquireResult) *AccountSelectionResult {
		boundAccountID := int64(0)
		if account != nil {
			boundAccountID = account.ID
		}
		s.bindCodexChainSuccess(ctx, store, codexChainBinding{
			AccountID:   boundAccountID,
			GroupID:     derefGroupID(groupID),
			ResponseID:  responseID,
			ResponseTTL: s.openAIWSResponseStickyTTL(),
			Transport:   OpenAIUpstreamTransportResponsesWebsocketV2,
		})
		return newOpenAIAcquiredSelection(account, acquired)
	}
}
