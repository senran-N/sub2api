package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/service"
)

func runtimeSlotAccountID(slot service.RuntimeAccountSlotResult) int64 {
	if slot.Account == nil {
		return 0
	}
	return slot.Account.ID
}

func runtimeSlotMaxWaiting(slot service.RuntimeAccountSlotResult) int {
	if slot.WaitPlan == nil {
		return 0
	}
	return slot.WaitPlan.MaxWaiting
}

func (h *GatewayHandler) acquireRuntimeAccountSlot(
	ctx context.Context,
	c *gin.Context,
	selection *service.AccountSelectionResult,
	groupID *int64,
	sessionHash string,
	reqStream bool,
	streamStarted *bool,
	bindSticky bool,
) service.RuntimeAccountSlotResult {
	var bindStickyFunc service.RuntimeAccountSlotBindStickyFunc
	if bindSticky {
		bindStickyFunc = h.gatewayService.BindStickySession
	}
	return service.AcquireRuntimeAccountSlot(ctx, service.RuntimeAccountSlotRequest{
		GroupID:        groupID,
		SessionHash:    sessionHash,
		Selection:      selection,
		AcquireOrQueue: h.concurrencyHelper.AcquireAccountSlotOrQueue,
		WaitForSlot: func(_ context.Context, wait service.RuntimeAccountSlotWaitRequest) (func(), error) {
			return h.concurrencyHelper.AcquireAccountSlotAfterQueueingWithWaitTimeout(
				c,
				wait.AccountID,
				wait.MaxConcurrency,
				wait.Timeout,
				reqStream,
				streamStarted,
			)
		},
		DecrementWait: func(ctx context.Context, accountID int64) {
			h.concurrencyHelper.DecrementAccountWaitCount(ctx, accountID)
		},
		BindSticky: bindStickyFunc,
	})
}

func (h *GatewayHandler) selectRuntimeAccount(ctx context.Context, request service.SelectionRequest) (*service.AccountSelectionResult, service.SelectionDecision, error) {
	return service.NewGatewaySelectionKernel(h.gatewayService).Select(ctx, request)
}
