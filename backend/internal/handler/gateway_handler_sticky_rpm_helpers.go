package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/service"
	"go.uber.org/zap"
)

func (h *GatewayHandler) prefetchStickySessionBinding(c *gin.Context, groupID *int64, sessionKey string) int64 {
	if c == nil || h == nil || h.gatewayService == nil || sessionKey == "" {
		return 0
	}

	accountID, _ := h.gatewayService.GetCachedSessionAccountID(c.Request.Context(), groupID, sessionKey)
	if accountID <= 0 {
		return 0
	}

	prefetchedGroupID := int64(0)
	if groupID != nil {
		prefetchedGroupID = *groupID
	}
	ctx := service.WithPrefetchedStickySession(c.Request.Context(), accountID, prefetchedGroupID, h.metadataBridgeEnabled())
	c.Request = c.Request.WithContext(ctx)
	return accountID
}

func (h *GatewayHandler) tryReserveAccountRPMForForward(
	ctx context.Context,
	reqLog *zap.Logger,
	account *service.Account,
	stickyBoundAccountID int64,
	groupID *int64,
	sessionKey string,
) bool {
	if h == nil || h.gatewayService == nil || account == nil {
		return true
	}

	allowStickyOverflow := stickyBoundAccountID > 0 && stickyBoundAccountID == account.ID
	allowed, count, err := h.gatewayService.TryReserveAccountRPM(ctx, account, allowStickyOverflow)
	if err != nil {
		if reqLog != nil {
			reqLog.Warn("gateway.rpm_reserve_failed",
				zap.Int64("account_id", account.ID),
				zap.Error(err),
			)
		}
		return true
	}
	if allowed {
		return true
	}

	if reqLog != nil {
		reqLog.Info("gateway.rpm_reservation_denied",
			zap.Int64("account_id", account.ID),
			zap.Int("rpm_count", count),
			zap.Int("base_rpm", account.GetBaseRPM()),
			zap.Bool("sticky_bound", allowStickyOverflow),
		)
	}
	if err := h.gatewayService.ClearStickySessionBinding(ctx, groupID, sessionKey); err != nil && reqLog != nil {
		reqLog.Warn("gateway.clear_sticky_session_after_rpm_deny_failed",
			zap.Int64("account_id", account.ID),
			zap.Error(err),
		)
	}
	return false
}

func (h *GatewayHandler) tryReserveAccountWindowCostForForward(
	ctx context.Context,
	reqLog *zap.Logger,
	account *service.Account,
	apiKey *service.APIKey,
	parsed *service.ParsedRequest,
	billingModel string,
	stickyBoundAccountID int64,
	groupID *int64,
	sessionKey string,
) (*service.WindowCostReservation, bool) {
	if h == nil || h.gatewayService == nil || account == nil {
		return nil, true
	}

	allowStickyOverflow := stickyBoundAccountID > 0 && stickyBoundAccountID == account.ID
	reservation, allowed, total, err := h.gatewayService.TryReserveAccountWindowCost(
		ctx,
		account,
		apiKey,
		parsed,
		billingModel,
		allowStickyOverflow,
	)
	if err != nil {
		if reqLog != nil {
			reqLog.Warn("gateway.window_cost_reserve_failed",
				zap.Int64("account_id", account.ID),
				zap.Error(err),
			)
		}
		return nil, true
	}
	if allowed {
		return reservation, true
	}

	if reqLog != nil {
		reqLog.Info("gateway.window_cost_reservation_denied",
			zap.Int64("account_id", account.ID),
			zap.Float64("window_cost_total", total),
			zap.Float64("window_cost_limit", account.GetWindowCostLimit()),
			zap.Bool("sticky_bound", allowStickyOverflow),
		)
	}
	if err := h.gatewayService.ClearStickySessionBinding(ctx, groupID, sessionKey); err != nil && reqLog != nil {
		reqLog.Warn("gateway.clear_sticky_session_after_window_cost_deny_failed",
			zap.Int64("account_id", account.ID),
			zap.Error(err),
		)
	}
	return nil, false
}

func (h *GatewayHandler) releaseWindowCostReservation(ctx context.Context, reqLog *zap.Logger, reservation *service.WindowCostReservation) {
	if h == nil || h.gatewayService == nil || reservation == nil {
		return
	}
	if err := h.gatewayService.ReleaseWindowCostReservation(ctx, reservation); err != nil && reqLog != nil {
		reqLog.Warn("gateway.window_cost_release_failed",
			zap.Int64("account_id", reservation.AccountID),
			zap.Error(err),
		)
	}
}
