package service

import (
	"context"
	"errors"
)

type RuntimeCountTokensSelectionRequest struct {
	GroupID              *int64
	SessionHash          string
	Model                string
	StickyBoundAccountID int64
}

type RuntimeCountTokensAdmissionEvent struct {
	Account   *Account
	Admission RuntimeRPMAdmissionResult
}

type RuntimeCountTokensSelectionResult struct {
	Account          *Account
	Err              error
	FailedAccountIDs map[int64]struct{}
	AdmissionEvents  []RuntimeCountTokensAdmissionEvent
}

type runtimeCountTokensSelectAccountFunc func(context.Context, *int64, string, string, map[int64]struct{}) (*Account, error)
type runtimeCountTokensReserveRPMFunc func(context.Context, RuntimeRPMAdmissionRequest) RuntimeRPMAdmissionResult

type runtimeCountTokensSelectionHooks struct {
	selectAccount runtimeCountTokensSelectAccountFunc
	reserveRPM    runtimeCountTokensReserveRPMFunc
}

func (s *GatewayService) SelectRuntimeCountTokensAccount(ctx context.Context, req RuntimeCountTokensSelectionRequest) RuntimeCountTokensSelectionResult {
	if s == nil {
		return RuntimeCountTokensSelectionResult{
			Err:              errors.New("runtime count_tokens selection: gateway service is not configured"),
			FailedAccountIDs: make(map[int64]struct{}),
		}
	}
	return selectRuntimeCountTokensAccount(ctx, req, runtimeCountTokensSelectionHooks{
		selectAccount: s.SelectAccountForModelWithExclusions,
		reserveRPM:    s.ReserveRuntimeRPMAdmission,
	})
}

func selectRuntimeCountTokensAccount(
	ctx context.Context,
	req RuntimeCountTokensSelectionRequest,
	hooks runtimeCountTokensSelectionHooks,
) RuntimeCountTokensSelectionResult {
	if ctx == nil {
		ctx = context.Background()
	}
	result := RuntimeCountTokensSelectionResult{
		FailedAccountIDs: make(map[int64]struct{}),
	}
	if hooks.selectAccount == nil || hooks.reserveRPM == nil {
		result.Err = errors.New("runtime count_tokens selection hooks are not configured")
		return result
	}

	for {
		account, err := hooks.selectAccount(ctx, req.GroupID, req.SessionHash, req.Model, result.FailedAccountIDs)
		if err != nil {
			result.Err = err
			return result
		}
		if account == nil {
			result.Err = errors.New("runtime count_tokens selection returned nil account")
			return result
		}

		admission := hooks.reserveRPM(ctx, RuntimeRPMAdmissionRequest{
			Account:              account,
			StickyBoundAccountID: req.StickyBoundAccountID,
			GroupID:              req.GroupID,
			SessionHash:          req.SessionHash,
		})
		if admission.Err != nil {
			result.AdmissionEvents = append(result.AdmissionEvents, RuntimeCountTokensAdmissionEvent{
				Account:   account,
				Admission: admission,
			})
			result.Account = account
			return result
		}
		if admission.Outcome == RuntimeAdmissionSucceeded {
			result.Account = account
			return result
		}

		result.AdmissionEvents = append(result.AdmissionEvents, RuntimeCountTokensAdmissionEvent{
			Account:   account,
			Admission: admission,
		})
		result.FailedAccountIDs[account.ID] = struct{}{}
	}
}
