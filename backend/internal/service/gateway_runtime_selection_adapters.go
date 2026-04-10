package service

import (
	"context"
	"time"
)

func (s *GatewayService) buildGatewayStickyBindAdapter(
	ctx context.Context,
	groupID *int64,
	sessionHash string,
	bindSticky bool,
) func(account *Account) {
	if !bindSticky {
		return nil
	}
	return func(account *Account) {
		s.bindStickySelection(ctx, groupID, sessionHash, account.ID)
	}
}

func (s *GatewayService) buildGatewayRuntimeAcquireSpec(
	ctx context.Context,
	account *Account,
	sessionHash string,
	finalize func(account *Account) *Account,
	bind func(account *Account),
) runtimeAcquireSelectionSpec {
	return runtimeAcquireSelectionSpec{
		account: account,
		acquire: func(account *Account) (*AcquireResult, error) {
			return acquireAccountSlotWithConcurrencyService(ctx, s.concurrencyService, account.ID, account.Concurrency)
		},
		finalize: finalize,
		allowSession: func(account *Account) bool {
			return s.checkAndRegisterSession(ctx, account, sessionHash)
		},
		bind: bind,
	}
}

func (s *GatewayService) buildGatewayRuntimeWaitPlanSpec(
	ctx context.Context,
	account *Account,
	sessionHash string,
	finalize func(account *Account) *Account,
	timeout time.Duration,
	maxWaiting int,
) runtimeWaitPlanSpec {
	return runtimeWaitPlanSpec{
		account: account,
		allowWait: func(account *Account) (string, bool) {
			if s.concurrencyService != nil {
				waitingCount, _ := s.concurrencyService.GetAccountWaitingCount(ctx, account.ID)
				if waitingCount >= maxWaiting {
					return "wait_queue_full", false
				}
			}
			return "", true
		},
		allowSession: func(account *Account) bool {
			return s.checkAndRegisterSession(ctx, account, sessionHash)
		},
		finalize:   finalize,
		timeout:    timeout,
		maxWaiting: maxWaiting,
	}
}
