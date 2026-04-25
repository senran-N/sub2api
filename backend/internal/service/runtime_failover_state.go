package service

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/logger"
	"go.uber.org/zap"
)

type RuntimeFailoverAction int

const (
	RuntimeFailoverContinue RuntimeFailoverAction = iota
	RuntimeFailoverExhausted
	RuntimeFailoverCanceled
)

// TempUnscheduler temporarily removes accounts after same-account retries are exhausted.
type TempUnscheduler interface {
	TempUnscheduleRetryableError(ctx context.Context, accountID int64, failoverErr *UpstreamFailoverError)
}

const (
	maxSameAccountRetries = 3
	sameAccountRetryDelay = 500 * time.Millisecond

	// Single-account Antigravity groups already do bounded in-place retries in service code.
	// This delay only spaces the next selection pass after all candidates were excluded.
	singleAccountBackoffDelay = 2 * time.Second
)

// RuntimeFailoverState carries failover state across gateway forwarding attempts.
type RuntimeFailoverState struct {
	SwitchCount           int
	MaxSwitches           int
	FailedAccountIDs      map[int64]struct{}
	SameAccountRetryCount map[int64]int
	LastFailoverErr       *UpstreamFailoverError
	ForceCacheBilling     bool
	hasBoundSession       bool
}

type RuntimeForwardFailoverResult struct {
	Handled         bool
	Action          RuntimeFailoverAction
	FailoverErr     *UpstreamFailoverError
	ResponseStarted bool
}

type RuntimeSelectionFailureOutcome string

const (
	RuntimeSelectionFailureInitialUnavailable RuntimeSelectionFailureOutcome = "initial_unavailable"
	RuntimeSelectionFailureRetry              RuntimeSelectionFailureOutcome = "retry"
	RuntimeSelectionFailureCanceled           RuntimeSelectionFailureOutcome = "canceled"
	RuntimeSelectionFailureExhausted          RuntimeSelectionFailureOutcome = "exhausted"
)

type RuntimeSelectionFailureResult struct {
	Handled     bool
	Outcome     RuntimeSelectionFailureOutcome
	Action      RuntimeFailoverAction
	Err         error
	FailoverErr *UpstreamFailoverError
}

func NewRuntimeFailoverState(maxSwitches int, hasBoundSession bool) *RuntimeFailoverState {
	return &RuntimeFailoverState{
		MaxSwitches:           maxSwitches,
		FailedAccountIDs:      make(map[int64]struct{}),
		SameAccountRetryCount: make(map[int64]int),
		hasBoundSession:       hasBoundSession,
	}
}

func (s *RuntimeFailoverState) HandleSelectionError(ctx context.Context, err error) RuntimeSelectionFailureResult {
	if err == nil {
		return RuntimeSelectionFailureResult{}
	}

	result := RuntimeSelectionFailureResult{
		Handled: true,
		Err:     err,
	}
	if s == nil || len(s.FailedAccountIDs) == 0 {
		result.Outcome = RuntimeSelectionFailureInitialUnavailable
		result.Action = RuntimeFailoverExhausted
		return result
	}

	result.Action = s.HandleSelectionExhausted(ctx)
	result.FailoverErr = s.LastFailoverErr
	switch result.Action {
	case RuntimeFailoverContinue:
		result.Outcome = RuntimeSelectionFailureRetry
	case RuntimeFailoverCanceled:
		result.Outcome = RuntimeSelectionFailureCanceled
	default:
		result.Outcome = RuntimeSelectionFailureExhausted
	}
	return result
}

func (s *RuntimeFailoverState) HandleForwardError(
	ctx context.Context,
	gatewayService TempUnscheduler,
	accountID int64,
	platform string,
	err error,
	responseStarted bool,
) RuntimeForwardFailoverResult {
	var failoverErr *UpstreamFailoverError
	if !errors.As(err, &failoverErr) {
		return RuntimeForwardFailoverResult{}
	}

	result := RuntimeForwardFailoverResult{
		Handled:         true,
		FailoverErr:     failoverErr,
		ResponseStarted: responseStarted,
	}
	if responseStarted || s == nil {
		result.Action = RuntimeFailoverExhausted
		return result
	}

	result.Action = s.HandleFailoverError(ctx, gatewayService, accountID, platform, failoverErr)
	result.FailoverErr = s.LastFailoverErr
	return result
}

func (s *RuntimeFailoverState) HandleFailoverError(
	ctx context.Context,
	gatewayService TempUnscheduler,
	accountID int64,
	platform string,
	failoverErr *UpstreamFailoverError,
) RuntimeFailoverAction {
	s.LastFailoverErr = failoverErr

	if needForceCacheBilling(s.hasBoundSession, failoverErr) {
		s.ForceCacheBilling = true
	}

	if shouldExhaustFailoverImmediately(failoverErr, s.hasBoundSession) {
		s.FailedAccountIDs[accountID] = struct{}{}
		return RuntimeFailoverExhausted
	}

	if failoverErr.RetryableOnSameAccount && s.SameAccountRetryCount[accountID] < maxSameAccountRetries {
		s.SameAccountRetryCount[accountID]++
		logger.FromContext(ctx).Warn("gateway.failover_same_account_retry",
			zap.Int64("account_id", accountID),
			zap.Int("upstream_status", failoverErr.StatusCode),
			zap.Int("same_account_retry_count", s.SameAccountRetryCount[accountID]),
			zap.Int("same_account_retry_max", maxSameAccountRetries),
		)
		if err := sleepWithContext(ctx, sameAccountRetryDelay); err != nil {
			return RuntimeFailoverCanceled
		}
		return RuntimeFailoverContinue
	}

	if shouldPreserveBoundSessionOnRateLimit(failoverErr, s.hasBoundSession) {
		s.FailedAccountIDs[accountID] = struct{}{}
		return RuntimeFailoverExhausted
	}

	if failoverErr.RetryableOnSameAccount {
		gatewayService.TempUnscheduleRetryableError(ctx, accountID, failoverErr)
	}

	s.FailedAccountIDs[accountID] = struct{}{}

	if s.SwitchCount >= s.MaxSwitches {
		return RuntimeFailoverExhausted
	}

	s.SwitchCount++
	logger.FromContext(ctx).Warn("gateway.failover_switch_account",
		zap.Int64("account_id", accountID),
		zap.Int("upstream_status", failoverErr.StatusCode),
		zap.Int("switch_count", s.SwitchCount),
		zap.Int("max_switches", s.MaxSwitches),
	)

	if platform == PlatformAntigravity {
		delay := time.Duration(s.SwitchCount-1) * time.Second
		if err := sleepWithContext(ctx, delay); err != nil {
			return RuntimeFailoverCanceled
		}
	}

	return RuntimeFailoverContinue
}

// HandleSelectionExhausted decides whether a single-account 503 can re-enter selection.
func (s *RuntimeFailoverState) HandleSelectionExhausted(ctx context.Context) RuntimeFailoverAction {
	if s.LastFailoverErr != nil &&
		s.LastFailoverErr.StatusCode == http.StatusServiceUnavailable &&
		s.SwitchCount <= s.MaxSwitches {

		logger.FromContext(ctx).Warn("gateway.failover_single_account_backoff",
			zap.Duration("backoff_delay", singleAccountBackoffDelay),
			zap.Int("switch_count", s.SwitchCount),
			zap.Int("max_switches", s.MaxSwitches),
		)
		if err := sleepWithContext(ctx, singleAccountBackoffDelay); err != nil {
			return RuntimeFailoverCanceled
		}
		logger.FromContext(ctx).Warn("gateway.failover_single_account_retry",
			zap.Int("switch_count", s.SwitchCount),
			zap.Int("max_switches", s.MaxSwitches),
		)
		s.FailedAccountIDs = make(map[int64]struct{})
		return RuntimeFailoverContinue
	}
	return RuntimeFailoverExhausted
}

func needForceCacheBilling(hasBoundSession bool, failoverErr *UpstreamFailoverError) bool {
	return hasBoundSession || (failoverErr != nil && failoverErr.ForceCacheBilling)
}

func shouldExhaustFailoverImmediately(failoverErr *UpstreamFailoverError, hasBoundSession bool) bool {
	if failoverErr == nil {
		return false
	}
	switch failoverErr.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return true
	case http.StatusTooManyRequests:
		return !hasBoundSession
	default:
		return false
	}
}

func shouldPreserveBoundSessionOnRateLimit(failoverErr *UpstreamFailoverError, hasBoundSession bool) bool {
	return hasBoundSession && failoverErr != nil && failoverErr.StatusCode == http.StatusTooManyRequests
}
