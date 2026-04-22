<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { opsAPI } from '@/api/admin/ops'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import type { OpsAlertRuntimeSettings } from '../types'
import BaseDialog from '@/components/common/BaseDialog.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const saving = ref(false)
let alertSettingsRequestSequence = 0

const alertSettings = ref<OpsAlertRuntimeSettings | null>(null)

const showAlertEditor = ref(false)
const draftAlert = ref<OpsAlertRuntimeSettings | null>(null)

type ValidationResult = { valid: boolean; errors: string[] }
type SilenceEntry = NonNullable<NonNullable<OpsAlertRuntimeSettings['silencing']>['entries']>[number]

function normalizeSeverities(input: Array<string | null | undefined> | null | undefined): string[] {
  if (!input || input.length === 0) return []
  const allowed = new Set(['P0', 'P1', 'P2', 'P3'])
  const out: string[] = []
  const seen = new Set<string>()
  for (const raw of input) {
    const s = String(raw || '')
      .trim()
      .toUpperCase()
    if (!s) continue
    if (!allowed.has(s)) continue
    if (seen.has(s)) continue
    seen.add(s)
    out.push(s)
  }
  return out
}

function getSilenceEntryRuleId(entry: SilenceEntry): number | undefined {
  return typeof entry.rule_id === 'number' ? entry.rule_id : undefined
}

function getSilenceEntrySeverities(entry: SilenceEntry): string[] {
  return Array.isArray(entry.severities) ? normalizeSeverities(entry.severities) : []
}

function validateRuntimeSettings(settings: OpsAlertRuntimeSettings): ValidationResult {
  const errors: string[] = []

  const evalSeconds = settings.evaluation_interval_seconds
  if (!Number.isFinite(evalSeconds) || evalSeconds < 1 || evalSeconds > 86400) {
    errors.push(t('admin.ops.runtime.validation.evalIntervalRange'))
  }

  // Thresholds validation
  const thresholds = settings.thresholds
  if (thresholds) {
    if (thresholds.sla_percent_min != null) {
      if (!Number.isFinite(thresholds.sla_percent_min) || thresholds.sla_percent_min < 0 || thresholds.sla_percent_min > 100) {
        errors.push(t('admin.ops.runtime.validation.slaMinPercentRange'))
      }
    }
    if (thresholds.ttft_p99_ms_max != null) {
      if (!Number.isFinite(thresholds.ttft_p99_ms_max) || thresholds.ttft_p99_ms_max < 0) {
        errors.push(t('admin.ops.runtime.validation.ttftP99MaxRange'))
      }
    }
    if (thresholds.request_error_rate_percent_max != null) {
      if (!Number.isFinite(thresholds.request_error_rate_percent_max) || thresholds.request_error_rate_percent_max < 0 || thresholds.request_error_rate_percent_max > 100) {
        errors.push(t('admin.ops.runtime.validation.requestErrorRateMaxRange'))
      }
    }
    if (thresholds.upstream_error_rate_percent_max != null) {
      if (!Number.isFinite(thresholds.upstream_error_rate_percent_max) || thresholds.upstream_error_rate_percent_max < 0 || thresholds.upstream_error_rate_percent_max > 100) {
        errors.push(t('admin.ops.runtime.validation.upstreamErrorRateMaxRange'))
      }
    }
  }

  const lock = settings.distributed_lock
  if (lock?.enabled) {
    if (!lock.key || lock.key.trim().length < 3) {
      errors.push(t('admin.ops.runtime.validation.lockKeyRequired'))
    } else if (!lock.key.startsWith('ops:')) {
      errors.push(t('admin.ops.runtime.validation.lockKeyPrefix', { prefix: 'ops:' }))
    }
    if (!Number.isFinite(lock.ttl_seconds) || lock.ttl_seconds < 1 || lock.ttl_seconds > 86400) {
      errors.push(t('admin.ops.runtime.validation.lockTtlRange'))
    }
  }

  // Silencing validation (alert-only)
  const silencing = settings.silencing
  if (silencing?.enabled) {
    const until = (silencing.global_until_rfc3339 || '').trim()
    if (until) {
      const parsed = Date.parse(until)
      if (!Number.isFinite(parsed)) errors.push(t('admin.ops.runtime.silencing.validation.timeFormat'))
    }

    const entries = Array.isArray(silencing.entries) ? silencing.entries : []
    for (let idx = 0; idx < entries.length; idx++) {
      const entry = entries[idx]
      const untilEntry = (entry?.until_rfc3339 || '').trim()
      if (!untilEntry) {
        errors.push(t('admin.ops.runtime.silencing.entries.validation.untilRequired'))
        break
      }
      const parsedEntry = Date.parse(untilEntry)
      if (!Number.isFinite(parsedEntry)) {
        errors.push(t('admin.ops.runtime.silencing.entries.validation.untilFormat'))
        break
      }
      const ruleId = getSilenceEntryRuleId(entry)
      if (typeof ruleId === 'number' && (!Number.isFinite(ruleId) || ruleId <= 0)) {
        errors.push(t('admin.ops.runtime.silencing.entries.validation.ruleIdPositive'))
        break
      }
      const raw = entry.severities
      if (raw) {
        const normalized = normalizeSeverities(Array.isArray(raw) ? raw : [raw])
        if (Array.isArray(raw) && raw.length > 0 && normalized.length === 0) {
          errors.push(t('admin.ops.runtime.silencing.entries.validation.severitiesFormat'))
          break
        }
      }
    }
  }

  return { valid: errors.length === 0, errors }
}

const alertValidation = computed(() => {
  if (!draftAlert.value) return { valid: true, errors: [] as string[] }
  return validateRuntimeSettings(draftAlert.value)
})

async function loadSettings() {
  const requestSequence = ++alertSettingsRequestSequence
  loading.value = true
  try {
    const nextSettings = await opsAPI.getAlertRuntimeSettings()
    if (requestSequence !== alertSettingsRequestSequence) return
    alertSettings.value = nextSettings
  } catch (err: unknown) {
    if (requestSequence !== alertSettingsRequestSequence) return
    console.error('[OpsRuntimeSettingsCard] Failed to load runtime settings', err)
    appStore.showError(resolveRequestErrorMessage(err, t('admin.ops.runtime.loadFailed')))
  } finally {
    if (requestSequence === alertSettingsRequestSequence) {
      loading.value = false
    }
  }
}

function openAlertEditor() {
  if (!alertSettings.value) return
  draftAlert.value = JSON.parse(JSON.stringify(alertSettings.value))

  // Backwards-compat: ensure nested settings exist even if API payload is older.
  if (draftAlert.value) {
    if (!draftAlert.value.distributed_lock) {
      draftAlert.value.distributed_lock = { enabled: true, key: 'ops:alert:evaluator:leader', ttl_seconds: 30 }
    }
    if (!draftAlert.value.silencing) {
      draftAlert.value.silencing = { enabled: false, global_until_rfc3339: '', global_reason: '', entries: [] }
    }
    if (!Array.isArray(draftAlert.value.silencing.entries)) {
      draftAlert.value.silencing.entries = []
    }
    if (!draftAlert.value.thresholds) {
      draftAlert.value.thresholds = {
        sla_percent_min: 99.5,
        ttft_p99_ms_max: 500,
        request_error_rate_percent_max: 5,
        upstream_error_rate_percent_max: 5
      }
    }
  }

  showAlertEditor.value = true
}

function addSilenceEntry() {
  if (!draftAlert.value) return
  if (!draftAlert.value.silencing) {
    draftAlert.value.silencing = { enabled: true, global_until_rfc3339: '', global_reason: '', entries: [] }
  }
  if (!Array.isArray(draftAlert.value.silencing.entries)) {
    draftAlert.value.silencing.entries = []
  }
  draftAlert.value.silencing.entries.push({
    rule_id: undefined,
    severities: [],
    until_rfc3339: '',
    reason: ''
  })
}

function removeSilenceEntry(index: number) {
  if (!draftAlert.value?.silencing?.entries) return
  draftAlert.value.silencing.entries.splice(index, 1)
}

function updateSilenceEntryRuleId(index: number, raw: string) {
  const entries = draftAlert.value?.silencing?.entries
  if (!entries || !entries[index]) return
  const trimmed = raw.trim()
  if (!trimmed) {
    delete entries[index].rule_id
    return
  }
  const n = Number.parseInt(trimmed, 10)
  entries[index].rule_id = Number.isFinite(n) ? n : undefined
}

function updateSilenceEntrySeverities(index: number, raw: string) {
  const entries = draftAlert.value?.silencing?.entries
  if (!entries || !entries[index]) return
  const parts = raw
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean)
  entries[index].severities = normalizeSeverities(parts)
}

async function saveAlertSettings() {
  if (!draftAlert.value) return
  if (!alertValidation.value.valid) {
    appStore.showError(alertValidation.value.errors[0] || t('admin.ops.runtime.validation.invalid'))
    return
  }

  const requestSequence = ++alertSettingsRequestSequence
  loading.value = false
  saving.value = true
  try {
    const savedSettings = await opsAPI.updateAlertRuntimeSettings(draftAlert.value)
    if (requestSequence !== alertSettingsRequestSequence) return
    alertSettings.value = savedSettings
    showAlertEditor.value = false
    appStore.showSuccess(t('admin.ops.runtime.saveSuccess'))
  } catch (err: unknown) {
    if (requestSequence !== alertSettingsRequestSequence) return
    console.error('[OpsRuntimeSettingsCard] Failed to save alert runtime settings', err)
    appStore.showError(resolveRequestErrorMessage(err, t('admin.ops.runtime.saveFailed')))
  } finally {
    if (requestSequence === alertSettingsRequestSequence) {
      saving.value = false
    }
  }
}

onMounted(() => {
  loadSettings()
})
</script>

<template>
  <div class="ops-runtime-settings-card">
    <div class="ops-runtime-settings-card__header">
      <div class="ops-runtime-settings-card__header-copy">
        <h3 class="ops-runtime-settings-card__title ops-runtime-settings-card__title--section">{{ t('admin.ops.runtime.title') }}</h3>
        <p class="ops-runtime-settings-card__subtitle ops-runtime-settings-card__subtitle--hint">{{ t('admin.ops.runtime.description') }}</p>
      </div>
      <button
        class="ops-runtime-settings-card__refresh"
        :disabled="loading || saving"
        @click="loadSettings"
      >
        <svg class="ops-runtime-settings-card__refresh-icon" :class="{ 'animate-spin': loading }" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
        {{ t('common.refresh') }}
      </button>
    </div>

    <div v-if="!alertSettings" class="ops-runtime-settings-card__state">
      <span v-if="loading">{{ t('admin.ops.runtime.loading') }}</span>
      <span v-else>{{ t('admin.ops.runtime.noData') }}</span>
    </div>

    <div v-else class="ops-runtime-settings-card__content">
      <div class="ops-runtime-settings-card__panel">
        <div class="ops-runtime-settings-card__panel-header">
          <h4 class="ops-runtime-settings-card__title ops-runtime-settings-card__title--section">{{ t('admin.ops.runtime.alertTitle') }}</h4>
          <button class="btn btn-sm btn-secondary" @click="openAlertEditor">{{ t('common.edit') }}</button>
        </div>
        <div class="ops-runtime-settings-card__meta-grid">
          <div class="ops-runtime-settings-card__meta">
            {{ t('admin.ops.runtime.evalIntervalSeconds') }}:
            <span class="ops-runtime-settings-card__meta-value ops-runtime-settings-card__meta-value--inline">{{ alertSettings.evaluation_interval_seconds }}s</span>
          </div>
          <div
            v-if="alertSettings.silencing?.enabled && alertSettings.silencing.global_until_rfc3339"
            class="ops-runtime-settings-card__meta ops-runtime-settings-card__meta--wide"
          >
            {{ t('admin.ops.runtime.silencing.globalUntil') }}:
            <span class="ops-runtime-settings-card__meta-value ops-runtime-settings-card__meta-value--inline ops-runtime-settings-card__meta-value--mono">
              {{ alertSettings.silencing.global_until_rfc3339 }}
            </span>
          </div>

          <details class="ops-runtime-settings-card__details ops-runtime-settings-card__details--wide">
            <summary class="ops-runtime-settings-card__summary">
              {{ t('admin.ops.runtime.showAdvancedDeveloperSettings') }}
            </summary>
            <div class="ops-runtime-settings-card__advanced-grid">
              <div class="ops-runtime-settings-card__subtitle">
                {{ t('admin.ops.runtime.lockEnabled') }}:
                <span class="ops-runtime-settings-card__advanced-value ops-runtime-settings-card__advanced-value--inline ops-runtime-settings-card__advanced-value--mono">
                  {{ alertSettings.distributed_lock.enabled }}
                </span>
              </div>
              <div class="ops-runtime-settings-card__subtitle">
                {{ t('admin.ops.runtime.lockKey') }}:
                <span class="ops-runtime-settings-card__advanced-value ops-runtime-settings-card__advanced-value--inline ops-runtime-settings-card__advanced-value--mono">
                  {{ alertSettings.distributed_lock.key }}
                </span>
              </div>
              <div class="ops-runtime-settings-card__subtitle">
                {{ t('admin.ops.runtime.lockTTLSeconds') }}:
                <span class="ops-runtime-settings-card__advanced-value ops-runtime-settings-card__advanced-value--inline ops-runtime-settings-card__advanced-value--mono">
                  {{ alertSettings.distributed_lock.ttl_seconds }}s
                </span>
              </div>
            </div>
          </details>
        </div>
      </div>
    </div>
  </div>

  <BaseDialog :show="showAlertEditor" :title="t('admin.ops.runtime.alertTitle')" width="extra-wide" @close="showAlertEditor = false">
    <div v-if="draftAlert" class="ops-runtime-settings-card__dialog-body">
      <div
        v-if="!alertValidation.valid"
        class="ops-runtime-settings-card__notice ops-runtime-settings-card__notice--warning"
      >
        <div class="ops-runtime-settings-card__title ops-runtime-settings-card__title--compact">{{ t('admin.ops.runtime.validation.title') }}</div>
        <ul class="ops-runtime-settings-card__validation-list">
          <li v-for="msg in alertValidation.errors" :key="msg">{{ msg }}</li>
        </ul>
      </div>

      <div class="ops-runtime-settings-card__field">
        <div class="ops-runtime-settings-card__field-label">{{ t('admin.ops.runtime.evalIntervalSeconds') }}</div>
        <input
          v-model.number="draftAlert.evaluation_interval_seconds"
          type="number"
          min="1"
          max="86400"
          class="input"
          :aria-invalid="!alertValidation.valid"
        />
        <p class="ops-runtime-settings-card__subtitle ops-runtime-settings-card__subtitle--hint">{{ t('admin.ops.runtime.evalIntervalHint') }}</p>
      </div>

      <div class="ops-runtime-settings-card__panel">
        <div class="ops-runtime-settings-card__title ops-runtime-settings-card__title--section">{{ t('admin.ops.runtime.metricThresholds') }}</div>
        <p class="ops-runtime-settings-card__subtitle ops-runtime-settings-card__subtitle--section">{{ t('admin.ops.runtime.metricThresholdsHint') }}</p>

        <div class="ops-runtime-settings-card__fields-grid">
          <div class="ops-runtime-settings-card__field">
            <div class="ops-runtime-settings-card__field-label">{{ t('admin.ops.runtime.slaMinPercent') }}</div>
            <input
              v-model.number="draftAlert.thresholds.sla_percent_min"
              type="number"
              min="0"
              max="100"
              step="0.1"
              class="input"
              placeholder="99.5"
            />
            <p class="ops-runtime-settings-card__subtitle ops-runtime-settings-card__subtitle--hint">{{ t('admin.ops.runtime.slaMinPercentHint') }}</p>
          </div>

          <div class="ops-runtime-settings-card__field">
            <div class="ops-runtime-settings-card__field-label">{{ t('admin.ops.runtime.ttftP99MaxMs') }}</div>
            <input
              v-model.number="draftAlert.thresholds.ttft_p99_ms_max"
              type="number"
              min="0"
              step="100"
              class="input"
              placeholder="500"
            />
            <p class="ops-runtime-settings-card__subtitle ops-runtime-settings-card__subtitle--hint">{{ t('admin.ops.runtime.ttftP99MaxMsHint') }}</p>
          </div>

          <div class="ops-runtime-settings-card__field">
            <div class="ops-runtime-settings-card__field-label">{{ t('admin.ops.runtime.requestErrorRateMaxPercent') }}</div>
            <input
              v-model.number="draftAlert.thresholds.request_error_rate_percent_max"
              type="number"
              min="0"
              max="100"
              step="0.1"
              class="input"
              placeholder="5"
            />
            <p class="ops-runtime-settings-card__subtitle ops-runtime-settings-card__subtitle--hint">
              {{ t('admin.ops.runtime.requestErrorRateMaxPercentHint') }}
            </p>
          </div>

          <div class="ops-runtime-settings-card__field">
            <div class="ops-runtime-settings-card__field-label">{{ t('admin.ops.runtime.upstreamErrorRateMaxPercent') }}</div>
            <input
              v-model.number="draftAlert.thresholds.upstream_error_rate_percent_max"
              type="number"
              min="0"
              max="100"
              step="0.1"
              class="input"
              placeholder="5"
            />
            <p class="ops-runtime-settings-card__subtitle ops-runtime-settings-card__subtitle--hint">
              {{ t('admin.ops.runtime.upstreamErrorRateMaxPercentHint') }}
            </p>
          </div>
        </div>
      </div>

      <div class="ops-runtime-settings-card__panel">
        <div class="ops-runtime-settings-card__title ops-runtime-settings-card__title--section">{{ t('admin.ops.runtime.silencing.title') }}</div>

        <label class="ops-runtime-settings-card__toggle">
          <input v-model="draftAlert.silencing.enabled" type="checkbox" class="ops-runtime-settings-card__checkbox" />
          <span>{{ t('admin.ops.runtime.silencing.enabled') }}</span>
        </label>

        <div v-if="draftAlert.silencing.enabled" class="ops-runtime-settings-card__silencing-body">
          <div class="ops-runtime-settings-card__field">
            <div class="ops-runtime-settings-card__field-label">{{ t('admin.ops.runtime.silencing.globalUntil') }}</div>
            <input
              v-model="draftAlert.silencing.global_until_rfc3339"
              type="text"
              class="input ops-runtime-settings-card__input--mono"
              placeholder="2026-01-05T00:00:00Z"
            />
            <p class="ops-runtime-settings-card__subtitle ops-runtime-settings-card__subtitle--hint">{{ t('admin.ops.runtime.silencing.untilHint') }}</p>
          </div>

          <div class="ops-runtime-settings-card__field">
            <div class="ops-runtime-settings-card__field-label">{{ t('admin.ops.runtime.silencing.reason') }}</div>
            <input
              v-model="draftAlert.silencing.global_reason"
              type="text"
              class="input"
              :placeholder="t('admin.ops.runtime.silencing.reasonPlaceholder')"
            />
          </div>

          <div class="ops-runtime-settings-card__subpanel">
            <div class="ops-runtime-settings-card__subpanel-header">
              <div class="ops-runtime-settings-card__header-copy">
                <div class="ops-runtime-settings-card__title ops-runtime-settings-card__title--compact">
                  {{ t('admin.ops.runtime.silencing.entries.title') }}
                </div>
                <p class="ops-runtime-settings-card__subtitle ops-runtime-settings-card__subtitle--micro">
                  {{ t('admin.ops.runtime.silencing.entries.hint') }}
                </p>
              </div>
              <button class="btn btn-sm btn-secondary" type="button" @click="addSilenceEntry">
                {{ t('admin.ops.runtime.silencing.entries.add') }}
              </button>
            </div>

            <div v-if="!draftAlert.silencing.entries?.length" class="ops-runtime-settings-card__empty">
              {{ t('admin.ops.runtime.silencing.entries.empty') }}
            </div>

            <div v-else class="ops-runtime-settings-card__entries">
              <div
                v-for="(entry, idx) in draftAlert.silencing.entries"
                :key="idx"
                class="ops-runtime-settings-card__entry"
              >
                <div class="ops-runtime-settings-card__entry-header">
                  <div class="ops-runtime-settings-card__title ops-runtime-settings-card__title--compact">
                    {{ t('admin.ops.runtime.silencing.entries.entryTitle', { n: idx + 1 }) }}
                  </div>
                  <button class="btn btn-sm btn-danger" type="button" @click="removeSilenceEntry(idx)">{{ t('common.delete') }}</button>
                </div>

                <div class="ops-runtime-settings-card__fields-grid">
                  <div class="ops-runtime-settings-card__field">
                    <div class="ops-runtime-settings-card__field-label">{{ t('admin.ops.runtime.silencing.entries.ruleId') }}</div>
                    <input
                      :value="typeof getSilenceEntryRuleId(entry) === 'number' ? String(getSilenceEntryRuleId(entry)) : ''"
                      type="text"
                      class="input ops-runtime-settings-card__input--mono"
                      :placeholder="t('admin.ops.runtime.silencing.entries.ruleIdPlaceholder')"
                      @input="updateSilenceEntryRuleId(idx, ($event.target as HTMLInputElement).value)"
                    />
                  </div>

                  <div class="ops-runtime-settings-card__field">
                    <div class="ops-runtime-settings-card__field-label">{{ t('admin.ops.runtime.silencing.entries.severities') }}</div>
                    <input
                      :value="getSilenceEntrySeverities(entry).join(', ')"
                      type="text"
                      class="input ops-runtime-settings-card__input--mono"
                      :placeholder="t('admin.ops.runtime.silencing.entries.severitiesPlaceholder')"
                      @input="updateSilenceEntrySeverities(idx, ($event.target as HTMLInputElement).value)"
                    />
                  </div>

                  <div class="ops-runtime-settings-card__field">
                    <div class="ops-runtime-settings-card__field-label">{{ t('admin.ops.runtime.silencing.entries.until') }}</div>
                    <input
                      v-model="entry.until_rfc3339"
                      type="text"
                      class="input ops-runtime-settings-card__input--mono"
                      placeholder="2026-01-05T00:00:00Z"
                    />
                  </div>

                  <div class="ops-runtime-settings-card__field">
                    <div class="ops-runtime-settings-card__field-label">{{ t('admin.ops.runtime.silencing.entries.reason') }}</div>
                    <input
                      v-model="entry.reason"
                      type="text"
                      class="input"
                      :placeholder="t('admin.ops.runtime.silencing.reasonPlaceholder')"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <details class="ops-runtime-settings-card__advanced">
        <summary class="ops-runtime-settings-card__summary">{{ t('admin.ops.runtime.advancedSettingsSummary') }}</summary>
        <div class="ops-runtime-settings-card__fields-grid ops-runtime-settings-card__fields-grid--advanced">
          <div class="ops-runtime-settings-card__field">
            <label class="ops-runtime-settings-card__toggle">
              <input v-model="draftAlert.distributed_lock.enabled" type="checkbox" class="ops-runtime-settings-card__checkbox" />
              <span>{{ t('admin.ops.runtime.lockEnabled') }}</span>
            </label>
          </div>
          <div class="ops-runtime-settings-card__field ops-runtime-settings-card__field--wide">
            <div class="ops-runtime-settings-card__field-label">{{ t('admin.ops.runtime.lockKey') }}</div>
            <input v-model="draftAlert.distributed_lock.key" type="text" class="input ops-runtime-settings-card__input--mono ops-runtime-settings-card__input--compact" />
            <p v-if="draftAlert.distributed_lock.enabled" class="ops-runtime-settings-card__subtitle ops-runtime-settings-card__subtitle--micro">
              {{ t('admin.ops.runtime.validation.lockKeyHint', { prefix: 'ops:' }) }}
            </p>
          </div>
          <div class="ops-runtime-settings-card__field">
            <div class="ops-runtime-settings-card__field-label">{{ t('admin.ops.runtime.lockTTLSeconds') }}</div>
            <input
              v-model.number="draftAlert.distributed_lock.ttl_seconds"
              type="number"
              min="1"
              max="86400"
              class="input ops-runtime-settings-card__input--mono ops-runtime-settings-card__input--compact"
            />
          </div>
        </div>
      </details>
    </div>

    <template #footer>
      <div class="ops-runtime-settings-card__footer">
        <button class="btn btn-secondary" @click="showAlertEditor = false">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary" :disabled="saving || !alertValidation.valid" @click="saveAlertSettings">
          {{ saving ? t('common.saving') : t('common.save') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<style scoped>
.ops-runtime-settings-card {
  padding: var(--theme-ops-card-padding);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  border-radius: var(--theme-surface-radius);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
}

.ops-runtime-settings-card__title,
.ops-runtime-settings-card__meta-value,
.ops-runtime-settings-card__advanced-value {
  color: var(--theme-page-text);
}

.ops-runtime-settings-card__subtitle,
.ops-runtime-settings-card__meta,
.ops-runtime-settings-card__field-label,
.ops-runtime-settings-card__toggle,
.ops-runtime-settings-card__state {
  color: var(--theme-page-muted);
}

.ops-runtime-settings-card__header,
.ops-runtime-settings-card__panel-header,
.ops-runtime-settings-card__subpanel-header,
.ops-runtime-settings-card__entry-header,
.ops-runtime-settings-card__footer {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--theme-ops-runtime-header-gap);
}

.ops-runtime-settings-card__header-copy,
.ops-runtime-settings-card__field,
.ops-runtime-settings-card__dialog-body,
.ops-runtime-settings-card__content,
.ops-runtime-settings-card__silencing-body,
.ops-runtime-settings-card__entries {
  display: flex;
  flex-direction: column;
}

.ops-runtime-settings-card__header-copy,
.ops-runtime-settings-card__field {
  gap: var(--theme-ops-runtime-field-gap);
}

.ops-runtime-settings-card__dialog-body,
.ops-runtime-settings-card__silencing-body {
  gap: var(--theme-ops-runtime-panel-gap);
}

.ops-runtime-settings-card__content {
  gap: var(--theme-ops-runtime-section-gap);
}

.ops-runtime-settings-card__entries {
  margin-top: var(--theme-ops-runtime-panel-gap);
  gap: var(--theme-ops-runtime-entry-gap);
}

.ops-runtime-settings-card__header {
  margin-bottom: var(--theme-ops-runtime-panel-gap);
}

.ops-runtime-settings-card__panel-header,
.ops-runtime-settings-card__subpanel-header,
.ops-runtime-settings-card__entry-header {
  margin-bottom: var(--theme-ops-runtime-panel-gap);
}

.ops-runtime-settings-card__title--section {
  font-size: 0.875rem;
  font-weight: 700;
}

.ops-runtime-settings-card__title--compact {
  font-size: 0.75rem;
  font-weight: 700;
}

.ops-runtime-settings-card__subtitle {
  font-size: 0.75rem;
}

.ops-runtime-settings-card__subtitle--hint {
  margin-top: var(--theme-ops-runtime-field-gap);
}

.ops-runtime-settings-card__subtitle--section {
  margin-top: calc(var(--theme-ops-runtime-field-gap) * 2);
  margin-bottom: var(--theme-ops-runtime-panel-gap);
}

.ops-runtime-settings-card__subtitle--micro {
  font-size: 0.6875rem;
}

.ops-runtime-settings-card__state {
  font-size: 0.875rem;
}

.ops-runtime-settings-card__refresh {
  display: inline-flex;
  align-items: center;
  gap: calc(var(--theme-ops-runtime-toggle-gap) * 0.75);
  padding: calc(var(--theme-button-padding-y) * 0.6) calc(var(--theme-button-padding-x) * 0.75);
  border-radius: var(--theme-button-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: var(--theme-page-text);
  font-size: 0.75rem;
  font-weight: 700;
  transition: background-color 0.2s ease, color 0.2s ease;
}

.ops-runtime-settings-card__refresh-icon {
  width: 0.875rem;
  height: 0.875rem;
}

.ops-runtime-settings-card__refresh {
  align-self: flex-start;
}

.ops-runtime-settings-card__refresh:hover {
  background: color-mix(in srgb, var(--theme-page-border) 68%, var(--theme-surface));
}

.ops-runtime-settings-card__panel,
.ops-runtime-settings-card__entry,
.ops-runtime-settings-card__advanced {
  padding: var(--theme-ops-panel-padding);
  border-radius: var(--theme-select-panel-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.ops-runtime-settings-card__meta-grid,
.ops-runtime-settings-card__fields-grid,
.ops-runtime-settings-card__advanced-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: var(--theme-ops-runtime-grid-gap);
}

.ops-runtime-settings-card__advanced-grid {
  margin-top: calc(var(--theme-ops-runtime-field-gap) * 2);
}

.ops-runtime-settings-card__fields-grid--advanced {
  margin-top: var(--theme-ops-runtime-panel-gap);
}

.ops-runtime-settings-card__meta,
.ops-runtime-settings-card__field-label {
  font-size: 0.75rem;
}

.ops-runtime-settings-card__field-label {
  font-weight: 500;
}

.ops-runtime-settings-card__meta-value--inline,
.ops-runtime-settings-card__advanced-value--inline {
  margin-left: 0.25rem;
}

.ops-runtime-settings-card__meta-value--mono,
.ops-runtime-settings-card__advanced-value--mono,
.ops-runtime-settings-card__input--mono {
  font-family: var(--theme-font-mono, ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, Liberation Mono, Courier New, monospace);
}

.ops-runtime-settings-card__input--mono {
  font-size: 0.875rem;
}

.ops-runtime-settings-card__input--compact {
  font-size: 0.75rem;
}

.ops-runtime-settings-card__entry,
.ops-runtime-settings-card__advanced {
  border: 1px solid color-mix(in srgb, var(--theme-page-border) 74%, transparent);
}

.ops-runtime-settings-card__subpanel {
  padding: var(--theme-ops-panel-padding);
  border-radius: var(--theme-select-panel-radius);
  border: 1px solid color-mix(in srgb, var(--theme-page-border) 74%, transparent);
  background: var(--theme-surface);
}

.ops-runtime-settings-card__advanced-grid,
.ops-runtime-settings-card__empty {
  padding: calc(var(--theme-ops-panel-padding) * 0.75);
  border-radius: var(--theme-button-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 92%, var(--theme-surface));
}

.ops-runtime-settings-card__summary {
  display: inline-flex;
  align-items: center;
  cursor: pointer;
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
  font-size: 0.75rem;
  font-weight: 500;
}

.ops-runtime-settings-card__summary:hover {
  color: rgb(var(--theme-info-rgb));
}

.ops-runtime-settings-card__notice {
  padding: calc(var(--theme-ops-panel-padding) * 0.75);
  border-radius: var(--theme-button-radius);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
  border-color: color-mix(in srgb, var(--theme-card-border) 68%, transparent);
  font-size: 0.75rem;
}

.ops-runtime-settings-card__notice--warning {
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.ops-runtime-settings-card__validation-list {
  margin-top: var(--theme-ops-runtime-field-gap);
  padding-left: 1rem;
  list-style: disc;
}

.ops-runtime-settings-card__validation-list li + li {
  margin-top: var(--theme-ops-runtime-field-gap);
}

.ops-runtime-settings-card__toggle {
  display: inline-flex;
  align-items: center;
  gap: var(--theme-ops-runtime-toggle-gap);
  font-size: 0.875rem;
}

.ops-runtime-settings-card__checkbox {
  width: 1rem;
  height: 1rem;
  border-radius: 0.25rem;
  border-color: color-mix(in srgb, var(--theme-input-border) 82%, transparent);
  color: var(--theme-accent);
}

.ops-runtime-settings-card__checkbox:focus {
  outline: none;
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--theme-accent) 18%, transparent);
}

.ops-runtime-settings-card__advanced > summary {
  padding: 0;
}

.ops-runtime-settings-card__footer {
  justify-content: flex-end;
  gap: var(--theme-ops-runtime-footer-gap);
}

@media (min-width: 768px) {
  .ops-runtime-settings-card__meta-grid,
  .ops-runtime-settings-card__fields-grid,
  .ops-runtime-settings-card__advanced-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .ops-runtime-settings-card__meta--wide,
  .ops-runtime-settings-card__details--wide,
  .ops-runtime-settings-card__field--wide {
    grid-column: span 2;
  }
}
</style>
