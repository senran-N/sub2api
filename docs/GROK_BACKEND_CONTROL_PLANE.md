# Grok Backend Control Plane

This document records the current Grok-owned backend and admin seams that are intentionally separate from the generic compatible-gateway path.

## Session Batch Import

- Endpoint: `POST /api/v1/admin/accounts/grok/session/batch-import`
- Handler: `backend/internal/handler/admin/account_handler.go`
- Admin entry: `AdminService.BatchImportGrokSessionAccounts`
- Core service: `backend/internal/service/grok_session_batch_import_service.go`

Behavior:

- Input is a single `raw_input` blob with one token or cookie string per line.
- Server-side parsing accepts bare `sso` values, `sso=...`, or full cookie headers.
- Import validation requires a normalized cookie set that contains `sso`.
- Storage still uses `credentials.session_token`; there is no separate `sso_token` field.
- Fingerprints are derived from the normalized cookie header and persisted at `extra.grok.auth_fingerprint`.
- Responses expose only masked fingerprints, line numbers, and summary counts.
- Account creation is delegated to the existing `CreateAccount` path so validation, group binding, and extra normalization stay shared with single-account creation.

## Provider-Owned Grok Transport

- Selector: `backend/internal/service/grok_account_selector.go`
- Transport resolver: `backend/internal/service/grok_transport.go`
- Session HTTP helper: `backend/internal/service/grok_session_http.go`

Behavior:

- Text, `/messages`, image, and video routing all resolve through Grok-owned transport semantics.
- API key and upstream accounts still forward through compatible upstream `/v1/...` HTTP.
- Session accounts use normalized Grok Web session cookies against `https://grok.com/v1/...`.
- Session capability probing and admin account testing reuse the same browser-style request contract instead of duplicating header logic.

## Provider-Owned Grok Media

- Media service: `backend/internal/service/grok_media_service.go`
- Media asset service: `backend/internal/service/grok_media_asset_service.go`

Behavior:

- Grok media selection uses the Grok selector for both compatible accounts and session accounts.
- Session accounts are allowed into media scheduling through a dedicated Grok media runtime allowance instead of the old shared-runtime-only gate.
- Video create, poll, and content follow-up stay bound to the original account ID stored in the Grok video job record.
- Media asset replay reuses the same Grok transport auth material, so follow-up fetches stay account-bound.

## Runtime Settings Control Plane

- Settings keys: `backend/internal/service/domain_constants.go`
- View contract: `backend/internal/service/settings_view.go`
- Admin handler DTO contract: `backend/internal/handler/dto/settings.go`
- Shared normalization seam: `backend/internal/service/grok_runtime_settings.go`
- Live consumers:
  - `backend/internal/service/grok_quota_sync_service.go`
  - `backend/internal/service/grok_capability_probe_service.go`

Behavior:

- These Grok settings are live admin-owned control-plane inputs:
  - `fallback_model_grok`
  - `grok_image_output_format`
  - `grok_video_output_format`
  - `grok_media_proxy_enabled`
  - `grok_media_cache_retention_hours`
  - `grok_quota_sync_interval_seconds`
  - `grok_capability_probe_interval_seconds`
- `SettingService.GetGrokRuntimeSettings(ctx)` is the single normalization seam for the Grok runtime intervals. It applies defaults plus range clamping before background services consume them.
- The quota-sync and capability-probe loops resolve the current interval on each cycle, so admin changes take effect without introducing a parallel runtime config path.
- New Grok settings must only be added when a real backend consumer exists. Do not add dead control-plane fields with no live service seam behind them.
- `grok_session_validity_check_interval_seconds` and `grok_video_timeout_seconds` remain intentionally deferred until their owning backend consumer exists.

## Capability And Tier Refresh

- Probe service: `backend/internal/service/grok_capability_probe_service.go`
- Probe state persistence: `backend/internal/service/grok_account_state.go`

Behavior:

- Capability probing covers Grok session accounts as well as compatible and API key accounts.
- Probe gating requires valid auth material for the selected transport before an account enters the periodic probe loop.
- Unknown-tier accounts bootstrap through a shared probe candidate list rather than a single basic chat probe.
- Known `heavy` and `super` accounts also refresh through the same tier-representative high-tier probe candidate before falling back to the default basic probe.
- Successful probe persistence widens the stored Grok capability snapshot to the selector-visible baseline whenever tier is already known or can be inferred from the probe result. This keeps media scheduling and selector state aligned.

## Admin Model Catalog Contract

- Endpoint: `GET /api/v1/admin/model-catalog?platform=grok`
- Backend source of truth: `backend/internal/pkg/grok/registry.go`
- Frontend consumer seam: `frontend/src/composables/useModelWhitelist.ts`

Behavior:

- Grok model options, preset mappings, whitelist IDs, and fallback selections are registry-driven from the backend catalog instead of a static frontend `grokModels` list.
- `getModelsByPlatform('grok')`, `getModelOptionsByPlatform('grok')`, and `getPresetMappingsByPlatform('grok')` are pure cache reads.
- Grok-aware UI surfaces must explicitly call `ensureModelCatalogLoaded('grok')` before relying on Grok catalog-backed options.
- Failed Grok catalog fetches remain retryable; a transient admin API failure must not permanently freeze the SPA to an empty Grok catalog.

## Admin Runtime Visibility Contract

- Shared parser: `frontend/src/utils/grokAccountRuntime.ts`
- List surfaces:
  - `frontend/src/views/admin/accounts/AccountNameCell.vue`
  - `frontend/src/views/admin/accounts/AccountPlatformTypeCell.vue`
  - `frontend/src/components/account/AccountCapacityCell.vue`
- Detail surface: `frontend/src/components/account/EditAccountModal.vue`

Behavior:

- Admin-facing Grok runtime display is normalized through `getGrokAccountRuntime(account)` rather than ad hoc `account.extra.grok` reads in each component.
- The account list exposes auth fingerprint, normalized tier, quota windows, capabilities, sync/probe recency, and recent probe/runtime error summary.
- The edit modal reuses the same normalized runtime contract for a read-only Grok runtime summary block.
- If the Grok runtime payload shape changes, update the shared parser first and keep list/detail surfaces as consumers only.

## Verification Anchors

Run backend checks from `backend/`:

```bash
go test ./internal/handler/admin -run 'TestSettingHandlerGetModelCatalog_Grok|TestSettingHandlerGetModelCatalog_UnsupportedPlatform'
go test ./internal/service -run 'TestSettingService_UpdateSettings_PersistsGrokFallbackModel|TestSettingService_UpdateSettings_NormalizesGrok(Runtime|Media)Settings'
```

Run frontend checks from `frontend/`:

```bash
pnpm vitest run src/composables/__tests__/useModelWhitelist.grokCatalog.spec.ts src/views/admin/__tests__/settingsForm.spec.ts src/views/admin/__tests__/settingsView.spec.ts src/views/admin/__tests__/useSettingsViewForm.spec.ts
pnpm vitest run src/components/account/__tests__/CreateAccountModal.spec.ts src/components/account/__tests__/EditAccountModal.spec.ts src/views/admin/__tests__/accountsTableCells.spec.ts src/utils/__tests__/grokAccountRuntime.spec.ts
```
