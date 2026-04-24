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
- create/edit/bulk account mutation payload assembly lives in `frontend/src/components/account/accountMutationPayload.ts`; modal files should call that helper for final request payloads instead of rebuilding JSONB merge sentinels, quota overlays, OpenAI runtime fields, or provider-specific credential mutations locally;
- lower-level credential and form helpers remain in `accountModalShared.ts`, `createAccountModalHelpers.ts`, `editAccountModalHelpers.ts`, and `credentialsBuilder.ts`;
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
