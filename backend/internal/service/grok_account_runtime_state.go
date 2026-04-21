package service

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/grok"
)

type grokStateSyncSnapshot struct {
	AuthMode     string
	Tier         map[string]any
	QuotaWindows map[string]any
	Capabilities map[string]any
	SyncState    map[string]any
}

type GrokRuntimeFeedbackInput struct {
	Account        *Account
	RequestedModel string
	UpstreamModel  string
	Result         *OpenAIForwardResult
	StatusCode     int
	ProtocolFamily grok.ProtocolFamily
	Endpoint       string
	Err            error
}

type grokRuntimeOutcome string

const (
	grokRuntimeOutcomeSuccess  grokRuntimeOutcome = "success"
	grokRuntimeOutcomeError    grokRuntimeOutcome = "error"
	grokRuntimeOutcomeFailover grokRuntimeOutcome = "failover"
)

type grokRuntimeStateWriter interface {
	UpdateGrokRuntimeState(ctx context.Context, id int64, runtimeState map[string]any) error
}

func buildGrokSyncStateExtraUpdates(account *Account, snapshot grokStateSyncSnapshot) map[string]any {
	if account == nil {
		return nil
	}

	incomingGrok := make(map[string]any, 5)
	if authMode := strings.TrimSpace(snapshot.AuthMode); authMode != "" {
		incomingGrok["auth_mode"] = authMode
	}
	if len(snapshot.Tier) > 0 {
		incomingGrok["tier"] = cloneAnyMap(snapshot.Tier)
	}
	if len(snapshot.QuotaWindows) > 0 {
		incomingGrok["quota_windows"] = cloneAnyMap(snapshot.QuotaWindows)
	}
	if len(snapshot.Capabilities) > 0 {
		incomingGrok["capabilities"] = cloneAnyMap(snapshot.Capabilities)
	}
	if len(snapshot.SyncState) > 0 {
		incomingGrok["sync_state"] = cloneAnyMap(snapshot.SyncState)
	}

	return buildGrokStateExtraPatch(account, incomingGrok)
}

func buildGrokStateExtraPatch(account *Account, incomingGrok map[string]any) map[string]any {
	if account == nil || len(incomingGrok) == 0 {
		return nil
	}

	existing := map[string]any{}
	if current := account.grokExtraMap(); len(current) > 0 {
		existing["grok"] = cloneAnyMap(current)
	}

	normalized := normalizeGrokAccountExtra(existing, map[string]any{"grok": incomingGrok}, account.Type)
	grokExtra := grokExtraMap(normalized)
	if len(grokExtra) == 0 {
		return nil
	}
	return map[string]any{"grok": grokExtra}
}

func buildGrokProbeStateExtraUpdates(account *Account, modelID string, resp *http.Response, probeErr error, now time.Time) map[string]any {
	if account == nil || NormalizeCompatibleGatewayPlatform(account.Platform) != PlatformGrok {
		return nil
	}

	timestamp := now.UTC().Format(time.RFC3339)
	currentExtra := account.grokExtraMap()
	syncState := cloneAnyMap(grokNestedMap(currentExtra["sync_state"]))
	if len(syncState) == 0 {
		syncState = make(map[string]any, 4)
	}
	syncState["last_probe_at"] = timestamp

	if resp != nil && resp.StatusCode > 0 {
		syncState["last_probe_status_code"] = resp.StatusCode
	} else {
		delete(syncState, "last_probe_status_code")
	}

	success := probeErr == nil && resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300
	if success {
		syncState["last_probe_ok_at"] = timestamp
		delete(syncState, "last_probe_error")
		delete(syncState, "last_probe_error_at")
	} else {
		syncState["last_probe_error_at"] = timestamp
		if probeErr != nil {
			syncState["last_probe_error"] = strings.TrimSpace(probeErr.Error())
		} else if resp != nil {
			syncState["last_probe_error"] = strings.TrimSpace(resp.Status)
		}
	}

	snapshot := grokStateSyncSnapshot{
		SyncState: syncState,
	}
	if success {
		snapshot.Capabilities = buildGrokProbeCapabilities(account, modelID)
	}
	return buildGrokSyncStateExtraUpdates(account, snapshot)
}

func buildGrokProbeCapabilities(account *Account, modelID string) map[string]any {
	capabilities := buildGrokKnownCapabilities(account, modelID, "")
	if account == nil {
		return capabilities
	}

	tier := account.GrokTierState().Normalized
	if tier == grok.TierUnknown {
		tier = grokInferTierFromCapabilityState(parseGrokCapabilityState(capabilities))
	}
	if tier == grok.TierUnknown {
		return capabilities
	}

	widened := buildGrokTierCapabilityBaselineFromState(account, parseGrokCapabilityState(capabilities), tier)
	if len(widened) == 0 {
		return capabilities
	}
	return widened
}

func buildGrokKnownCapabilities(account *Account, modelID string, capabilityHint grok.Capability) map[string]any {
	if account == nil {
		return nil
	}

	currentExtra := account.grokExtraMap()
	capabilities := cloneAnyMap(grokNestedMap(currentExtra["capabilities"]))
	if len(capabilities) == 0 {
		capabilities = make(map[string]any, 2)
	}

	state := account.grokCapabilities()
	models := make(map[string]struct{}, len(state.models)+1)
	for model := range state.models {
		models[model] = struct{}{}
	}
	operations := make(map[grok.Capability]bool, len(state.operations)+1)
	for capability, allowed := range state.operations {
		operations[capability] = allowed
	}

	if canonicalModel := grok.ResolveCanonicalModelID(modelID); canonicalModel != "" {
		models[canonicalModel] = struct{}{}
		if spec, ok := grok.LookupModelSpec(canonicalModel); ok && spec.Capability != "" {
			operations[spec.Capability] = true
		}
	}
	if capabilityHint != "" {
		operations[capabilityHint] = true
	}

	if len(models) > 0 {
		modelIDs := make([]string, 0, len(models))
		for model := range models {
			modelIDs = append(modelIDs, model)
		}
		sort.Strings(modelIDs)
		capabilities["models"] = modelIDs
	} else if state.hasModelSignal {
		capabilities["models"] = []string{}
	} else {
		delete(capabilities, "models")
	}

	if len(operations) > 0 {
		enabled := make([]string, 0, len(operations))
		for capability, allowed := range operations {
			if allowed {
				enabled = append(enabled, string(capability))
				delete(capabilities, string(capability))
				continue
			}
			capabilities[string(capability)] = false
		}
		if len(enabled) > 0 {
			sort.Strings(enabled)
			capabilities["operations"] = enabled
		} else {
			delete(capabilities, "operations")
		}
	}

	return capabilities
}

func (s *OpenAIGatewayService) PersistGrokRuntimeFeedback(ctx context.Context, input GrokRuntimeFeedbackInput) {
	if s == nil {
		return
	}
	persistGrokRuntimeFeedbackToRepo(ctx, s.accountRepo, input)
}

func persistGrokRuntimeFeedbackToRepo(ctx context.Context, repo AccountRepository, input GrokRuntimeFeedbackInput) {
	if repo == nil || input.Account == nil {
		return
	}
	if NormalizeCompatibleGatewayPlatform(input.Account.Platform) != PlatformGrok {
		return
	}

	now := time.Now().UTC()
	upstreamModel := resolveGrokRuntimeUpstreamModel(input)
	protocolFamily, capability := resolveGrokRuntimeProtocolAndCapability(
		input.RequestedModel,
		upstreamModel,
		input.ProtocolFamily,
		input.Endpoint,
	)

	updateCtx, cancel := newGrokAccountStateContext(ctx)
	defer cancel()

	if updates := buildGrokRuntimeCapabilityExtraUpdates(input.Account, input, upstreamModel, capability); len(updates) > 0 {
		if err := repo.UpdateExtra(updateCtx, input.Account.ID, updates); err == nil {
			mergeAccountExtra(input.Account, updates)
		}
	}

	runtimeState := buildGrokRuntimeState(input, upstreamModel, protocolFamily, capability, now)
	if len(runtimeState) == 0 {
		return
	}

	writer, ok := repo.(grokRuntimeStateWriter)
	if !ok {
		return
	}
	if err := writer.UpdateGrokRuntimeState(updateCtx, input.Account.ID, runtimeState); err != nil {
		return
	}
	mergeGrokRuntimeState(input.Account, runtimeState)
}

func buildGrokRuntimeCapabilityExtraUpdates(account *Account, input GrokRuntimeFeedbackInput, upstreamModel string, capability grok.Capability) map[string]any {
	if account == nil {
		return nil
	}

	capabilities := buildGrokRuntimeCapabilities(account, input, upstreamModel, capability)
	if len(capabilities) == 0 {
		return nil
	}
	return buildGrokStateExtraPatch(account, map[string]any{"capabilities": capabilities})
}

func mergeGrokRuntimeState(account *Account, runtimeState map[string]any) {
	if account == nil || len(runtimeState) == 0 {
		return
	}
	mergeAccountExtra(account, buildGrokStateExtraPatch(account, map[string]any{"runtime_state": runtimeState}))
}

func resolveGrokRuntimeUpstreamModel(input GrokRuntimeFeedbackInput) string {
	if input.Result != nil {
		if model := strings.TrimSpace(input.Result.UpstreamModel); model != "" {
			return grok.ResolveCanonicalModelID(model)
		}
		if model := strings.TrimSpace(input.Result.Model); model != "" {
			return grok.ResolveCanonicalModelID(model)
		}
	}
	if model := strings.TrimSpace(input.UpstreamModel); model != "" {
		return grok.ResolveCanonicalModelID(model)
	}
	if model := strings.TrimSpace(input.RequestedModel); model != "" {
		return grok.ResolveCanonicalModelID(model)
	}
	return ""
}

func resolveGrokRuntimeProtocolAndCapability(requestedModel string, upstreamModel string, protocolHint grok.ProtocolFamily, endpoint string) (grok.ProtocolFamily, grok.Capability) {
	protocolFamily := protocolHint
	var capability grok.Capability

	for _, candidate := range []string{upstreamModel, requestedModel} {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		spec, ok := grok.LookupModelSpec(candidate)
		if !ok {
			continue
		}
		if capability == "" {
			capability = spec.Capability
		}
		if protocolFamily == "" {
			protocolFamily = spec.ProtocolFamily
		}
		break
	}

	endpointProtocol, endpointCapability := resolveGrokRuntimeEndpointSignal(endpoint)
	if protocolFamily == "" {
		protocolFamily = endpointProtocol
	}
	if capability == "" {
		capability = endpointCapability
	}

	return protocolFamily, capability
}

func resolveGrokRuntimeEndpointSignal(endpoint string) (grok.ProtocolFamily, grok.Capability) {
	endpoint = strings.ToLower(strings.TrimSpace(endpoint))
	switch {
	case strings.HasSuffix(endpoint, "/chat/completions"):
		return grok.ProtocolFamilyChatCompletions, grok.CapabilityChat
	case strings.HasSuffix(endpoint, "/responses"):
		return grok.ProtocolFamilyResponses, grok.CapabilityChat
	case strings.HasSuffix(endpoint, "/messages"):
		return grok.ProtocolFamilyMessages, grok.CapabilityChat
	case strings.HasSuffix(endpoint, "/images/generations"):
		return grok.ProtocolFamilyResponses, grok.CapabilityImage
	case strings.HasSuffix(endpoint, "/images/edits"):
		return grok.ProtocolFamilyResponses, grok.CapabilityImageEdit
	case strings.HasSuffix(endpoint, "/videos"),
		strings.Contains(endpoint, "/videos/"):
		return grok.ProtocolFamilyMediaJob, grok.CapabilityVideo
	default:
		return "", ""
	}
}

func buildGrokRuntimeState(input GrokRuntimeFeedbackInput, upstreamModel string, protocolFamily grok.ProtocolFamily, capability grok.Capability, now time.Time) map[string]any {
	account := input.Account
	if account == nil {
		return nil
	}

	runtimeState := cloneAnyMap(grokNestedMap(account.grokExtraMap()["runtime_state"]))
	if len(runtimeState) == 0 {
		runtimeState = make(map[string]any, 12)
	}

	timestamp := now.UTC().Format(time.RFC3339)
	outcome, statusCode, failReason := classifyGrokRuntimeOutcome(input)
	classification := grokRuntimeErrorClassification{}
	if outcome != grokRuntimeOutcomeSuccess {
		classification = classifyGrokRuntimeError(input)
		if classification.StatusCode > 0 {
			statusCode = classification.StatusCode
		}
		if strings.TrimSpace(classification.Reason) != "" {
			failReason = classification.Reason
		}
	}

	runtimeState["last_request_at"] = timestamp
	runtimeState["last_outcome"] = string(outcome)

	if statusCode > 0 {
		runtimeState["last_request_status_code"] = statusCode
	} else {
		delete(runtimeState, "last_request_status_code")
	}

	if requestedModel := strings.TrimSpace(input.RequestedModel); requestedModel != "" {
		runtimeState["last_request_model"] = requestedModel
	} else {
		delete(runtimeState, "last_request_model")
	}

	if upstreamModel != "" {
		runtimeState["last_request_upstream_model"] = upstreamModel
	} else {
		delete(runtimeState, "last_request_upstream_model")
	}

	if protocolFamily != "" {
		runtimeState["last_request_protocol_family"] = string(protocolFamily)
	} else {
		delete(runtimeState, "last_request_protocol_family")
	}

	if capability != "" {
		runtimeState["last_request_capability"] = string(capability)
	} else {
		delete(runtimeState, "last_request_capability")
	}

	switch outcome {
	case grokRuntimeOutcomeSuccess:
		runtimeState["last_use_at"] = timestamp
		delete(runtimeState, "selection_cooldown_until")
		delete(runtimeState, "selection_cooldown_scope")
		delete(runtimeState, "selection_cooldown_model")
	case grokRuntimeOutcomeError, grokRuntimeOutcomeFailover:
		runtimeState["last_fail_at"] = timestamp
		if statusCode > 0 {
			runtimeState["last_fail_status_code"] = statusCode
		} else {
			delete(runtimeState, "last_fail_status_code")
		}
		if failReason != "" {
			runtimeState["last_fail_reason"] = failReason
		} else {
			delete(runtimeState, "last_fail_reason")
		}
		if classification.Class != "" {
			runtimeState["last_fail_class"] = string(classification.Class)
		} else {
			delete(runtimeState, "last_fail_class")
		}
		if classification.Scope != "" {
			runtimeState["last_fail_scope"] = string(classification.Scope)
		} else {
			delete(runtimeState, "last_fail_scope")
		}
		runtimeState["last_fail_retryable"] = classification.Retryable
		if cooldownUntil := grokRuntimeCooldownUntil(now, classification); cooldownUntil != nil {
			runtimeState["selection_cooldown_until"] = cooldownUntil.Format(time.RFC3339)
			if classification.Scope != "" {
				runtimeState["selection_cooldown_scope"] = string(classification.Scope)
			} else {
				delete(runtimeState, "selection_cooldown_scope")
			}
			if classification.Scope == grokRuntimePenaltyScopeModel {
				if cooldownModel := grokRuntimeCooldownModel(input, upstreamModel); cooldownModel != "" {
					runtimeState["selection_cooldown_model"] = cooldownModel
				} else {
					delete(runtimeState, "selection_cooldown_model")
				}
			} else {
				delete(runtimeState, "selection_cooldown_model")
			}
		} else {
			delete(runtimeState, "selection_cooldown_until")
			delete(runtimeState, "selection_cooldown_scope")
			delete(runtimeState, "selection_cooldown_model")
		}
		if outcome == grokRuntimeOutcomeFailover {
			runtimeState["last_failover_at"] = timestamp
		}
	}

	return runtimeState
}

func buildGrokRuntimeCapabilities(account *Account, input GrokRuntimeFeedbackInput, upstreamModel string, capability grok.Capability) map[string]any {
	if account == nil {
		return nil
	}

	currentCapabilities := cloneAnyMap(grokNestedMap(account.grokExtraMap()["capabilities"]))
	var nextCapabilities map[string]any
	if input.Err == nil {
		nextCapabilities = buildGrokKnownCapabilities(account, upstreamModel, capability)
	} else if shouldPruneGrokRuntimeModelSignal(input, upstreamModel) {
		nextCapabilities = buildGrokPrunedCapabilities(account, upstreamModel)
	}

	if len(nextCapabilities) == 0 {
		return nil
	}

	if grokCapabilityStatesEqual(
		grokCapabilityStateFromCapabilities(currentCapabilities),
		grokCapabilityStateFromCapabilities(nextCapabilities),
	) {
		return nil
	}

	return nextCapabilities
}

func buildGrokPrunedCapabilities(account *Account, modelID string) map[string]any {
	if account == nil {
		return nil
	}

	canonicalModel := grok.ResolveCanonicalModelID(modelID)
	if canonicalModel == "" {
		return nil
	}

	currentExtra := account.grokExtraMap()
	capabilities := cloneAnyMap(grokNestedMap(currentExtra["capabilities"]))
	if len(capabilities) == 0 {
		return nil
	}

	state := account.grokCapabilities()
	if !state.hasModelSignal {
		return nil
	}

	delete(state.models, canonicalModel)
	modelIDs := make([]string, 0, len(state.models))
	for model := range state.models {
		modelIDs = append(modelIDs, model)
	}
	sort.Strings(modelIDs)
	capabilities["models"] = modelIDs
	return capabilities
}

func classifyGrokRuntimeOutcome(input GrokRuntimeFeedbackInput) (grokRuntimeOutcome, int, string) {
	statusCode := input.StatusCode
	if input.Err == nil {
		if statusCode <= 0 {
			statusCode = http.StatusOK
		}
		return grokRuntimeOutcomeSuccess, statusCode, ""
	}

	var failoverErr *UpstreamFailoverError
	if errors.As(input.Err, &failoverErr) {
		if failoverErr != nil && failoverErr.StatusCode > 0 {
			statusCode = failoverErr.StatusCode
		}
		reason := strings.TrimSpace(ExtractUpstreamErrorMessage(failoverErr.ResponseBody))
		if reason == "" {
			reason = strings.TrimSpace(failoverErr.FailureReason)
		}
		if reason == "" {
			reason = strings.TrimSpace(input.Err.Error())
		}
		return grokRuntimeOutcomeFailover, statusCode, reason
	}

	return grokRuntimeOutcomeError, statusCode, strings.TrimSpace(input.Err.Error())
}

func shouldPruneGrokRuntimeModelSignal(input GrokRuntimeFeedbackInput, upstreamModel string) bool {
	if input.Err == nil || strings.TrimSpace(upstreamModel) == "" {
		return false
	}
	return classifyGrokRuntimeError(input).Class == grokRuntimeErrorClassModelUnsupported
}

func grokRuntimeCooldownUntil(now time.Time, classification grokRuntimeErrorClassification) *time.Time {
	if classification.Scope == grokRuntimePenaltyScopeNone || classification.Cooldown <= 0 {
		return nil
	}
	value := now.UTC().Add(classification.Cooldown)
	return &value
}

func grokRuntimeCooldownModel(input GrokRuntimeFeedbackInput, upstreamModel string) string {
	if model := grok.ResolveCanonicalModelID(strings.TrimSpace(upstreamModel)); model != "" {
		return model
	}
	return grok.ResolveCanonicalModelID(strings.TrimSpace(input.RequestedModel))
}

func grokCapabilityStateFromCapabilities(capabilities map[string]any) grokCapabilityState {
	account := &Account{
		Platform: PlatformGrok,
		Extra: map[string]any{
			"grok": map[string]any{
				"capabilities": cloneAnyMap(capabilities),
			},
		},
	}
	return account.grokCapabilities()
}

func grokCapabilityStatesEqual(left, right grokCapabilityState) bool {
	if left.hasModelSignal != right.hasModelSignal || left.hasOperationSignal != right.hasOperationSignal {
		return false
	}
	if len(left.models) != len(right.models) || len(left.operations) != len(right.operations) {
		return false
	}
	for model := range left.models {
		if _, ok := right.models[model]; !ok {
			return false
		}
	}
	for capability, allowed := range left.operations {
		if right.operations[capability] != allowed {
			return false
		}
	}
	return true
}
