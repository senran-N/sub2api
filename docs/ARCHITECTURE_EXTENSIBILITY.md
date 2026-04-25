# Architecture Extensibility Notes

This document records the current extension patterns that should be reused instead of re-implementing one-off flows.

## Frontend

### Admin paginated list views

Use `frontend/src/composables/useTableLoader.ts` as the default skeleton for admin list data:

- request cancellation
- persisted page size
- debounced reload
- caller-owned pagination state when the page already owns filter/page orchestration
- optional error callback
- optional post-load hook for page-local secondary hydration or side effects
- optional response-to-pagination sync
- explicit cleanup via `dispose()`

Current reference integrations:

- `frontend/src/views/admin/useAnnouncementsViewData.ts`
- `frontend/src/views/admin/usePromoCodesViewData.ts`
- `frontend/src/views/admin/redeem/useRedeemViewData.ts`
- `frontend/src/views/admin/useSubscriptionsViewData.ts`
- `frontend/src/views/admin/users/useUsersViewData.ts`
- `frontend/src/views/admin/proxies/useProxyListData.ts`
- `frontend/src/views/admin/usage/useUsageViewData.ts`
- `frontend/src/views/admin/orders/AdminOrdersView.vue`

When adding a new admin list page, prefer composing around `useTableLoader` instead of re-creating:

- `AbortController` state
- ad hoc debounce timers
- duplicated page/page_size reset logic
- duplicated abort-error checks
- bespoke "load success then hydrate secondary data" glue when `onLoaded` is enough

### Account modal structure

Keep account modal behavior behind shared helper seams instead of growing
`CreateAccountModal.vue` and `EditAccountModal.vue` with duplicated local conventions.

Current rule:

- shared modal class assembly lives in `frontend/src/components/account/accountModalClasses.ts`;
- stale async request guards live in `frontend/src/components/account/accountModalRequestGuard.ts`;
- account mutation section visibility lives in `frontend/src/components/account/accountMutationProfiles.ts`; create/edit UI should resolve a profile instead of re-stating platform/type matrices inline;
- account mutation section composition refs live in `frontend/src/components/account/useAccountMutationSections.ts`; modal files should consume its visibility/category refs rather than re-creating profile-derived section booleans locally;
- create/edit/bulk account mutation payload assembly lives in `frontend/src/components/account/accountMutationPayload.ts`; modal files should call that helper for final request payloads instead of rebuilding JSONB merge sentinels, quota overlays, OpenAI runtime fields, or provider-specific credential mutations locally;
- edit-account payload state adaptation lives in `frontend/src/components/account/useEditAccountMutationPayload.ts`; the edit modal should pass grouped section/composable state there and keep submit handlers focused on validation, mixed-channel confirmation, and API update orchestration;
- lower-level credential and form helpers remain in `accountModalShared.ts`, `createAccountModalHelpers.ts`, `editAccountModalHelpers.ts`, and `credentialsBuilder.ts`;
- create-account type cards should compose `frontend/src/components/account/CreateAccountChoiceCard.vue` instead of re-implementing selected/idle card and icon styling inside each provider section;
- reusable create-account basic name/notes fields live in `frontend/src/components/account/CreateAccountBasicInfoSection.vue`; the create modal should bind the field values and computed label/hint text there instead of owning those inputs inline;
- reusable Anthropic add-method radio controls live in `frontend/src/components/account/AnthropicAddMethodSection.vue`; create/edit forms should bind the selected OAuth/setup-token method there instead of duplicating radio-card markup inline;
- reusable create-account scheduling, proxy, expiry, mixed-scheduling, overage, and group controls live in `frontend/src/components/account/CreateAccountSchedulingSection.vue`; the create modal should bind scalar/group state there and keep submit-time payload assembly outside the component;
- reusable account-modal switch controls live in `frontend/src/components/account/AccountModalSwitch.vue`; extracted provider sections should compose it instead of relying on parent-scoped switch classes;
- reusable warmup interception controls live in `frontend/src/components/account/WarmupSection.vue`; create/edit/bulk flows should bind `interceptWarmupRequests` there instead of duplicating switch markup inline;
- reusable expiration auto-pause controls live in `frontend/src/components/account/AutoPauseOnExpiredSection.vue`; the create modal should bind `autoPauseOnExpired` there instead of duplicating switch markup inline;
- reusable pool-mode controls live in `frontend/src/components/account/PoolModeSection.vue`; create/edit sections should bind `enabled` and retry count there instead of duplicating pool-mode notice/input markup for compatible API-key and Bedrock forms;
- edit-account compatible credential field and pool-mode hydration state lives in `frontend/src/components/account/useEditCredentialFields.ts`; the edit modal should consume its refs and hydrate/reset methods instead of owning base URL, replacement secret, session token, and pool retry defaults inline;
- reusable custom-error-code controls live in `frontend/src/components/account/CustomErrorCodesSection.vue`; modal files should bind enable/input/code state there and keep provider validation, confirmation, and toast behavior in the parent orchestration layer;
- edit-account custom-error-code state lives in `frontend/src/components/account/useEditCustomErrorCodes.ts`; the edit modal should inject confirm/toast callbacks and consume its hydrate/reset and mutation handlers instead of owning code-list mechanics inline;
- bulk-only "apply this field" chrome lives in `frontend/src/components/account/BulkEditApplySection.vue`; bulk forms should wrap shared sections there when they need a separate batch-apply toggle instead of adding bulk flags to reusable create/edit controls;
- bulk-only account-edit wrappers live in `frontend/src/components/account/BulkEditBaseUrlSection.vue`, `BulkEditOpenAIOptionsSection.vue`, `BulkEditProxySection.vue`, `BulkEditGroupsSection.vue`, `BulkEditStatusSection.vue`, `BulkEditScalarFieldsSection.vue`, and `BulkEditContextNoticeSection.vue`; they own batch-apply framing and bulk-only layout while `BulkEditAccountModal.vue` keeps form lifecycle, payload assembly, submit orchestration, and confirmation handling;
- bulk-only scalar field markup lives in `frontend/src/components/account/BulkEditNumberField.vue`; use it from bulk wrapper sections instead of duplicating checkbox/input pairs in the modal body;
- reusable compatible API-key/upstream credential composition lives in `frontend/src/components/account/CompatibleCredentialsSection.vue`; create/edit modals should pass base URL/API key, Gemini API-key tier where applicable, model restriction, pool mode, and custom-error-code state into it instead of owning the shared compatible credential block inline;
- reusable temporary-unschedulable rule controls live in `frontend/src/components/account/TempUnschedRulesSection.vue`; modal files should bind rule arrays there and keep rule creation, ordering, and payload validation in the parent/helper layer;
- edit-account temporary-unschedulable rule state lives in `frontend/src/components/account/useEditAccountTempUnschedRules.ts`; the edit modal should consume its refs, stable key, hydrate method, and mutation handlers instead of owning rule array mechanics inline;
- reusable model restriction controls live in `frontend/src/components/account/ModelRestrictionSection.vue`; create/edit forms should bind whitelist/mapping state there and keep provider-specific mapping mutation helpers in the parent instead of duplicating whitelist/mapping UI blocks inline;
- edit-account model restriction and Antigravity mapping state lives in `frontend/src/components/account/useEditAccountModelRestrictions.ts`; the edit modal should consume its refs, hydrate methods, and mapping mutation handlers instead of owning mapping state machines inline;
- reusable account quota limit and notification controls live in `frontend/src/components/account/QuotaLimitSection.vue`; create/edit forms should bind quota reset and notification threshold state there instead of duplicating `QuotaLimitCard` plus three `QuotaNotifyToggle` blocks inline;
- edit-account quota limit hydration/reset state lives in `frontend/src/components/account/useEditAccountQuotaLimits.ts`; the edit modal should consume its refs and hydrate method instead of owning compatible/Bedrock quota limit state machines inline;
- reusable quota-control switch-card chrome lives in `frontend/src/components/account/QuotaControlCard.vue`; quota subsections should compose it so card framing, labels, hints, and switch behavior do not drift across create/edit forms;
- reusable Anthropic OAuth window-cost controls live in `frontend/src/components/account/WindowCostControlSection.vue`; modal files should bind cost limit and sticky-reserve values there instead of duplicating currency-affix inputs inline;
- reusable Anthropic OAuth session-limit controls live in `frontend/src/components/account/SessionLimitControlSection.vue`; modal files should bind max-session and idle-timeout values there instead of duplicating the session limit card inline;
- reusable Anthropic OAuth RPM controls live in `frontend/src/components/account/RpmLimitControlSection.vue`; modal files should bind RPM, strategy, sticky-buffer, and user-message queue values there instead of duplicating strategy cards and segment controls inline;
- reusable Anthropic OAuth TLS fingerprint controls live in `frontend/src/components/account/TlsFingerprintControlSection.vue`; modal files should bind profile state and profile lists there instead of owning the select card inline;
- reusable Anthropic OAuth session masking controls live in `frontend/src/components/account/SessionIdMaskingControlSection.vue`; modal files should bind the boolean there instead of duplicating a no-body quota card inline;
- reusable Anthropic OAuth cache TTL override controls live in `frontend/src/components/account/CacheTtlOverrideSection.vue`; modal files should bind enabled/target values there instead of duplicating the TTL select card inline;
- reusable Anthropic OAuth custom base URL controls live in `frontend/src/components/account/CustomBaseUrlControlSection.vue`; modal files should bind enabled/base URL values there instead of duplicating relay input markup inline;
- reusable Anthropic quota-control section assembly lives in `frontend/src/components/account/AnthropicQuotaControlsSection.vue`; create/edit modals should bind quota-control state there instead of duplicating the wrapper heading and lower-level quota control composition inline;
- edit-account Anthropic quota-control hydration/reset state lives in `frontend/src/components/account/useEditAccountQuotaControls.ts`; the edit modal should consume its refs and hydrate method instead of owning quota-control state machines inline;
- reusable create-account platform selection lives in `frontend/src/components/account/CreateAccountPlatformSelector.vue`; the create modal should bind it with `v-model` instead of duplicating platform segmented-control classes and button state locally;
- reusable OpenAI create-account type selection lives in `frontend/src/components/account/OpenAIAccountTypeSection.vue`; the create modal should bind `accountCategory` there instead of rendering OpenAI OAuth/API key cards inline;
- reusable OpenAI create-account runtime options live in `frontend/src/components/account/OpenAIOptionsSection.vue`; the create modal should bind passthrough, Responses WebSocket mode, and Codex CLI-only state there instead of owning provider-specific switch/select markup inline;
- edit-account OpenAI/Anthropic runtime option hydration/reset state lives in `frontend/src/components/account/useEditAccountRuntimeOptions.ts`; the edit modal should consume the composable refs and hydrate method instead of owning passthrough/WebSocket/Codex state machines inline;
- reusable Anthropic create-account type selection lives in `frontend/src/components/account/AnthropicAccountTypeSection.vue`; the create modal should bind `accountCategory` there instead of rendering Claude Code / Claude Console / Bedrock cards inline;
- reusable Anthropic create-account runtime options live in `frontend/src/components/account/AnthropicOptionsSection.vue`; the create modal should bind API-key passthrough state there instead of owning provider-specific switch markup inline;
- reusable Bedrock credential fields live in `frontend/src/components/account/BedrockCredentialsSection.vue`; create/edit modals should bind auth mode, keys, region, and force-global state there instead of owning Bedrock credential markup inline;
- edit-account Bedrock credential hydration/reset state lives in `frontend/src/components/account/useEditBedrockCredentials.ts`; the edit modal should consume its refs and hydrate method instead of owning saved-secret preservation and auth-mode defaults inline;
- reusable Bedrock edit section assembly lives in `frontend/src/components/account/EditBedrockCredentialsSection.vue`; the edit modal should bind Bedrock credential/model restriction/pool state there instead of duplicating that provider wrapper inline;
- reusable edit-account core fields live in `frontend/src/components/account/EditAccountCoreFieldsSection.vue`; the edit modal should bind basic identity, scheduling scalar, status, Antigravity overage, and group state there instead of owning those controls inline;
- edit-account base form state lives in `frontend/src/components/account/useEditAccountFormState.ts`; the edit modal should consume form refs, status options, expiration input, auto-pause, mixed scheduling, and overage hydration there instead of owning form defaults inline;
- edit-account mixed-channel warning state lives in `frontend/src/components/account/useEditMixedChannelWarning.ts`; the edit modal should delegate risk precheck, confirmation dialog state, conflict retry, and `confirm_mixed_channel_risk` payload flag handling there instead of rebuilding that flow inline;
- reusable Grok create-account type selection lives in `frontend/src/components/account/GrokAccountTypeSection.vue`; the create modal should bind `accountCategory` into that section instead of rendering Grok API key/upstream/session cards inline;
- reusable Grok session credential/import UI lives in `frontend/src/components/account/GrokSessionCredentialsSection.vue`; the create modal should bind session token, batch import, dry-run, and result state there instead of mounting `GrokSessionBatchImportPanel` directly;
- reusable Grok session edit-token UI lives in `frontend/src/components/account/EditGrokSessionCredentialsSection.vue`; the edit modal should bind replacement session-token state there and leave existing-token preservation in payload helpers;
- reusable Gemini create-account type selection lives in `frontend/src/components/account/GeminiAccountTypeSection.vue`; the create modal should bind `accountCategory` and parent-owned help dialog state there instead of rendering Gemini OAuth/API key cards inline;
- reusable Gemini setup/quota help dialog lives in `frontend/src/components/account/GeminiHelpDialog.vue`; the create modal should pass parent-owned visibility and catalog links into it instead of keeping the long-form help template and table styles inline;
- reusable Gemini OAuth subtype and tier fallback selection lives in `frontend/src/components/account/GeminiOAuthOptionsSection.vue`; the create modal should bind `oauthType`, capability flags, and tier refs there instead of owning the nested Google One / Code Assist / AI Studio card template inline;
- reusable Gemini API-key tier fallback selection lives in `frontend/src/components/account/GeminiApiKeyTierSection.vue`; the create modal should bind the AI Studio tier ref there instead of keeping provider-specific select markup inline;
- reusable Antigravity create-account type selection lives in `frontend/src/components/account/AntigravityAccountTypeSection.vue`; the create modal should bind the selected account type there instead of rendering provider-specific choice cards inline;
- reusable Antigravity upstream credential fields live in `frontend/src/components/account/AntigravityUpstreamCredentialsSection.vue`; the create modal should bind `baseUrl` / `apiKey` into that section instead of rendering provider-specific credential inputs inline;
- reusable Antigravity model mapping UI lives in `frontend/src/components/account/AntigravityModelMappingSection.vue`; create/edit modals should pass mapping state/events into that section instead of owning the provider-specific mapping row template inline;
- Grok runtime summary rendering lives in `frontend/src/components/account/GrokRuntimeSummary.vue`; the edit modal should delegate the display block there instead of owning Grok runtime computed state directly;
- modal `.vue` files should own rendering and orchestration only, not new reusable state machines or class-building helpers.

### Ops dashboard data flow

Use `frontend/src/views/admin/ops/useOpsDashboardData.ts` for the Ops dashboard's core data request lifecycle:

- dashboard request cancellation and stale response suppression
- overview/snapshot/trend/latency/error-distribution loading state
- dashboard refresh token cadence for child cards
- metric threshold loading
- request error normalization through `resolveRequestErrorMessage`

`OpsDashboard.vue` should keep route-query sync, fullscreen/dialog state, and layout orchestration. Do not reintroduce dashboard fetch sequence counters or AbortController ownership in the page component.

### Request error normalization

Use `frontend/src/utils/requestError.ts` for:

- `isAbortError`
- `resolveRequestErrorMessage`

This keeps request cancellation and API error extraction consistent across composables.

## Backend

### Account mutation normalization

Admin account create/update/bulk paths and legacy account service mutations should reuse
the normalization helpers in `backend/internal/service/admin_service_account_config.go`.

Current rule:

- `accountMutationBuilder` owns admin create/update group, proxy, load-factor, rate-multiplier, mixed-channel, credentials, and mutable `extra` normalization;
- `normalizeAccountMutationPayload` is the shared create-time payload normalizer for credentials and platform `extra`;
- `applyMutableAccountExtra` is the update-time `extra` entry point and must preserve quota usage plus provider-owned nested state such as `extra.grok.sync_state` and `extra.grok.runtime_state`;
- admin bulk updates that contain `credentials` or `extra` must hydrate and save each account individually, because repository-level JSONB top-level merge is only safe for scalar-only bulk changes;
- bulk `credentials` updates are top-level merges over the existing credentials so partial bulk edits do not drop secrets such as `api_key`, while `extra` updates still flow through provider-aware deep normalization.

### Account snapshot sync after writes

Single-account repository writes that must refresh scheduler/sticky-session visibility should reuse
`backend/internal/repository/account_repo.go:syncSchedulerAccountSnapshot`.

Current rule:

- when a single-account cache entry already exists, refresh it from the authoritative account base row and preserve cached group/proxy edges instead of re-querying the full hydrated account graph;
- when the cache entry is missing, fall back to the full repository hydration path;
- do not re-introduce ad hoc `GetByID` + `SetAccount` sequences in individual write methods.

### Compatible gateway platforms

Treat the OpenAI-compatible gateway as shared protocol infrastructure, not as OpenAI-owned platform logic.

Current rule:

- `backend/internal/service/compatible_gateway_platform.go` is the platform authority for compatible routes.
- `backend/internal/handler/compatible_gateway_handler.go` owns the shared compatible runtime surface, including compatible `/models` discovery/default listing.
- `backend/internal/handler/grok_gateway_handler.go` is the provider-owned Grok facade. Explicit `/grok/v1/*` routes and generic `/v1/*` dispatch for Grok should terminate there instead of entering `OpenAIGatewayHandler` directly.
- `backend/internal/handler/compatible_gateway_runtime_handler.go` is the neutral shared execution seam for compatible chat/responses/messages/passthrough traffic. It now depends on a neutral protocol-runtime interface rather than a concrete `OpenAIGatewayHandler`.
- `backend/internal/handler/compatible_gateway_text_handler.go` owns the shared `/responses`, `/chat/completions`, and `/messages` HTTP orchestration for compatible text traffic. `OpenAIGatewayHandler` should delegate there instead of remaining the owner of shared text-route bodies.
- `backend/internal/handler/compatible_gateway_handler.go` should dispatch OpenAI-family compatible traffic through `CompatibleGatewayRuntimeHandler`, not by selecting `OpenAIGatewayHandler` directly. `OpenAIGatewayHandler` is an implementation behind that seam instead of the shared control-plane owner.
- `backend/internal/service/gateway_runtime_kernel.go` defines the provider-neutral runtime envelope (`GatewayRequest`, `GatewayResponse`, `GatewayRuntime`, `GatewayTransport`) that new provider runtimes should target before adding HTTP handler-specific adapters.
- The Grok facade must force-bind `PlatformGrok` into request context before delegating into shared compatible runtime paths so selection, models, and downstream transport helpers never fall back to OpenAI ownership by accident.
- `backend/internal/handler/handler.go` should expose both `Handlers.CompatibleGateway` for shared compatible helpers and `Handlers.GrokGateway` for Grok-owned control-plane entrypoints.
- `openai` and `grok` both attach through that helper layer and may reuse the same HTTP/streaming forwarding stack.
- Route forcing and request context decide platform ownership; selection and scheduler code must resolve the effective compatible platform from context instead of hardcoding `PlatformOpenAI`.
- `backend/internal/handler/gateway_handler.go` should retain only native Anthropic/Gemini/Antigravity ownership; compatible control-plane entrypoints should move behind the shared compatible layer instead of growing more platform branches there.
- New compatible platforms should extend the platform helper/context path first, then reuse the shared gateway/service primitives, instead of branching more platform special cases inside OpenAI-only code.

Grok-specific follow-on rule:

- Shared compatible transport does not own Grok product semantics.
- Grok model capability, tier, and protocol metadata live in `backend/internal/pkg/grok/registry.go`.
- Grok quota window defaults and pool inference live in `backend/internal/pkg/grok/quota.go`, based on the provider-owned Grok control-plane semantics rather than OpenAI extras.
- Grok runtime account state lives under `account.extra.grok` and is read through `backend/internal/service/grok_account_state.go`.
- Grok state writes should flow through `backend/internal/service/grok_account_state_updates.go` so admin create/update/import paths deep-merge `extra.grok` and preserve quota/tier/runtime snapshots instead of replacing the whole subtree.
- Provider-owned Grok runtime probe/sync patch building lives in `backend/internal/service/grok_account_runtime_state.go`; shared compatible probe entrypoints should call into that Grok module instead of assembling `extra.grok` patches inside OpenAI- or compatible-named services.
- Provider-owned Grok probe/sync persistence should enter `backend/internal/service/grok_account_state_service.go`, which owns `UpdateExtra` writes for normalized `extra.grok` probe snapshots. Shared probe callers such as `backend/internal/service/compatible_gateway_probe_state.go` should dispatch there for Grok instead of writing probe patches inline.
- Provider-owned Grok tier/quota freshness should now be refreshed by `backend/internal/service/grok_quota_sync_service.go`, which periodically canonicalizes `tier`, `quota_windows`, and `sync_state.last_sync_at` through `GrokAccountStateService` instead of leaving selector freshness dependent on imported snapshots.
- Provider-owned compatible probe/runtime writes now dispatch through `backend/internal/service/compatible_gateway_probe_state.go`; Grok probe results must persist a normalized full `extra.grok` subtree there instead of extending OpenAI OAuth-only snapshot writers.
- Real Grok request outcomes now persist through the provider-owned patch builder in `backend/internal/service/grok_account_runtime_state.go`, which merges `runtime_state`, `sync_state.last_runtime_*`, and capability/model evidence into the normalized `extra.grok` subtree before writing through `UpdateExtra`.
- Shared compatible request handlers should emit runtime success/failure/failover signals through `backend/internal/service/compatible_gateway_runtime_feedback.go` and `backend/internal/handler/compatible_gateway_runtime_feedback.go` rather than constructing `GrokRuntimeFeedbackInput` directly inside `openai_*` handler files.
- Runtime success/failure/failover signals must not call `UpdateExtra(..., {"grok": ...})` with ad hoc partial payloads or rely on repo-specific shortcut writers for scheduler-relevant Grok state, because shallow top-level JSONB merges would overwrite sibling Grok state and bypass the normalized provider-owned patch semantics.
- Compatible text forwarding feedback for `/responses`, `/chat/completions`, and `/messages` is owned by `backend/internal/service/compatible_gateway_text_runtime.go`. Handler orchestration must not emit a second runtime feedback write after delegated text forwarding; passthrough handlers that do not enter the text runtime may still emit feedback at the handler boundary.
- `backend/internal/service/compatible_gateway_text_runtime.go` must stay an executor facade: provider-specific OpenAI forwarding lives in `OpenAITextExecutor`, while the runtime itself holds only the neutral `CompatibleGatewayTextExecutor` contract plus explicit feedback ownership metadata.
- Runtime composition must use explicit feedback ownership (`OwnsCompatibleGatewayRuntimeFeedback`) instead of detecting concrete executor types. Grok compatible runtime must decide whether to persist `extra.grok.runtime_state` from that ownership flag, not from `*CompatibleGatewayTextRuntime` type assertions.
- Scheduler-relevant Grok capability/tier/quota enrichment still uses full normalized `extra.grok` writes, but only when provider-owned Grok logic learns genuinely new capability state from probes or successful traffic.
- Grok runtime selection eligibility, candidate filtering, and requested-model availability live in `backend/internal/service/grok_account_selector.go`.
- Grok selection and available-model derivation must consult that provider-owned registry/state layer before falling back to generic compatible `model_mapping` behavior.
- Admin/API-key probe flows should enter through compatible-gateway helpers (`testCompatibleGatewayAccountConnection`, `testCompatibleGatewayAPIKeyConnection`) before any provider-owned transport specialization is applied.
- Shared compatible upstream endpoint/auth builders now live in `backend/internal/service/compatible_upstream_target.go`; use those transport primitives for compatible models discovery, probes, passthrough routing, and Responses endpoint derivation instead of reintroducing `openai_*` transport helpers for Grok/OpenAI siblings.
- The shared compatible runtime only owns Grok `apikey` and `upstream` execution today. Grok `session` remains a provider-owned transport boundary and must stay out of shared-compatible selection/model-availability checks until a dedicated `GrokSessionTransport` exists.
- Provider-owned Grok auth/transport resolution now lives in `backend/internal/service/grok_transport.go`. Grok account-test and future execution paths should parse `credentials.session_token` into the Grok Web `Cookie` header there instead of teaching `OpenAIGatewayService.GetAccessToken` about Grok session auth.
- Grok session probes now enter through `backend/internal/service/account_test_service_grok.go`, which owns the provider-specific Grok Web probe request and persists normalized `extra.grok` probe state without falling back to `testCompatibleGatewayAPIKeyConnection`.
- Provider-owned Grok app-chat request shaping now lives in `backend/internal/service/grok_session_text_request.go`. Future Grok `session` runtime work should reuse that builder for `/responses`, `/chat/completions`, and `/messages`, then add a matching provider-owned Grok stream/response adapter before relaxing selector/runtime eligibility.
- Provider-owned Grok text ownership is now split across `backend/internal/service/grok_gateway_service.go`, `backend/internal/service/grok_text_runtime.go`, `backend/internal/service/grok_compatible_runtime.go`, and `backend/internal/service/grok_session_runtime.go`: `backend/internal/handler/grok_gateway_handler.go` binds Grok text runtime, replays the request body, and gives the Grok service first chance to select and execute against Grok accounts for `/responses`, `/chat/completions`, and `/messages`. Compatible `apikey` and `upstream` accounts flow through the Grok-owned compatible runtime wrapper around the neutral compatible executor, while Grok Web `session` traffic terminates in the dedicated session runtime plus `backend/internal/service/grok_session_text_response.go` instead of re-entering OpenAI-named runtime ownership.
- Grok text account choice is now provider-owned too: `backend/internal/service/grok_text_runtime.go` should ask `backend/internal/service/grok_account_selector.go` for the best candidate, and that selector should score Grok accounts with provider state from `extra.grok` (`tier`, `capabilities`, `quota_windows`, `sync_state`, `runtime_state`) plus request-scoped load snapshots. Do not fall back to the generic priority/LRU helper for Grok text once provider state is available.
- Grok scheduling freshness now has two provider-owned background sources: `backend/internal/service/grok_quota_sync_service.go` refreshes tier/quota/sync timestamps for all Grok accounts, while `backend/internal/service/grok_capability_probe_service.go` probes only unknown-tier Grok `apikey`/`upstream` accounts to seed `last_probe_*` and baseline capability signals without routing them back through OpenAI services.
- Grok quota sync should also repair the narrow "single chat probe" capability snapshot once tier becomes known. If a previous Grok probe only learned one chat model, `backend/internal/service/grok_quota_sync_service.go` should widen that provider state from the Grok registry and tier rules so later selection does not get stuck on a stale one-model allowlist.
- Provider-owned Grok runtime failure semantics now live in `backend/internal/service/grok_runtime_error_classifier.go`. `backend/internal/service/grok_account_runtime_state.go` must persist normalized `last_fail_class`, `last_fail_scope`, and `selection_cooldown_*` state there, and `backend/internal/service/grok_account_selector.go` should treat those fields as the Grok scheduling authority instead of inferring penalties directly from shared/OpenAI-style HTTP status handling.
- Grok media ownership now begins in `backend/internal/service/grok_media_service.go`, not in the shared passthrough handler. `backend/internal/handler/grok_gateway_handler.go` and `backend/internal/server/routes/gateway_dispatcher.go` should route Grok `/images/*` and `/videos*` traffic there first so Grok can own account choice and follow-up binding before any shared compatible transport primitive is used.
- Provider-owned Grok media response rendering and cache policy now live in `backend/internal/service/grok_media_asset_service.go` plus `backend/internal/service/grok_media_settings.go`. Image/video output format (`local_url`, `upstream_url`, `markdown`, `base64`, `html` as applicable), proxy enablement, and cache retention must resolve through `SettingService` keys instead of being hardcoded inside the media rewriter.
- Grok async video replay state now persists in `backend/internal/repository/grok_video_job_repo.go` backed by `grok_video_jobs`. `/videos` creation must record the originating Grok account, and later `/videos/:id` or `/videos/:id/content` requests must resolve that binding instead of re-running scheduler selection.
- Shared compatible text forwarding for `/responses`, `/chat/completions`, and `/messages` now lives behind `backend/internal/service/compatible_gateway_text_runtime.go`. `OpenAIGatewayService` delegates into that neutral runtime, and `GrokGatewayService` should depend on the neutral runtime instead of holding an `OpenAIGatewayService` reference directly.
- Shared compatible text request execution for `/responses`, `/chat/completions`, and `/messages` now flows through `backend/internal/service/compatible_text_execution_kernel.go`. That kernel owns account selection, explicit model fallback, slot acquisition, pool-mode session hash creation, failover, scheduler result feedback, account-switch metrics, and Codex usage snapshot writes.
- OpenAI-compatible HTTP passthrough now enters the same execution kernel through `NewCompatiblePassthroughExecutionKernel`; passthrough-specific account eligibility and per-attempt runtime feedback are hooks, not a second handler-side selection/failover loop.
- OpenAI Responses WebSocket ingress initial routing now enters `backend/internal/service/openai_ws_ingress_selection_kernel.go`. The service kernel owns scheduler selection, explicit model fallback, initial account slot acquisition, sticky-session binding, and Codex scheduling observation; the handler only maps outcomes to websocket close codes and wires the long-lived proxy hooks.
- `backend/internal/handler/compatible_gateway_text_flow.go` and `backend/internal/handler/openai_gateway_handler_passthrough.go` should stay HTTP adapters around that kernel: they may validate bodies, render route-specific errors, record usage, and pass logging/latency/feedback hooks, but they must not rebuild selection/failover loops.
- `backend/internal/service/selection_kernel.go` is the shared selection entrypoint for provider/protocol/model/session/transport intent. Existing OpenAI scheduler calls and native GatewayService load-aware selection should enter through that kernel instead of growing more handler-side scheduling branches.
- Generic runtime failover state for native Gemini/Anthropic/Antigravity routes lives in `backend/internal/service/runtime_failover_state.go`. Handler files may keep route-specific selection, forwarding, and HTTP error rendering during the migration, but same-account retry limits, failed-account exclusion state, bound-session rate-limit preservation, force-cache-billing flags, and single-account 503 backoff must stay service-owned.
- Selection failure outcomes for native routes are also owned by `backend/internal/service/runtime_failover_state.go`. Handler files should call `HandleSelectionError` and render based on `initial_unavailable`, `retry`, `canceled`, or `exhausted`; they should not duplicate failed-account-count checks or call `HandleSelectionExhausted` directly in route loops.
- Forward error failover classification for native routes is also part of `backend/internal/service/runtime_failover_state.go`. Handler files should pass forward errors and response-started state into `HandleForwardError`; they should not repeat `UpstreamFailoverError` detection, response-started exhaustion checks, or direct `HandleFailoverError` calls around every `Forward*` branch.
- Runtime session preparation for native routes lives in `backend/internal/service/runtime_session.go`. Parsing the already-validated protocol body for scheduling metadata, attaching `SessionContext`, generating the sticky session hash/key, and prefetching sticky-session context metadata should flow through `PrepareRuntimeSession` / `PrefetchRuntimeStickySession`; handlers should only provide client IP/User-Agent/API key fields and write the returned context back to the request. Native `/v1/messages`, `count_tokens`, Gemini native `/v1beta/models/*`, and OpenAI-compatible native Anthropic text flows should not rebuild this session/sticky setup inline. Provider-specific session identifiers, such as Gemini CLI session hashes, should be passed as an explicit `SessionHash` override into the same service entrypoint instead of bypassing it.
- Native `count_tokens` account choice lives in `backend/internal/service/runtime_count_tokens_selection.go`. This path intentionally keeps the no-account-concurrency behavior of count_tokens, but the failed-account exclusion set and RPM admission retry loop are service-owned. Handlers should log returned admission events and render the final selection error instead of calling `SelectAccountForModelWithExclusions` directly.
- Account slot acquisition after selection is owned by `backend/internal/service/runtime_account_slot_acquisition.go`. Handlers may provide protocol-specific wait hooks because SSE/WebSocket ping behavior is HTTP-adapter work, but queue-vs-wait decisions, wait-count cleanup, and post-wait sticky binding must use the service kernel instead of duplicated handler blocks.
- Forward admission control is owned by `backend/internal/service/runtime_admission_control.go`. Handler helpers may log route-specific fields, but RPM reservation, window-cost reservation, sticky-overflow decisions, and sticky binding cleanup on admission denial must stay service-owned.
- Admission-denied cleanup is owned by `backend/internal/service/runtime_admission_cleanup.go`. Handler files may still decide which protocol error to render, but account release, queue release, upstream-accepted callback clearing, and failed-account marking after RPM/window-cost denial should flow through that helper.
- Runtime forward attempt cleanup is owned by `backend/internal/service/runtime_forward_attempt.go`. Native handlers may pass provider-specific forward functions and writer-size probes, but account slot release, user-message queue release, upstream-accepted callback clearing, and window-cost reservation release on pre-response errors must flow through this kernel instead of being repeated around each `Forward*` call.
- Native provider forward dispatch is owned by `backend/internal/service/native_gateway_runtime.go`. Handler files should pass provider/protocol/account request envelopes there instead of choosing between `GatewayService`, `GeminiMessagesCompatService`, and `AntigravityGatewayService` `Forward*` methods inline. Native `count_tokens` uses the same runtime boundary through `ForwardCountTokens`; handlers should not call `GatewayService.ForwardCountTokens` directly.
- OpenAI-compatible native Anthropic `/v1/chat/completions` and `/v1/responses` route bodies now share `backend/internal/handler/gateway_handler_openai_compatible_text_flow.go` for session hash setup, account selection, admission retry, native runtime forwarding, failover rendering callbacks, and usage recording. Route files should keep request validation and protocol-specific JSON error shape only; do not copy the selection/admission/forward loop back into each endpoint.
- OpenAI pool failover decisions live in `backend/internal/service/openai_pool_failover_policy.go`; retry/switch/exhaust policy must not be reimplemented in handler loops or handler-local adapter enums.
- `backend/internal/handler/compatible_gateway_text_handler.go` should delegate actual upstream forwarding through `service.CompatibleGatewayTextRuntime` (`ForwardResponses`, `ForwardChatCompletions`, `ForwardMessages`). Route handlers still own request validation and protocol-specific request shaping, but the forwarding boundary belongs to the neutral runtime.
- Shared compatible handler dispatch now mirrors that service split: `CompatibleGatewayHandler` routes shared compatible traffic into `CompatibleGatewayRuntimeHandler`, and Grok consumes that same shared runtime after its provider-owned session check. Do not reintroduce direct `CompatibleGatewayHandler -> OpenAIGatewayHandler` control-plane coupling.
- The provider-owned Grok text runtime should keep a single upstream parse path and perform protocol conversion at the Grok boundary: Grok session deltas become Responses events first, then `/chat/completions` and `/messages` are derived from those Responses events through `apicompat` converters. Do not duplicate separate Grok Web parsers per protocol family.
- Grok text-route ownership is now request-scoped: `backend/internal/handler/grok_gateway_handler.go` sets a Grok-only session-runtime context flag for `/responses`, `/chat/completions`, and `/messages`, and `backend/internal/service/grok_account_selector.go` only admits `AccountTypeSession` when that flag is present. Non-text compatible routes, passthrough, and websocket flows must leave that flag unset so sticky/session reuse cannot leak Grok Web accounts back into shared compatible runtime paths.
- This keeps `openai` as a protocol sibling instead of the place where Grok capability/tier logic accumulates.

### OpenAI scheduler capability index

OpenAI scheduler model capability indexing must distinguish capability
declarations from model mapping rules:

- `account.extra.supported_models` is the OpenAI model capability source.
- Missing, empty-array, or empty-string `supported_models` means unknown capability and must place the account in `model_any`; it must not make the account unavailable for newly added models.
- Non-empty `supported_models` may be indexed as `model_exact` / `model_pattern` and may restrict scheduling.
- `credentials.model_mapping` remains a mapping or alias rule; it must not be treated as an OpenAI capability declaration when building scheduler indexes.

### Detached service tasks

Use `backend/internal/service/async_task.go` for fire-and-forget service work that should not block the request path.

Current reference integrations:

- `backend/internal/service/admin_service_account.go`
- `backend/internal/service/admin_service_proxy.go`
- `backend/internal/service/admin_service_group_delete.go`

Why:

- panic recovery is centralized
- logging shape is consistent
- tests can wait on the returned completion channel when needed

### Admin handler service ports

Admin HTTP handlers should depend on the narrow interfaces in
`backend/internal/handler/admin/ports.go` instead of directly coupling every handler to
`service.AdminService`.

Current rule:

- provider functions may still accept `service.AdminService` so Wire generation stays unchanged;
- handler constructors should accept the smallest handler-local port (`accountAdminService`, `proxyAdminService`, `redeemAdminService`, etc.);
- add a method to a narrow port only when that handler actually calls it;
- do not use the full `service.AdminService` interface as a shortcut for tests or new handlers.

## Next shortlist

The next structural cleanup should focus on:

1. Continuing account modal section extraction so provider-specific form blocks move out of the top-level modal files.
2. Reducing overlap between legacy service entry points and newer admin-specific service flows on the backend.
3. Extracting account modal state hydration/reset helpers now that create/edit/bulk payload construction shares the same helper/profile layer.
