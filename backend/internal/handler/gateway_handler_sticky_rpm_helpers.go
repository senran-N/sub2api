package handler

import (
	"context"

	"github.com/senran-N/sub2api/internal/service"
	"go.uber.org/zap"
)

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

	reservation := h.gatewayService.ReserveRuntimeRPMAdmission(ctx, service.RuntimeRPMAdmissionRequest{
		Account:              account,
		StickyBoundAccountID: stickyBoundAccountID,
		GroupID:              groupID,
		SessionHash:          sessionKey,
	})
	return logRuntimeRPMAdmissionResult(reqLog, account, reservation)
}

func logRuntimeRPMAdmissionResult(reqLog *zap.Logger, account *service.Account, reservation service.RuntimeRPMAdmissionResult) bool {
	if account == nil {
		return true
	}

	if reservation.Err != nil {
		if reqLog != nil {
			reqLog.Warn("gateway.rpm_reserve_failed",
				zap.Int64("account_id", account.ID),
				zap.Error(reservation.Err),
			)
		}
		return true
	}
	if reservation.Outcome == service.RuntimeAdmissionSucceeded {
		return true
	}
	if reqLog != nil {
		reqLog.Info("gateway.rpm_reservation_denied",
			zap.Int64("account_id", account.ID),
			zap.Int("rpm_count", reservation.RPMCount),
			zap.Int("base_rpm", account.GetBaseRPM()),
			zap.Bool("sticky_bound", reservation.StickyBound),
		)
	}
	if reservation.ClearStickyErr != nil && reqLog != nil {
		reqLog.Warn("gateway.clear_sticky_session_after_rpm_deny_failed",
			zap.Int64("account_id", account.ID),
			zap.Error(reservation.ClearStickyErr),
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

	admission := h.gatewayService.ReserveRuntimeWindowCostAdmission(ctx, service.RuntimeWindowCostAdmissionRequest{
		Account:              account,
		APIKey:               apiKey,
		Parsed:               parsed,
		BillingModel:         billingModel,
		StickyBoundAccountID: stickyBoundAccountID,
		GroupID:              groupID,
		SessionHash:          sessionKey,
	})
	if admission.Err != nil {
		if reqLog != nil {
			reqLog.Warn("gateway.window_cost_reserve_failed",
				zap.Int64("account_id", account.ID),
				zap.Error(admission.Err),
			)
		}
		return nil, true
	}
	if admission.Outcome == service.RuntimeAdmissionSucceeded {
		return admission.Reservation, true
	}

	if reqLog != nil {
		reqLog.Info("gateway.window_cost_reservation_denied",
			zap.Int64("account_id", account.ID),
			zap.Float64("window_cost_total", admission.Total),
			zap.Float64("window_cost_limit", account.GetWindowCostLimit()),
			zap.Bool("sticky_bound", admission.StickyBound),
		)
	}
	if admission.ClearStickyErr != nil && reqLog != nil {
		reqLog.Warn("gateway.clear_sticky_session_after_window_cost_deny_failed",
			zap.Int64("account_id", account.ID),
			zap.Error(admission.ClearStickyErr),
		)
	}
	return nil, false
}

func (h *GatewayHandler) executeRuntimeForwardAttempt(
	ctx context.Context,
	reqLog *zap.Logger,
	req service.RuntimeForwardAttemptRequest,
) service.RuntimeForwardAttemptResult {
	if h != nil && h.gatewayService != nil && req.ReleaseWindowCost == nil {
		req.ReleaseWindowCost = h.gatewayService.ReleaseWindowCostReservation
	}
	attempt := service.ExecuteRuntimeForwardAttempt(ctx, req)
	h.logWindowCostReleaseError(reqLog, req.WindowCostReservation, attempt.WindowCostReleaseErr)
	return attempt
}

func (h *GatewayHandler) logWindowCostReleaseError(reqLog *zap.Logger, reservation *service.WindowCostReservation, err error) {
	if err != nil && reqLog != nil && reservation != nil {
		reqLog.Warn("gateway.window_cost_release_failed",
			zap.Int64("account_id", reservation.AccountID),
			zap.Error(err),
		)
	}
}
