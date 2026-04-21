package service

import (
	"context"
	"net/http"
	"time"
)

const grokAccountStatePersistTimeout = 5 * time.Second

type grokAccountStateExtraWriter interface {
	UpdateExtra(ctx context.Context, id int64, updates map[string]any) error
}

type grokAccountStateErrorWriter interface {
	SetError(ctx context.Context, id int64, errorMsg string) error
}

type grokAccountStateRecoveryWriter interface {
	ClearError(ctx context.Context, id int64) error
}

type GrokAccountStateService struct {
	accountRepo grokAccountStateExtraWriter
	now         func() time.Time
}

func NewGrokAccountStateService(accountRepo grokAccountStateExtraWriter) *GrokAccountStateService {
	return &GrokAccountStateService{
		accountRepo: accountRepo,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s *GrokAccountStateService) PersistProbeResult(
	ctx context.Context,
	account *Account,
	modelID string,
	resp *http.Response,
	probeErr error,
) {
	if s == nil || s.accountRepo == nil || account == nil {
		return
	}
	if NormalizeCompatibleGatewayPlatform(account.Platform) != PlatformGrok {
		return
	}

	updates := buildGrokProbeStateExtraUpdates(account, modelID, resp, probeErr, s.now())
	s.persistExtraUpdates(ctx, account, updates)
	s.persistBackgroundRuntimeState(ctx, GrokRuntimeFeedbackInput{
		Account:        account,
		RequestedModel: modelID,
		StatusCode:     grokProbeStatusCode(resp),
		Err:            probeErr,
	})
	s.persistInvalidCredentialAccountError(ctx, GrokRuntimeFeedbackInput{
		Account:        account,
		RequestedModel: modelID,
		StatusCode:     grokProbeStatusCode(resp),
		Err:            probeErr,
	})
	s.clearRecoveredAccountError(ctx, account, grokProbeStatusCode(resp), probeErr)
}

func (s *GrokAccountStateService) PersistSyncSnapshot(
	ctx context.Context,
	account *Account,
	snapshot grokStateSyncSnapshot,
	statusCode int,
	syncErr error,
) {
	if s == nil || s.accountRepo == nil || account == nil {
		return
	}
	if NormalizeCompatibleGatewayPlatform(account.Platform) != PlatformGrok {
		return
	}

	updates := buildGrokSyncStateExtraUpdates(account, snapshot)
	s.persistExtraUpdates(ctx, account, updates)
	s.persistBackgroundRuntimeState(ctx, GrokRuntimeFeedbackInput{
		Account:    account,
		StatusCode: statusCode,
		Err:        syncErr,
		Endpoint:   grokSessionRateLimitsEndpoint,
	})
	s.persistInvalidCredentialAccountError(ctx, GrokRuntimeFeedbackInput{
		Account:    account,
		StatusCode: statusCode,
		Err:        syncErr,
		Endpoint:   grokSessionRateLimitsEndpoint,
	})
	s.clearRecoveredAccountError(ctx, account, statusCode, syncErr)
}

func (s *GrokAccountStateService) persistExtraUpdates(ctx context.Context, account *Account, updates map[string]any) {
	if s == nil || s.accountRepo == nil || account == nil || len(updates) == 0 {
		return
	}

	updateCtx, cancel := newGrokAccountStateContext(ctx)
	defer cancel()

	if err := s.accountRepo.UpdateExtra(updateCtx, account.ID, updates); err != nil {
		return
	}
	mergeAccountExtra(account, updates)
}

func (s *GrokAccountStateService) persistBackgroundRuntimeState(ctx context.Context, input GrokRuntimeFeedbackInput) {
	if s == nil || s.accountRepo == nil || input.Account == nil {
		return
	}
	if !shouldPersistGrokBackgroundRuntimeState(input) {
		return
	}

	writer, ok := s.accountRepo.(grokRuntimeStateWriter)
	if !ok {
		return
	}

	now := s.now().UTC()
	upstreamModel := resolveGrokRuntimeUpstreamModel(input)
	protocolFamily, capability := resolveGrokRuntimeProtocolAndCapability(
		input.RequestedModel,
		upstreamModel,
		input.ProtocolFamily,
		input.Endpoint,
	)
	runtimeState := buildGrokRuntimeState(input, upstreamModel, protocolFamily, capability, now)
	if len(runtimeState) == 0 {
		return
	}

	updateCtx, cancel := newGrokAccountStateContext(ctx)
	defer cancel()

	if err := writer.UpdateGrokRuntimeState(updateCtx, input.Account.ID, runtimeState); err != nil {
		return
	}
	mergeGrokRuntimeState(input.Account, runtimeState)
}

func (s *GrokAccountStateService) persistInvalidCredentialAccountError(ctx context.Context, input GrokRuntimeFeedbackInput) {
	if s == nil || input.Account == nil {
		return
	}
	writer, ok := s.accountRepo.(grokAccountStateErrorWriter)
	if !ok {
		return
	}
	persistGrokInvalidCredentialAccountError(ctx, writer, input)
}

func (s *GrokAccountStateService) clearRecoveredAccountError(
	ctx context.Context,
	account *Account,
	statusCode int,
	resultErr error,
) {
	if s == nil || account == nil || account.Status != StatusError {
		return
	}
	if resultErr != nil || statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return
	}

	writer, ok := s.accountRepo.(grokAccountStateRecoveryWriter)
	if !ok {
		return
	}

	updateCtx, cancel := newGrokAccountStateContext(ctx)
	defer cancel()

	if err := writer.ClearError(updateCtx, account.ID); err != nil {
		return
	}
	account.Status = StatusActive
	account.ErrorMessage = ""
}

func shouldPersistGrokBackgroundRuntimeState(input GrokRuntimeFeedbackInput) bool {
	account := input.Account
	if account == nil {
		return false
	}

	if input.Err != nil {
		return classifyGrokRuntimeError(input).Class == grokRuntimeErrorClassAuth
	}
	if input.StatusCode <= 0 {
		return false
	}

	runtimeState := account.grokRuntimeSelectionState()
	if !grokBackgroundSuccessShouldClearCooldown(runtimeState) {
		return false
	}
	if runtimeState.LastFailAt == nil {
		return runtimeState.CooldownUntil != nil
	}
	return runtimeState.LastUseAt == nil || runtimeState.LastUseAt.Before(*runtimeState.LastFailAt)
}

func grokBackgroundSuccessShouldClearCooldown(runtimeState grokRuntimeSelectionState) bool {
	switch runtimeState.LastFailClass {
	case grokRuntimeErrorClassAuth, grokRuntimeErrorClassRateLimited:
	default:
		return false
	}

	switch runtimeState.CooldownScope {
	case grokRuntimePenaltyScopeNone, grokRuntimePenaltyScopeAccount:
		return true
	default:
		return false
	}
}

func grokProbeStatusCode(resp *http.Response) int {
	if resp == nil {
		return 0
	}
	return resp.StatusCode
}

func newGrokAccountStateContext(ctx context.Context) (context.Context, context.CancelFunc) {
	base := context.Background()
	if ctx != nil {
		base = context.WithoutCancel(ctx)
	}
	return context.WithTimeout(base, grokAccountStatePersistTimeout)
}
