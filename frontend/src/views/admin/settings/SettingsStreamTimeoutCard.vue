<template>
  <div class="card">
    <div class="settings-stream-timeout-card__header">
      <h2 class="settings-stream-timeout-card__title text-lg font-semibold">
        {{ t('admin.settings.streamTimeout.title') }}
      </h2>
      <p class="settings-stream-timeout-card__description mt-1 text-sm">
        {{ t('admin.settings.streamTimeout.description') }}
      </p>
    </div>
    <div class="settings-stream-timeout-card__body space-y-5">
      <div v-if="loading" class="settings-stream-timeout-card__loading flex items-center gap-2">
        <div class="settings-stream-timeout-card__spinner h-4 w-4 animate-spin rounded-full border-b-2"></div>
        {{ t('common.loading') }}
      </div>

      <template v-else>
        <div class="flex items-center justify-between">
          <div>
            <label class="settings-stream-timeout-card__label font-medium">
              {{ t('admin.settings.streamTimeout.enabled') }}
            </label>
            <p class="settings-stream-timeout-card__description text-sm">
              {{ t('admin.settings.streamTimeout.enabledHint') }}
            </p>
          </div>
          <Toggle
            v-model="form.enabled"
            :aria-label="t('admin.settings.streamTimeout.enabled')"
          />
        </div>

        <div
          v-if="form.enabled"
          class="settings-stream-timeout-card__section space-y-4 pt-4"
        >
          <div>
            <label class="settings-stream-timeout-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.streamTimeout.action') }}
            </label>
            <select v-model="form.action" class="input w-64">
              <option value="temp_unsched">
                {{ t('admin.settings.streamTimeout.actionTempUnsched') }}
              </option>
              <option value="error">
                {{ t('admin.settings.streamTimeout.actionError') }}
              </option>
              <option value="none">
                {{ t('admin.settings.streamTimeout.actionNone') }}
              </option>
            </select>
            <p class="settings-stream-timeout-card__description mt-1.5 text-xs">
              {{ t('admin.settings.streamTimeout.actionHint') }}
            </p>
          </div>

          <div v-if="form.action === 'temp_unsched'">
            <label class="settings-stream-timeout-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.streamTimeout.tempUnschedMinutes') }}
            </label>
            <input
              v-model.number="form.temp_unsched_minutes"
              type="number"
              min="1"
              max="60"
              class="input w-32"
            />
            <p class="settings-stream-timeout-card__description mt-1.5 text-xs">
              {{ t('admin.settings.streamTimeout.tempUnschedMinutesHint') }}
            </p>
          </div>

          <div>
            <label class="settings-stream-timeout-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.streamTimeout.thresholdCount') }}
            </label>
            <input
              v-model.number="form.threshold_count"
              type="number"
              min="1"
              max="10"
              class="input w-32"
            />
            <p class="settings-stream-timeout-card__description mt-1.5 text-xs">
              {{ t('admin.settings.streamTimeout.thresholdCountHint') }}
            </p>
          </div>

          <div>
            <label class="settings-stream-timeout-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.streamTimeout.thresholdWindowMinutes') }}
            </label>
            <input
              v-model.number="form.threshold_window_minutes"
              type="number"
              min="1"
              max="60"
              class="input w-32"
            />
            <p class="settings-stream-timeout-card__description mt-1.5 text-xs">
              {{ t('admin.settings.streamTimeout.thresholdWindowMinutesHint') }}
            </p>
          </div>
        </div>

        <div class="settings-stream-timeout-card__footer flex justify-end pt-4">
          <button
            type="button"
            :disabled="saving"
            class="btn btn-primary btn-sm"
            @click="$emit('save')"
          >
            <svg
              v-if="saving"
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
            {{ saving ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { StreamTimeoutSettings } from '@/api/admin/settings'
import Toggle from '@/components/common/Toggle.vue'

defineProps<{
  loading: boolean
  saving: boolean
  form: StreamTimeoutSettings
}>()

defineEmits<{
  save: []
}>()

const { t } = useI18n()
</script>

<style scoped>
.settings-stream-timeout-card__header,
.settings-stream-timeout-card__body,
.settings-stream-timeout-card__section,
.settings-stream-timeout-card__footer {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.settings-stream-timeout-card__header {
  padding:
    var(--theme-settings-card-header-padding-y)
    var(--theme-settings-card-header-padding-x);
  border-top: none;
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.settings-stream-timeout-card__body {
  padding: var(--theme-settings-card-body-padding);
}

.settings-stream-timeout-card__title,
.settings-stream-timeout-card__label,
.settings-stream-timeout-card__field-label {
  color: var(--theme-page-text);
}

.settings-stream-timeout-card__description,
.settings-stream-timeout-card__loading {
  color: var(--theme-page-muted);
}

.settings-stream-timeout-card__spinner {
  border-color: color-mix(in srgb, var(--theme-card-border) 70%, transparent);
  border-bottom-color: var(--theme-accent);
}
</style>
