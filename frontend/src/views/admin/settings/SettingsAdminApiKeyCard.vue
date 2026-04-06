<template>
  <div class="card">
    <div class="settings-admin-api-key-card__header settings-admin-api-key-card__header-spacing">
      <h2 class="settings-admin-api-key-card__title text-lg font-semibold">
        {{ t('admin.settings.adminApiKey.title') }}
      </h2>
      <p class="settings-admin-api-key-card__description mt-1 text-sm">
        {{ t('admin.settings.adminApiKey.description') }}
      </p>
    </div>
    <div class="settings-admin-api-key-card__body space-y-4">
      <div class="settings-admin-api-key-card__warning settings-admin-api-key-card__panel">
        <div class="flex items-start">
          <Icon
            name="exclamationTriangle"
            size="md"
            class="settings-admin-api-key-card__warning-icon mt-0.5 flex-shrink-0"
          />
          <p class="settings-admin-api-key-card__warning-text ml-3 text-sm">
            {{ t('admin.settings.adminApiKey.securityWarning') }}
          </p>
        </div>
      </div>

      <div v-if="loading" class="settings-admin-api-key-card__loading flex items-center gap-2">
        <div class="settings-admin-api-key-card__spinner h-4 w-4 animate-spin rounded-full border-b-2"></div>
        {{ t('common.loading') }}
      </div>

      <div v-else-if="!exists" class="flex items-center justify-between">
        <span class="settings-admin-api-key-card__description">
          {{ t('admin.settings.adminApiKey.notConfigured') }}
        </span>
        <button
          type="button"
          :disabled="operating"
          class="btn btn-primary btn-sm"
          @click="$emit('create')"
        >
          <svg
            v-if="operating"
            class="mr-1 h-4 w-4 animate-spin"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            ></circle>
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
          {{
            operating
              ? t('admin.settings.adminApiKey.creating')
              : t('admin.settings.adminApiKey.create')
          }}
        </button>
      </div>

      <div v-else class="space-y-4">
        <div class="flex items-center justify-between">
          <div>
            <label class="settings-admin-api-key-card__field-label mb-1 block text-sm font-medium">
              {{ t('admin.settings.adminApiKey.currentKey') }}
            </label>
            <code class="settings-admin-api-key-card__code settings-admin-api-key-card__code-spacing font-mono text-sm">
              {{ maskedKey }}
            </code>
          </div>
          <div class="flex gap-2">
            <button
              type="button"
              :disabled="operating"
              class="btn btn-secondary btn-sm"
              @click="$emit('regenerate')"
            >
              {{
                operating
                  ? t('admin.settings.adminApiKey.regenerating')
                  : t('admin.settings.adminApiKey.regenerate')
              }}
            </button>
            <button
              type="button"
              :disabled="operating"
              class="btn btn-secondary settings-admin-api-key-card__delete-button btn-sm"
              @click="$emit('delete')"
            >
              {{ t('admin.settings.adminApiKey.delete') }}
            </button>
          </div>
        </div>

        <div v-if="newKey" class="settings-admin-api-key-card__success settings-admin-api-key-card__panel space-y-3">
          <p class="settings-admin-api-key-card__success-title text-sm font-medium">
            {{ t('admin.settings.adminApiKey.keyWarning') }}
          </p>
          <div class="flex items-center gap-2">
            <code class="settings-admin-api-key-card__new-key settings-admin-api-key-card__new-key-spacing flex-1 select-all break-all font-mono text-sm">
              {{ newKey }}
            </code>
            <button
              type="button"
              class="btn btn-primary btn-sm flex-shrink-0"
              @click="$emit('copy')"
            >
              {{ t('admin.settings.adminApiKey.copyKey') }}
            </button>
          </div>
          <p class="settings-admin-api-key-card__success-hint text-xs">
            {{ t('admin.settings.adminApiKey.usage') }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  loading: boolean
  exists: boolean
  maskedKey: string
  operating: boolean
  newKey: string
}>()

defineEmits<{
  create: []
  regenerate: []
  delete: []
  copy: []
}>()

const { t } = useI18n()
</script>

<style scoped>
.settings-admin-api-key-card__header {
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.settings-admin-api-key-card__header-spacing {
  padding: var(--theme-settings-card-header-padding-y) var(--theme-settings-card-header-padding-x);
}

.settings-admin-api-key-card__body {
  padding: var(--theme-settings-card-body-padding);
}

.settings-admin-api-key-card__panel {
  border-radius: var(--theme-settings-card-panel-radius);
  padding: var(--theme-settings-card-panel-padding);
}

.settings-admin-api-key-card__title,
.settings-admin-api-key-card__field-label {
  color: var(--theme-page-text);
}

.settings-admin-api-key-card__description,
.settings-admin-api-key-card__loading {
  color: var(--theme-page-muted);
}

.settings-admin-api-key-card__warning {
  border: 1px solid color-mix(in srgb, rgb(var(--theme-warning-rgb)) 24%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 10%, var(--theme-surface));
}

.settings-admin-api-key-card__warning-icon,
.settings-admin-api-key-card__warning-text {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.settings-admin-api-key-card__spinner {
  border-color: color-mix(in srgb, var(--theme-card-border) 70%, transparent);
  border-bottom-color: var(--theme-accent);
}

.settings-admin-api-key-card__code {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: var(--theme-page-text);
}

.settings-admin-api-key-card__code-spacing {
  border-radius: var(--theme-settings-inline-button-radius);
  padding: var(--theme-settings-code-padding-y) var(--theme-settings-code-padding-x);
}

.settings-admin-api-key-card__delete-button {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 78%, var(--theme-page-text));
}

.settings-admin-api-key-card__delete-button:hover {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 92%, var(--theme-page-text));
}

.settings-admin-api-key-card__success {
  border: 1px solid color-mix(in srgb, rgb(var(--theme-success-rgb)) 24%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 10%, var(--theme-surface));
}

.settings-admin-api-key-card__success-title,
.settings-admin-api-key-card__success-hint {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.settings-admin-api-key-card__new-key {
  border: 1px solid color-mix(in srgb, rgb(var(--theme-success-rgb)) 28%, var(--theme-card-border));
  background: var(--theme-surface);
  color: var(--theme-page-text);
}

.settings-admin-api-key-card__new-key-spacing {
  border-radius: var(--theme-settings-inline-button-radius);
  padding: var(--theme-settings-new-key-padding-y) var(--theme-settings-new-key-padding-x);
}
</style>
