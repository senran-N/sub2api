<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { opsAPI } from '@/api/admin/ops'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import type { EmailNotificationConfig, AlertSeverity } from '../types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const config = ref<EmailNotificationConfig | null>(null)
let configRequestSequence = 0

const showEditor = ref(false)
const saving = ref(false)
const draft = ref<EmailNotificationConfig | null>(null)
const alertRecipientInput = ref('')
const reportRecipientInput = ref('')
const alertRecipientError = ref('')
const reportRecipientError = ref('')

const severityOptions: Array<{ value: AlertSeverity | ''; label: string }> = [
  { value: '', label: t('admin.ops.email.minSeverityAll') },
  { value: 'critical', label: t('common.critical') },
  { value: 'warning', label: t('common.warning') },
  { value: 'info', label: t('common.info') }
]

async function loadConfig() {
  const requestSequence = ++configRequestSequence
  loading.value = true
  try {
    const data = await opsAPI.getEmailNotificationConfig()
    if (requestSequence !== configRequestSequence) return
    config.value = data
  } catch (err: unknown) {
    if (requestSequence !== configRequestSequence) return
    console.error('[OpsEmailNotificationCard] Failed to load config', err)
    appStore.showError(resolveRequestErrorMessage(err, t('admin.ops.email.loadFailed')))
  } finally {
    if (requestSequence === configRequestSequence) {
      loading.value = false
    }
  }
}

async function saveConfig() {
  if (!draft.value) return
  if (!editorValidation.value.valid) {
    appStore.showError(editorValidation.value.errors[0] || t('admin.ops.email.validation.invalid'))
    return
  }
  const requestSequence = ++configRequestSequence
  loading.value = false
  saving.value = true
  try {
    const savedConfig = await opsAPI.updateEmailNotificationConfig(draft.value)
    if (requestSequence !== configRequestSequence) return
    config.value = savedConfig
    showEditor.value = false
    appStore.showSuccess(t('admin.ops.email.saveSuccess'))
  } catch (err: unknown) {
    if (requestSequence !== configRequestSequence) return
    console.error('[OpsEmailNotificationCard] Failed to save config', err)
    appStore.showError(resolveRequestErrorMessage(err, t('admin.ops.email.saveFailed')))
  } finally {
    if (requestSequence === configRequestSequence) {
      saving.value = false
    }
  }
}

function openEditor() {
  if (!config.value) return
  draft.value = JSON.parse(JSON.stringify(config.value))
  alertRecipientInput.value = ''
  reportRecipientInput.value = ''
  alertRecipientError.value = ''
  reportRecipientError.value = ''
  showEditor.value = true
}

function isValidEmailAddress(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
}

function isNonNegativeNumber(value: unknown): boolean {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0
}

function validateCronField(enabled: boolean, cron: string): string | null {
  if (!enabled) return null
  if (!cron || !cron.trim()) return t('admin.ops.email.validation.cronRequired')
  if (cron.trim().split(/\s+/).length < 5) return t('admin.ops.email.validation.cronFormat')
  return null
}

const editorValidation = computed(() => {
  const errors: string[] = []
  if (!draft.value) return { valid: true, errors }

  if (draft.value.alert.enabled && draft.value.alert.recipients.length === 0) {
    errors.push(t('admin.ops.email.validation.alertRecipientsRequired'))
  }
  if (draft.value.report.enabled && draft.value.report.recipients.length === 0) {
    errors.push(t('admin.ops.email.validation.reportRecipientsRequired'))
  }

  const invalidAlertRecipients = draft.value.alert.recipients.filter((e) => !isValidEmailAddress(e))
  if (invalidAlertRecipients.length > 0) errors.push(t('admin.ops.email.validation.invalidRecipients'))

  const invalidReportRecipients = draft.value.report.recipients.filter((e) => !isValidEmailAddress(e))
  if (invalidReportRecipients.length > 0) errors.push(t('admin.ops.email.validation.invalidRecipients'))

  if (!isNonNegativeNumber(draft.value.alert.rate_limit_per_hour)) {
    errors.push(t('admin.ops.email.validation.rateLimitRange'))
  }
  if (
    !isNonNegativeNumber(draft.value.alert.batching_window_seconds) ||
    draft.value.alert.batching_window_seconds > 86400
  ) {
    errors.push(t('admin.ops.email.validation.batchWindowRange'))
  }

  const dailyErr = validateCronField(
    draft.value.report.daily_summary_enabled,
    draft.value.report.daily_summary_schedule
  )
  if (dailyErr) errors.push(dailyErr)
  const weeklyErr = validateCronField(
    draft.value.report.weekly_summary_enabled,
    draft.value.report.weekly_summary_schedule
  )
  if (weeklyErr) errors.push(weeklyErr)
  const digestErr = validateCronField(
    draft.value.report.error_digest_enabled,
    draft.value.report.error_digest_schedule
  )
  if (digestErr) errors.push(digestErr)
  const accErr = validateCronField(
    draft.value.report.account_health_enabled,
    draft.value.report.account_health_schedule
  )
  if (accErr) errors.push(accErr)

  if (!isNonNegativeNumber(draft.value.report.error_digest_min_count)) {
    errors.push(t('admin.ops.email.validation.digestMinCountRange'))
  }

  const thr = draft.value.report.account_health_error_rate_threshold
  if (!(typeof thr === 'number' && Number.isFinite(thr) && thr >= 0 && thr <= 100)) {
    errors.push(t('admin.ops.email.validation.accountHealthThresholdRange'))
  }

  return { valid: errors.length === 0, errors }
})

function addRecipient(target: 'alert' | 'report') {
  if (!draft.value) return
  const raw = (target === 'alert' ? alertRecipientInput.value : reportRecipientInput.value).trim()
  if (!raw) return

  if (!isValidEmailAddress(raw)) {
    const msg = t('common.invalidEmail')
    if (target === 'alert') alertRecipientError.value = msg
    else reportRecipientError.value = msg
    return
  }

  const normalized = raw.toLowerCase()
  const list = target === 'alert' ? draft.value.alert.recipients : draft.value.report.recipients
  if (!list.includes(normalized)) {
    list.push(normalized)
  }
  if (target === 'alert') alertRecipientInput.value = ''
  else reportRecipientInput.value = ''
  if (target === 'alert') alertRecipientError.value = ''
  else reportRecipientError.value = ''
}

function removeRecipient(target: 'alert' | 'report', email: string) {
  if (!draft.value) return
  const list = target === 'alert' ? draft.value.alert.recipients : draft.value.report.recipients
  const idx = list.indexOf(email)
  if (idx >= 0) list.splice(idx, 1)
}

onMounted(() => {
  loadConfig()
})
</script>

<template>
  <div class="ops-email-notification-card">
    <div class="ops-email-notification-card__header">
      <div class="ops-email-notification-card__header-copy">
        <h3 class="ops-email-notification-card__title ops-email-notification-card__title--section">
          {{ t('admin.ops.email.title') }}
        </h3>
        <p class="ops-email-notification-card__subtitle">{{ t('admin.ops.email.description') }}</p>
      </div>

      <div class="ops-email-notification-card__header-actions">
        <button
          class="ops-email-notification-card__refresh"
          :disabled="loading || saving"
          @click="loadConfig"
        >
          <svg
            class="ops-email-notification-card__refresh-icon"
            :class="{ 'animate-spin': loading }"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          {{ t('common.refresh') }}
        </button>
        <button class="btn btn-sm btn-secondary" :disabled="!config" @click="openEditor">{{ t('common.edit') }}</button>
      </div>
    </div>

    <div v-if="!config" class="ops-email-notification-card__state">
      <span v-if="loading">{{ t('admin.ops.email.loading') }}</span>
      <span v-else>{{ t('admin.ops.email.noData') }}</span>
    </div>

    <div v-else class="ops-email-notification-card__content">
      <div class="ops-email-notification-card__panel">
        <div class="ops-email-notification-card__panel-header">
          <h4 class="ops-email-notification-card__title ops-email-notification-card__title--compact">
            {{ t('admin.ops.email.alertTitle') }}
          </h4>
        </div>

        <div class="ops-email-notification-card__meta-grid">
          <div class="ops-email-notification-card__meta">
            {{ t('common.enabled') }}:
            <span class="ops-email-notification-card__meta-value ops-email-notification-card__meta-value--inline">
              {{ config.alert.enabled ? t('common.enabled') : t('common.disabled') }}
            </span>
          </div>

          <div class="ops-email-notification-card__meta">
            {{ t('admin.ops.email.recipients') }}:
            <span class="ops-email-notification-card__meta-value ops-email-notification-card__meta-value--inline">
              {{ config.alert.recipients.length }}
            </span>
          </div>

          <div class="ops-email-notification-card__meta">
            {{ t('admin.ops.email.minSeverity') }}:
            <span class="ops-email-notification-card__meta-value ops-email-notification-card__meta-value--inline">{{
              config.alert.min_severity || t('admin.ops.email.minSeverityAll')
            }}</span>
          </div>

          <div class="ops-email-notification-card__meta">
            {{ t('admin.ops.email.rateLimitPerHour') }}:
            <span class="ops-email-notification-card__meta-value ops-email-notification-card__meta-value--inline">
              {{ config.alert.rate_limit_per_hour }}
            </span>
          </div>
        </div>
      </div>

      <div class="ops-email-notification-card__panel">
        <div class="ops-email-notification-card__panel-header">
          <h4 class="ops-email-notification-card__title ops-email-notification-card__title--compact">
            {{ t('admin.ops.email.reportTitle') }}
          </h4>
        </div>

        <div class="ops-email-notification-card__meta-grid">
          <div class="ops-email-notification-card__meta">
            {{ t('common.enabled') }}:
            <span class="ops-email-notification-card__meta-value ops-email-notification-card__meta-value--inline">
              {{ config.report.enabled ? t('common.enabled') : t('common.disabled') }}
            </span>
          </div>

          <div class="ops-email-notification-card__meta">
            {{ t('admin.ops.email.recipients') }}:
            <span class="ops-email-notification-card__meta-value ops-email-notification-card__meta-value--inline">
              {{ config.report.recipients.length }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>

  <BaseDialog :show="showEditor" :title="t('admin.ops.email.title')" width="extra-wide" @close="showEditor = false">
    <div v-if="draft" class="ops-email-notification-card__dialog-body">
      <div
        v-if="!editorValidation.valid"
        class="ops-email-notification-card__notice ops-email-notification-card__notice--warning"
      >
        <div class="ops-email-notification-card__title ops-email-notification-card__title--compact">
          {{ t('admin.ops.email.validation.title') }}
        </div>
        <ul class="ops-email-notification-card__validation-list">
          <li v-for="msg in editorValidation.errors" :key="msg">{{ msg }}</li>
        </ul>
      </div>

      <div class="ops-email-notification-card__panel">
        <div class="ops-email-notification-card__panel-header">
          <h4 class="ops-email-notification-card__title ops-email-notification-card__title--compact">
            {{ t('admin.ops.email.alertTitle') }}
          </h4>
        </div>

        <div class="ops-email-notification-card__fields-grid">
          <div class="ops-email-notification-card__field">
            <div class="ops-email-notification-card__field-label">{{ t('common.enabled') }}</div>
            <label class="ops-email-notification-card__toggle">
              <input v-model="draft.alert.enabled" type="checkbox" class="ops-email-notification-card__checkbox" />
              <span>{{ draft.alert.enabled ? t('common.enabled') : t('common.disabled') }}</span>
            </label>
          </div>

          <div class="ops-email-notification-card__field">
            <div class="ops-email-notification-card__field-label">{{ t('admin.ops.email.minSeverity') }}</div>
            <Select v-model="draft.alert.min_severity" :options="severityOptions" />
          </div>

          <div class="ops-email-notification-card__field ops-email-notification-card__field--wide">
            <div class="ops-email-notification-card__field-label">{{ t('admin.ops.email.recipients') }}</div>
            <div class="ops-email-notification-card__recipient-input-row">
              <input
                v-model="alertRecipientInput"
                type="email"
                class="input ops-email-notification-card__recipient-input"
                :placeholder="t('admin.ops.email.recipients')"
                @keydown.enter.prevent="addRecipient('alert')"
              />
              <button class="btn btn-secondary ops-email-notification-card__recipient-action" type="button" @click="addRecipient('alert')">
                {{ t('common.add') }}
              </button>
            </div>
            <p v-if="alertRecipientError" class="ops-email-notification-card__error">{{ alertRecipientError }}</p>
            <div class="ops-email-notification-card__recipient-list">
              <span
                v-for="email in draft.alert.recipients"
                :key="email"
                class="ops-email-notification-card__recipient-chip theme-chip theme-chip--info theme-chip--regular"
              >
                {{ email }}
                <button
                  type="button"
                  class="ops-email-notification-card__chip-remove"
                  @click="removeRecipient('alert', email)"
                >
                  ×
                </button>
              </span>
            </div>
            <div class="ops-email-notification-card__subtitle ops-email-notification-card__subtitle--hint">
              {{ t('admin.ops.email.recipientsHint') }}
            </div>
          </div>

          <div class="ops-email-notification-card__field">
            <div class="ops-email-notification-card__field-label">{{ t('admin.ops.email.rateLimitPerHour') }}</div>
            <input v-model.number="draft.alert.rate_limit_per_hour" type="number" min="0" max="100000" class="input" />
          </div>

          <div class="ops-email-notification-card__field">
            <div class="ops-email-notification-card__field-label">{{ t('admin.ops.email.batchWindowSeconds') }}</div>
            <input v-model.number="draft.alert.batching_window_seconds" type="number" min="0" max="86400" class="input" />
          </div>

          <div class="ops-email-notification-card__field">
            <div class="ops-email-notification-card__field-label">{{ t('admin.ops.email.includeResolved') }}</div>
            <label class="ops-email-notification-card__toggle">
              <input v-model="draft.alert.include_resolved_alerts" type="checkbox" class="ops-email-notification-card__checkbox" />
              <span>{{ draft.alert.include_resolved_alerts ? t('common.enabled') : t('common.disabled') }}</span>
            </label>
          </div>
        </div>
      </div>

      <div class="ops-email-notification-card__panel">
        <div class="ops-email-notification-card__panel-header">
          <h4 class="ops-email-notification-card__title ops-email-notification-card__title--compact">
            {{ t('admin.ops.email.reportTitle') }}
          </h4>
        </div>

        <div class="ops-email-notification-card__fields-grid">
          <div class="ops-email-notification-card__field">
            <div class="ops-email-notification-card__field-label">{{ t('common.enabled') }}</div>
            <label class="ops-email-notification-card__toggle">
              <input v-model="draft.report.enabled" type="checkbox" class="ops-email-notification-card__checkbox" />
              <span>{{ draft.report.enabled ? t('common.enabled') : t('common.disabled') }}</span>
            </label>
          </div>

          <div class="ops-email-notification-card__field ops-email-notification-card__field--wide">
            <div class="ops-email-notification-card__field-label">{{ t('admin.ops.email.recipients') }}</div>
            <div class="ops-email-notification-card__recipient-input-row">
              <input
                v-model="reportRecipientInput"
                type="email"
                class="input ops-email-notification-card__recipient-input"
                :placeholder="t('admin.ops.email.recipients')"
                @keydown.enter.prevent="addRecipient('report')"
              />
              <button class="btn btn-secondary ops-email-notification-card__recipient-action" type="button" @click="addRecipient('report')">
                {{ t('common.add') }}
              </button>
            </div>
            <p v-if="reportRecipientError" class="ops-email-notification-card__error">{{ reportRecipientError }}</p>
            <div class="ops-email-notification-card__recipient-list">
              <span
                v-for="email in draft.report.recipients"
                :key="email"
                class="ops-email-notification-card__recipient-chip theme-chip theme-chip--info theme-chip--regular"
              >
                {{ email }}
                <button
                  type="button"
                  class="ops-email-notification-card__chip-remove"
                  @click="removeRecipient('report', email)"
                >
                  ×
                </button>
              </span>
            </div>
          </div>

          <div class="ops-email-notification-card__field ops-email-notification-card__field--wide">
            <div class="ops-email-notification-card__schedule-grid">
              <div class="ops-email-notification-card__field">
                <div class="ops-email-notification-card__field-label">{{ t('admin.ops.email.dailySummary') }}</div>
                <div class="ops-email-notification-card__schedule-input-row">
                  <label class="ops-email-notification-card__toggle">
                    <input v-model="draft.report.daily_summary_enabled" type="checkbox" class="ops-email-notification-card__checkbox" />
                  </label>
                  <input
                    v-model="draft.report.daily_summary_schedule"
                    type="text"
                    class="input ops-email-notification-card__schedule-input"
                    :placeholder="t('admin.ops.email.cronPlaceholder')"
                  />
                </div>
              </div>

              <div class="ops-email-notification-card__field">
                <div class="ops-email-notification-card__field-label">{{ t('admin.ops.email.weeklySummary') }}</div>
                <div class="ops-email-notification-card__schedule-input-row">
                  <label class="ops-email-notification-card__toggle">
                    <input v-model="draft.report.weekly_summary_enabled" type="checkbox" class="ops-email-notification-card__checkbox" />
                  </label>
                  <input
                    v-model="draft.report.weekly_summary_schedule"
                    type="text"
                    class="input ops-email-notification-card__schedule-input"
                    :placeholder="t('admin.ops.email.cronPlaceholder')"
                  />
                </div>
              </div>

              <div class="ops-email-notification-card__field">
                <div class="ops-email-notification-card__field-label">{{ t('admin.ops.email.errorDigest') }}</div>
                <div class="ops-email-notification-card__schedule-input-row">
                  <label class="ops-email-notification-card__toggle">
                    <input v-model="draft.report.error_digest_enabled" type="checkbox" class="ops-email-notification-card__checkbox" />
                  </label>
                  <input
                    v-model="draft.report.error_digest_schedule"
                    type="text"
                    class="input ops-email-notification-card__schedule-input"
                    :placeholder="t('admin.ops.email.cronPlaceholder')"
                  />
                </div>
              </div>

              <div class="ops-email-notification-card__field">
                <div class="ops-email-notification-card__field-label">{{ t('admin.ops.email.errorDigestMinCount') }}</div>
                <input v-model.number="draft.report.error_digest_min_count" type="number" min="0" max="1000000" class="input" />
              </div>

              <div class="ops-email-notification-card__field">
                <div class="ops-email-notification-card__field-label">{{ t('admin.ops.email.accountHealth') }}</div>
                <div class="ops-email-notification-card__schedule-input-row">
                  <label class="ops-email-notification-card__toggle">
                    <input v-model="draft.report.account_health_enabled" type="checkbox" class="ops-email-notification-card__checkbox" />
                  </label>
                  <input
                    v-model="draft.report.account_health_schedule"
                    type="text"
                    class="input ops-email-notification-card__schedule-input"
                    :placeholder="t('admin.ops.email.cronPlaceholder')"
                  />
                </div>
              </div>

              <div class="ops-email-notification-card__field">
                <div class="ops-email-notification-card__field-label">{{ t('admin.ops.email.accountHealthThreshold') }}</div>
                <input v-model.number="draft.report.account_health_error_rate_threshold" type="number" min="0" max="100" step="0.1" class="input" />
              </div>
            </div>
            <div class="ops-email-notification-card__subtitle ops-email-notification-card__subtitle--hint">
              {{ t('admin.ops.email.reportHint') }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="ops-email-notification-card__footer">
        <button class="btn btn-secondary" @click="showEditor = false">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary" :disabled="saving || !editorValidation.valid" @click="saveConfig">
          {{ saving ? t('common.saving') : t('common.save') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<style scoped>
.ops-email-notification-card {
  padding: var(--theme-ops-card-padding);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  border-radius: var(--theme-surface-radius);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
}

.ops-email-notification-card__title,
.ops-email-notification-card__meta-value {
  color: var(--theme-page-text);
}

.ops-email-notification-card__subtitle,
.ops-email-notification-card__meta,
.ops-email-notification-card__field-label,
.ops-email-notification-card__toggle,
.ops-email-notification-card__state {
  color: var(--theme-page-muted);
}

.ops-email-notification-card__header,
.ops-email-notification-card__panel-header,
.ops-email-notification-card__footer {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--theme-ops-email-header-gap, 1rem);
}

.ops-email-notification-card__header-copy,
.ops-email-notification-card__content,
.ops-email-notification-card__dialog-body,
.ops-email-notification-card__field {
  display: flex;
  flex-direction: column;
}

.ops-email-notification-card__header-copy,
.ops-email-notification-card__field {
  gap: var(--theme-ops-email-field-gap, 0.25rem);
}

.ops-email-notification-card__content,
.ops-email-notification-card__dialog-body {
  gap: var(--theme-ops-email-section-gap, 1.5rem);
}

.ops-email-notification-card__header {
  margin-bottom: var(--theme-ops-email-section-gap, 1.5rem);
}

.ops-email-notification-card__panel-header {
  margin-bottom: var(--theme-ops-email-grid-gap, 1rem);
}

.ops-email-notification-card__header-actions,
.ops-email-notification-card__recipient-input-row,
.ops-email-notification-card__schedule-input-row,
.ops-email-notification-card__toggle {
  display: flex;
  align-items: center;
  gap: var(--theme-ops-email-control-gap, 0.5rem);
}

.ops-email-notification-card__header-actions {
  flex-shrink: 0;
}

.ops-email-notification-card__title--section {
  font-size: 0.875rem;
  font-weight: 700;
}

.ops-email-notification-card__title--compact {
  font-size: 0.75rem;
  font-weight: 700;
}

.ops-email-notification-card__subtitle,
.ops-email-notification-card__meta,
.ops-email-notification-card__field-label,
.ops-email-notification-card__error {
  font-size: 0.75rem;
}

.ops-email-notification-card__subtitle--hint {
  margin-top: var(--theme-ops-email-field-gap, 0.25rem);
}

.ops-email-notification-card__state {
  font-size: 0.875rem;
}

.ops-email-notification-card__refresh {
  display: inline-flex;
  align-items: center;
  gap: calc(var(--theme-ops-email-control-gap, 0.5rem) * 0.75);
  padding: calc(var(--theme-button-padding-y) * 0.6) calc(var(--theme-button-padding-x) * 0.75);
  border-radius: var(--theme-button-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: var(--theme-page-text);
  font-size: 0.75rem;
  font-weight: 700;
  transition: background-color 0.2s ease, color 0.2s ease;
}

.ops-email-notification-card__refresh-icon {
  width: 0.875rem;
  height: 0.875rem;
}

.ops-email-notification-card__refresh:hover {
  background: color-mix(in srgb, var(--theme-page-border) 68%, var(--theme-surface));
}

.ops-email-notification-card__panel {
  padding: var(--theme-ops-panel-padding);
  border-radius: var(--theme-select-panel-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.ops-email-notification-card__meta-grid,
.ops-email-notification-card__fields-grid,
.ops-email-notification-card__schedule-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: var(--theme-ops-email-grid-gap, 1rem);
}

.ops-email-notification-card__meta-value--inline {
  margin-left: 0.25rem;
}

.ops-email-notification-card__notice {
  padding: calc(var(--theme-ops-panel-padding) * 0.75);
  border-radius: var(--theme-button-radius);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
  font-size: 0.75rem;
}

.ops-email-notification-card__notice--warning {
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.ops-email-notification-card__validation-list {
  margin-top: var(--theme-ops-email-field-gap, 0.25rem);
  padding-left: 1rem;
  list-style: disc;
}

.ops-email-notification-card__validation-list li + li {
  margin-top: var(--theme-ops-email-field-gap, 0.25rem);
}

.ops-email-notification-card__field-label {
  font-weight: 500;
}

.ops-email-notification-card__toggle {
  font-size: 0.875rem;
}

.ops-email-notification-card__checkbox {
  width: 1rem;
  height: 1rem;
  border-radius: 0.25rem;
  border-color: color-mix(in srgb, var(--theme-input-border) 82%, transparent);
  color: var(--theme-accent);
}

.ops-email-notification-card__checkbox:focus {
  outline: none;
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--theme-accent) 18%, transparent);
}

.ops-email-notification-card__error {
  color: rgb(var(--theme-danger-rgb));
}

.ops-email-notification-card__recipient-input,
.ops-email-notification-card__schedule-input {
  min-width: 0;
  flex: 1;
}

.ops-email-notification-card__recipient-action {
  flex-shrink: 0;
  white-space: nowrap;
}

.ops-email-notification-card__recipient-list {
  display: flex;
  flex-wrap: wrap;
  gap: var(--theme-ops-email-chip-gap, 0.5rem);
}

.ops-email-notification-card__recipient-chip {
  display: inline-flex;
  align-items: center;
  gap: var(--theme-ops-email-control-gap, 0.5rem);
  padding: calc(var(--theme-button-padding-y) * 0.35) calc(var(--theme-button-padding-x) * 0.75);
  font-size: 0.75rem;
  font-weight: 500;
}

.ops-email-notification-card__chip-remove {
  color: inherit;
  opacity: 0.72;
}

.ops-email-notification-card__chip-remove:hover {
  opacity: 1;
}

.ops-email-notification-card__footer {
  justify-content: flex-end;
  gap: var(--theme-ops-email-footer-gap, 0.5rem);
}

@media (min-width: 768px) {
  .ops-email-notification-card__meta-grid,
  .ops-email-notification-card__fields-grid,
  .ops-email-notification-card__schedule-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .ops-email-notification-card__field--wide {
    grid-column: span 2;
  }
}

@media (max-width: 767px) {
  .ops-email-notification-card__header,
  .ops-email-notification-card__header-actions,
  .ops-email-notification-card__recipient-input-row,
  .ops-email-notification-card__schedule-input-row,
  .ops-email-notification-card__footer {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
