package service

import (
	"context"
	"strings"
)

// SelectAccountByPreviousResponseID 按 previous_response_id 命中账号粘连。
// 未命中或账号不可用时返回 (nil, nil)，由调用方继续走常规调度。
func (s *OpenAIGatewayService) SelectAccountByPreviousResponseID(
	ctx context.Context,
	groupID *int64,
	previousResponseID string,
	requestedModel string,
	excludedIDs map[int64]struct{},
) (*AccountSelectionResult, error) {
	if s == nil {
		return nil, nil
	}
	responseID := strings.TrimSpace(previousResponseID)
	if responseID == "" {
		return nil, nil
	}
	store := s.getOpenAIWSStateStore()
	if store == nil {
		return nil, nil
	}

	accountID, err := store.GetResponseAccount(ctx, derefGroupID(groupID), responseID)
	if err != nil || accountID <= 0 {
		return nil, nil
	}
	if excludedIDs != nil {
		if _, excluded := excludedIDs[accountID]; excluded {
			return nil, nil
		}
	}

	account, err := s.getSchedulableAccount(ctx, accountID)
	if err != nil || account == nil {
		_ = store.DeleteResponseAccount(ctx, derefGroupID(groupID), responseID)
		return nil, nil
	}
	// 非 WSv2 场景（如 force_http/全局关闭）不应使用 previous_response_id 粘连，
	// 以保持“回滚到 HTTP”后的历史行为一致性。
	if s.getOpenAIWSProtocolResolver().Resolve(account).Transport != OpenAIUpstreamTransportResponsesWebsocketV2 {
		return nil, nil
	}
	if shouldClearStickySession(account, requestedModel) || !account.IsOpenAI() || !account.IsSchedulable() {
		_ = store.DeleteResponseAccount(ctx, derefGroupID(groupID), responseID)
		return nil, nil
	}
	if requestedModel != "" && !account.IsModelSupported(requestedModel) {
		return nil, nil
	}
	account = s.recheckSelectedOpenAIAccountFromDB(ctx, account, requestedModel)
	if account == nil {
		_ = store.DeleteResponseAccount(ctx, derefGroupID(groupID), responseID)
		return nil, nil
	}

	result, acquireErr := s.tryAcquireAccountSlot(ctx, accountID, account.Concurrency)
	if acquireErr == nil && result.Acquired {
		logOpenAIWSBindResponseAccountWarn(
			derefGroupID(groupID),
			accountID,
			responseID,
			store.BindResponseAccount(ctx, derefGroupID(groupID), responseID, accountID, s.openAIWSResponseStickyTTL()),
		)
		return &AccountSelectionResult{
			Account:     account,
			Acquired:    true,
			ReleaseFunc: result.ReleaseFunc,
		}, nil
	}

	cfg := s.schedulingConfig()
	if s.concurrencyService != nil {
		return &AccountSelectionResult{
			Account: account,
			WaitPlan: &AccountWaitPlan{
				AccountID:      accountID,
				MaxConcurrency: account.Concurrency,
				Timeout:        cfg.StickySessionWaitTimeout,
				MaxWaiting:     cfg.StickySessionMaxWaiting,
			},
		}, nil
	}
	return nil, nil
}
