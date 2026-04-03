package service

import "context"

// SelectAccountWithLoadAwareness selects account with load-awareness and wait plan.
// metadataUserID: 已废弃参数，会话限制现在统一使用 sessionHash
func (s *GatewayService) SelectAccountWithLoadAwareness(ctx context.Context, groupID *int64, sessionHash string, requestedModel string, excludedIDs map[int64]struct{}, metadataUserID string) (*AccountSelectionResult, error) {
	logLoadAwareSelectionStart(ctx, groupID, sessionHash, requestedModel, excludedIDs)
	cfg := s.schedulingConfig()

	scope, err := s.prepareLoadAwareSelectionScope(ctx, groupID, sessionHash)
	if err != nil {
		return nil, err
	}
	ctx = scope.ctx
	group := scope.group
	groupID = scope.groupID
	stickyAccountID := scope.stickyAccountID

	s.logLoadAwareSelectionEntry(group, groupID, sessionHash, requestedModel, stickyAccountID, cfg)

	if s.concurrencyService == nil || !cfg.LoadBatchEnabled {
		return s.selectAccountWithoutLoadBatch(ctx, groupID, sessionHash, requestedModel, excludedIDs, stickyAccountID, cfg)
	}

	schedulingState, err := s.prepareLoadAwareSchedulingState(ctx, groupID, group, requestedModel, sessionHash, stickyAccountID)
	if err != nil {
		return nil, err
	}

	return s.selectAccountWithLoadBatch(&loadAwareBatchSelectionInput{
		ctx:              schedulingState.ctx,
		groupID:          groupID,
		sessionHash:      sessionHash,
		requestedModel:   requestedModel,
		excludedIDs:      excludedIDs,
		stickyAccountID:  stickyAccountID,
		schedulingConfig: cfg,
		schedulingState:  schedulingState,
	})
}
