<template>
  <div class="card">
    <div class="card-header">
      <h2 class="settings-purchase-card__title text-lg font-semibold">
        {{ t('admin.settings.purchase.title') }}
      </h2>
      <p class="settings-purchase-card__description mt-1 text-sm">
        {{ t('admin.settings.purchase.description') }}
      </p>
    </div>
    <div class="settings-purchase-card__body">
      <div class="flex items-center justify-between">
        <div>
          <label id="settings-purchase-enabled-label" class="settings-purchase-card__title font-medium">
            {{ t('admin.settings.purchase.enabled') }}
          </label>
          <p class="settings-purchase-card__description text-sm">
            {{ t('admin.settings.purchase.enabledHint') }}
          </p>
        </div>
        <Toggle
          id="settings-purchase-enabled-toggle"
          v-model="form.purchase_subscription_enabled"
          name="purchase_subscription_enabled"
          :aria-label="t('admin.settings.purchase.enabled')"
          aria-labelledby="settings-purchase-enabled-label"
        />
      </div>

      <div>
        <label
          for="settings-purchase-url"
          class="settings-purchase-card__field-label mb-2 block text-sm font-medium"
        >
          {{ t('admin.settings.purchase.url') }}
        </label>
        <input
          id="settings-purchase-url"
          v-model="form.purchase_subscription_url"
          name="purchase_subscription_url"
          type="url"
          class="input font-mono text-sm"
          :placeholder="t('admin.settings.purchase.urlPlaceholder')"
        />
        <p class="settings-purchase-card__description mt-1.5 text-xs">
          {{ t('admin.settings.purchase.urlHint') }}
        </p>
        <p class="settings-purchase-card__warning mt-2 text-xs">
          {{ t('admin.settings.purchase.iframeWarning') }}
        </p>
      </div>

      <div class="flex items-center gap-2 text-sm">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="settings-purchase-card__doc-icon h-4 w-4 shrink-0"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
          />
        </svg>
        <a
          href="https://raw.githubusercontent.com/senran-N/sub2api/main/docs/ADMIN_PAYMENT_INTEGRATION_API.md"
          target="_blank"
          rel="noopener noreferrer"
          class="settings-purchase-card__doc-link"
          download="ADMIN_PAYMENT_INTEGRATION_API.md"
        >
          {{ t('admin.settings.purchase.integrationDoc') }}
        </a>
        <span class="settings-purchase-card__doc-divider">-</span>
        <span class="settings-purchase-card__description text-xs">
          {{ t('admin.settings.purchase.integrationDocHint') }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import type { SettingsPurchaseFields } from './settingsForm'

defineProps<{
  form: SettingsPurchaseFields
}>()

const { t } = useI18n()
</script>

<style scoped>
.settings-purchase-card__title,
.settings-purchase-card__field-label {
  color: var(--theme-page-text);
}

.settings-purchase-card__description,
.settings-purchase-card__doc-icon,
.settings-purchase-card__doc-divider {
  color: var(--theme-page-muted);
}

.settings-purchase-card__warning {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.settings-purchase-card__doc-link {
  color: var(--theme-accent);
  text-decoration: none;
}

.settings-purchase-card__doc-link:hover {
  text-decoration: underline;
}

.settings-purchase-card__body {
  padding: var(--theme-settings-card-panel-padding);
  display: flex;
  flex-direction: column;
  gap: var(--theme-settings-card-body-padding);
}
</style>
