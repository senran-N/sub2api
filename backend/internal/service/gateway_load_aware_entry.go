package service

import "context"

// SelectAccountWithLoadAwareness selects account with load-awareness and wait plan.
// metadataUserID: 已废弃参数，会话限制现在统一使用 sessionHash
func (s *GatewayService) SelectAccountWithLoadAwareness(ctx context.Context, groupID *int64, sessionHash string, requestedModel string, excludedIDs map[int64]struct{}, metadataUserID string) (*AccountSelectionResult, error) {
	logLoadAwareSelectionStart(ctx, groupID, sessionHash, requestedModel, excludedIDs)
	cfg := gatewaySchedulingConfigOrDefault(s.cfg)

	scope, err := s.prepareGatewaySelectionScope(ctx, groupID, sessionHash)
	if err != nil {
		return nil, err
	}
	ctx = scope.ctx
	group := scope.group
	groupID = scope.groupID
	stickyAccountID := scope.stickyAccountID
	plan := s.buildGatewaySelectionPlan(scope, requestedModel)

	s.logLoadAwareSelectionEntry(group, groupID, sessionHash, requestedModel, stickyAccountID, cfg)

	if s.concurrencyService == nil || !cfg.LoadBatchEnabled {
		return s.selectAccountWithoutLoadBatch(scope, plan, sessionHash, requestedModel, excludedIDs, cfg)
	}

	schedulingState, err := s.buildLoadAwareSchedulingState(scope, plan, requestedModel, sessionHash)
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
