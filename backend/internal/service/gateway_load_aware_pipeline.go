package service

import (
	"context"
	"log/slog"
	"sort"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

type loadAwareBatchSelectionInput struct {
	ctx              context.Context
	groupID          *int64
	sessionHash      string
	requestedModel   string
	excludedIDs      map[int64]struct{}
	stickyAccountID  int64
	schedulingConfig config.GatewaySchedulingConfig
	schedulingState  *loadAwareSchedulingState
}

func collectExcludedAccountIDs(excludedIDs map[int64]struct{}) []int64 {
	excludedIDsList := make([]int64, 0, len(excludedIDs))
	for accountID := range excludedIDs {
		excludedIDsList = append(excludedIDsList, accountID)
	}
	sort.Slice(excludedIDsList, func(i, j int) bool {
		return excludedIDsList[i] < excludedIDsList[j]
	})
	return excludedIDsList
}

func logLoadAwareSelectionStart(ctx context.Context, groupID *int64, sessionHash string, requestedModel string, excludedIDs map[int64]struct{}) {
	if !slog.Default().Enabled(ctx, slog.LevelDebug) {
		return
	}

	slog.Debug("account_scheduling_starting",
		"group_id", derefGroupID(groupID),
		"model", requestedModel,
		"session", shortSessionHash(sessionHash),
		"excluded_ids", collectExcludedAccountIDs(excludedIDs))
}

func (s *GatewayService) logLoadAwareSelectionEntry(
	group *Group,
	groupID *int64,
	sessionHash string,
	requestedModel string,
	stickyAccountID int64,
	cfg config.GatewaySchedulingConfig,
) {
	if !s.debugModelRoutingEnabled() || requestedModel == "" {
		return
	}

	groupPlatform := ""
	if group != nil {
		groupPlatform = group.Platform
	}
	logger.LegacyPrintf("service.gateway", "[ModelRoutingDebug] select entry: group_id=%v group_platform=%s model=%s session=%s sticky_account=%d load_batch=%v concurrency=%v",
		derefGroupID(groupID), groupPlatform, requestedModel, shortSessionHash(sessionHash), stickyAccountID, cfg.LoadBatchEnabled, s.concurrencyService != nil)
}

func (s *GatewayService) selectAccountWithLoadBatch(input *loadAwareBatchSelectionInput) (*AccountSelectionResult, error) {
	state := input.schedulingState
	if state == nil || state.plan == nil {
		return nil, ErrNoAvailableAccounts
	}
	plan := state.plan

	if result, ok := s.trySelectLoadAwareRoutedAccount(&loadAwareRoutedSelectionInput{
		ctx:             input.ctx,
		groupID:         input.groupID,
		sessionHash:     input.sessionHash,
		requestedModel:  input.requestedModel,
		stickyAccountID: input.stickyAccountID,
		excludedIDs:     input.excludedIDs,
		accountByID:     state.accountByID,
		plan:            plan,
		waitTimeout:     input.schedulingConfig.StickySessionWaitTimeout,
		maxWaiting:      input.schedulingConfig.StickySessionMaxWaiting,
	}); ok {
		return result, nil
	}

	return s.selectLoadAwareFallbackFlow(&loadAwareFallbackSelectionInput{
		ctx:              input.ctx,
		groupID:          input.groupID,
		sessionHash:      input.sessionHash,
		requestedModel:   input.requestedModel,
		stickyAccountID:  input.stickyAccountID,
		excludedIDs:      input.excludedIDs,
		accountByID:      state.accountByID,
		accounts:         state.accounts,
		plan:             plan,
		schedulingConfig: input.schedulingConfig,
	})
}
