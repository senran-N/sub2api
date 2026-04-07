# Admin View Structure

This directory uses a feature-first split with a thin root layer.

## Root Layer

Keep only:

- page entry views such as `SettingsView.vue` and `GroupsView.vue`
- cross-feature composables reused by multiple admin pages
- shared tests that verify page-level composition

Do not place feature-private form builders, view-model helpers, or section-only utilities here.

## Feature Directories

Feature-specific logic should live with the feature UI:

- `dashboard/`
  - stats and chart sections
  - dashboard-only data helpers
- `dataManagement/`
  - Sora profile cards and drawers
  - data-management-only form helpers
  - data-management-only composables
- `channels/`
  - channel form serializers and validators
  - channel-only style and table helpers
  - channels-only feature tests
- `settings/`
  - cards and tabs
  - settings form helpers
  - settings-only composables
- `groups/`
  - dialogs and sections
  - group form helpers
  - group list and table helpers
  - routing rule state helpers
- `proxies/`
  - proxy form, list, quality, and clipboard helpers
  - proxy-only composables and table fragments
- `accounts/`
  - table cells and toolbar fragments
  - account list and page helpers
  - account-only composables
- `subscriptions/`
  - dialogs and cells
  - subscription form helpers
  - subscription-only composables
- `usage/`
  - chart and table controls
  - usage state helpers
  - usage-only composables
- `backup/`
  - backup cards and guide modal
  - backup state helpers
  - backup-only composables
- `users/`
  - table cells and toolbar fragments
  - user table helpers
  - users-only composables
- `promocodes/`
  - dialog and badge fragments
  - promo code form helpers
  - promocodes-only composables
- `redeem/`
  - dialog and badge fragments
  - redeem form helpers
  - redeem-only composables
- `announcements/`
  - dialog and badge fragments
  - announcement form helpers
  - announcements-only composables

When a helper is only consumed by one admin feature, colocate it inside that feature directory even if it is not a Vue component.

## Form State Rule

Current admin forms still use parent-owned reactive objects that are edited by child sections.

- Keep mutation helpers near the owning feature form module.
- Prefer semantic events for list operations such as add/remove/move.
- If a future refactor introduces stricter one-way data flow, do it per feature rather than mixing patterns in the same feature tree.
