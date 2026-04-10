package service

import "context"

type openAISelectionDecisionSpec struct {
	layer             string
	stickyPreviousHit bool
	stickySessionHit  bool
	bindSession       bool
}

func openAISelectionAccount(selection *AccountSelectionResult) *Account {
	if selection == nil {
		return nil
	}
	return selection.Account
}

func (s *defaultOpenAIAccountScheduler) filterSchedulerSelectionByTransport(
	selection *AccountSelectionResult,
	requiredTransport OpenAIUpstreamTransport,
) *AccountSelectionResult {
	account := openAISelectionAccount(selection)
	if account == nil {
		return nil
	}
	if !s.isAccountTransportCompatible(account, requiredTransport) {
		return nil
	}
	return selection
}

func (s *defaultOpenAIAccountScheduler) applySchedulerSelectionDecision(
	ctx context.Context,
	req OpenAIAccountScheduleRequest,
	selection *AccountSelectionResult,
	spec openAISelectionDecisionSpec,
	decision *OpenAIAccountScheduleDecision,
) bool {
	account := openAISelectionAccount(selection)
	if account == nil || decision == nil {
		return false
	}

	decision.Layer = spec.layer
	decision.StickyPreviousHit = spec.stickyPreviousHit
	decision.StickySessionHit = spec.stickySessionHit
	decision.SelectedAccountID = account.ID
	decision.SelectedAccountType = account.Type
	if spec.bindSession && req.SessionHash != "" {
		_ = s.service.BindStickySession(ctx, req.GroupID, req.SessionHash, account.ID)
	}
	return true
}
