package service

import (
	"context"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/logger"
)

func stickyAccountCurrentRPM(ctx context.Context, accountID int64) int {
	if count, ok := rpmFromPrefetchContext(ctx, accountID); ok {
		return count
	}
	return 0
}

func (s *GatewayService) logStickyCacheMiss(account *Account, accountID int64, sessionHash, reason string, currentRPM int) {
	baseRPM := 0
	if account != nil {
		baseRPM = account.GetBaseRPM()
	}
	logger.LegacyPrintf("service.gateway", "[StickyCacheMiss] reason=%s account_id=%d session=%s current_rpm=%d base_rpm=%d",
		reason, accountID, shortSessionHash(sessionHash), currentRPM, baseRPM)
}

func (s *GatewayService) clearMissingStickyBinding(ctx context.Context, groupID *int64, sessionHash string, stickyAccountID int64) {
	if s.cache != nil {
		_ = s.cache.DeleteSessionAccountID(ctx, derefGroupID(groupID), sessionHash)
	}
	s.logStickyCacheMiss(nil, stickyAccountID, sessionHash, "account_cleared", 0)
}

func (s *GatewayService) trySelectRoutedStickyAccount(
	ctx context.Context,
	groupID *int64,
	sessionHash string,
	requestedModel string,
	stickyAccountID int64,
	routingAccountIDs []int64,
	excludedIDs map[int64]struct{},
	accountByID map[int64]*Account,
	platform string,
	useMixed bool,
	waitTimeout time.Duration,
	maxWaiting int,
) (*AccountSelectionResult, bool) {
	if sessionHash == "" || stickyAccountID <= 0 {
		return nil, false
	}
	if _, excluded := excludedIDs[stickyAccountID]; excluded {
		return nil, false
	}
	if !containsInt64(routingAccountIDs, stickyAccountID) {
		return nil, false
	}

	stickyAccount, ok := s.resolveSelectionAccountByID(ctx, accountByID, stickyAccountID)
	if !ok {
		s.clearMissingStickyBinding(ctx, groupID, sessionHash, stickyAccountID)
		return nil, false
	}

	gatePass := s.isAccountSchedulableForSelection(stickyAccount) &&
		s.isAccountAllowedForPlatform(stickyAccount, platform, useMixed) &&
		(requestedModel == "" || s.isModelSupportedByAccountWithContext(ctx, stickyAccount, requestedModel)) &&
		!s.isChannelModelRestrictedForSelectionWithGroup(ctx, groupID, stickyAccount, requestedModel) &&
		s.isAccountSchedulableForModelSelection(ctx, stickyAccount, requestedModel) &&
		s.isAccountSchedulableForQuota(stickyAccount) &&
		s.isAccountSchedulableForWindowCost(ctx, stickyAccount, true)

	rpmPass := gatePass && s.isAccountSchedulableForRPM(ctx, stickyAccount, true)
	stickyCacheMissReason := ""

	if rpmPass {
		result, missReason, ok := s.trySelectResolvedStickyAccount(ctx, nil, sessionHash, stickyAccount, false, waitTimeout, maxWaiting)
		if ok {
			if s.debugModelRoutingEnabled() {
				logger.LegacyPrintf("service.gateway", "[ModelRoutingDebug] routed sticky hit: group_id=%v model=%s session=%s account=%d",
					derefGroupID(groupID), requestedModel, shortSessionHash(sessionHash), stickyAccountID)
			}
			return result, true
		}
		stickyCacheMissReason = missReason
	} else if !gatePass {
		stickyCacheMissReason = "gate_check"
	} else {
		stickyCacheMissReason = "rpm_red"
	}

	if stickyCacheMissReason != "" {
		s.logStickyCacheMiss(stickyAccount, stickyAccountID, sessionHash, stickyCacheMissReason, stickyAccountCurrentRPM(ctx, stickyAccount.ID))
	}

	return nil, false
}
