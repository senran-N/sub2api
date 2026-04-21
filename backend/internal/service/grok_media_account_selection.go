package service

import (
	"context"
)

type grokAccountSelectionFilter func(*Account) bool

func selectSchedulableGrokMediaAccount(
	ctx context.Context,
	gatewayService *GatewayService,
	groupID *int64,
	requestedModel string,
	excludedIDs map[int64]struct{},
	filter grokAccountSelectionFilter,
	noAccountsErr string,
) (*Account, error) {
	ctx = WithGrokSessionMediaRuntimeAllowed(ctx)
	return selectSchedulableGrokAccount(
		ctx,
		gatewayService,
		groupID,
		requestedModel,
		excludedIDs,
		filter,
		firstNonEmpty(noAccountsErr, "no compatible grok media accounts"),
	)
}
