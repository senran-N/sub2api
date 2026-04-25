package service

import "context"

type RuntimeAdmissionOutcome string

const (
	RuntimeAdmissionSucceeded        RuntimeAdmissionOutcome = "succeeded"
	RuntimeAdmissionRPMDenied        RuntimeAdmissionOutcome = "rpm_denied"
	RuntimeAdmissionWindowCostDenied RuntimeAdmissionOutcome = "window_cost_denied"
)

type RuntimeRPMAdmissionRequest struct {
	Account              *Account
	StickyBoundAccountID int64
	GroupID              *int64
	SessionHash          string
}

type RuntimeRPMAdmissionResult struct {
	Outcome        RuntimeAdmissionOutcome
	Account        *Account
	StickyBound    bool
	RPMCount       int
	Err            error
	ClearStickyErr error
}

type RuntimeWindowCostAdmissionRequest struct {
	Account              *Account
	APIKey               *APIKey
	Parsed               *ParsedRequest
	BillingModel         string
	StickyBoundAccountID int64
	GroupID              *int64
	SessionHash          string
}

type RuntimeWindowCostAdmissionResult struct {
	Outcome        RuntimeAdmissionOutcome
	Account        *Account
	StickyBound    bool
	Reservation    *WindowCostReservation
	Total          float64
	Err            error
	ClearStickyErr error
}

func (s *GatewayService) ReserveRuntimeRPMAdmission(ctx context.Context, req RuntimeRPMAdmissionRequest) RuntimeRPMAdmissionResult {
	result := RuntimeRPMAdmissionResult{
		Outcome: RuntimeAdmissionSucceeded,
		Account: req.Account,
	}
	if s == nil || req.Account == nil {
		return result
	}

	result.StickyBound = isRuntimeAdmissionStickyBound(req.Account, req.StickyBoundAccountID)
	allowed, count, err := s.TryReserveAccountRPM(ctx, req.Account, result.StickyBound)
	result.RPMCount = count
	if err != nil {
		result.Err = err
		return result
	}
	if allowed {
		return result
	}

	result.Outcome = RuntimeAdmissionRPMDenied
	result.ClearStickyErr = s.ClearStickySessionBinding(ctx, req.GroupID, req.SessionHash)
	return result
}

func (s *GatewayService) ReserveRuntimeWindowCostAdmission(ctx context.Context, req RuntimeWindowCostAdmissionRequest) RuntimeWindowCostAdmissionResult {
	result := RuntimeWindowCostAdmissionResult{
		Outcome: RuntimeAdmissionSucceeded,
		Account: req.Account,
	}
	if s == nil || req.Account == nil {
		return result
	}

	result.StickyBound = isRuntimeAdmissionStickyBound(req.Account, req.StickyBoundAccountID)
	reservation, allowed, total, err := s.TryReserveAccountWindowCost(
		ctx,
		req.Account,
		req.APIKey,
		req.Parsed,
		req.BillingModel,
		result.StickyBound,
	)
	result.Reservation = reservation
	result.Total = total
	if err != nil {
		result.Err = err
		return result
	}
	if allowed {
		return result
	}

	result.Outcome = RuntimeAdmissionWindowCostDenied
	result.ClearStickyErr = s.ClearStickySessionBinding(ctx, req.GroupID, req.SessionHash)
	return result
}

func isRuntimeAdmissionStickyBound(account *Account, stickyBoundAccountID int64) bool {
	return account != nil && stickyBoundAccountID > 0 && stickyBoundAccountID == account.ID
}
