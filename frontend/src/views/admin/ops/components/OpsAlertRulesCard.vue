<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import { adminAPI } from '@/api'
import { opsAPI } from '@/api/admin/ops'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import type { AlertRule, MetricType, Operator } from '../types'
import type { OpsSeverity } from '@/api/admin/ops'
import { formatDateTime } from '../utils/opsFormatters'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const rules = ref<AlertRule[]>([])
let rulesRequestSequence = 0

async function load(requestSequence = ++rulesRequestSequence) {
  loading.value = true
  try {
    const nextRules = await opsAPI.listAlertRules()
    if (requestSequence !== rulesRequestSequence) return
    rules.value = nextRules
  } catch (err: unknown) {
    if (requestSequence !== rulesRequestSequence) return
    console.error('[OpsAlertRulesCard] Failed to load rules', err)
    appStore.showError(resolveRequestErrorMessage(err, t('admin.ops.alertRules.loadFailed')))
    rules.value = []
  } finally {
    if (requestSequence === rulesRequestSequence) {
      loading.value = false
    }
  }
}

onMounted(() => {
  load()
  loadGroups()
})

const sortedRules = computed(() => {
  return [...rules.value].sort((a, b) => (b.id || 0) - (a.id || 0))
})

const showEditor = ref(false)
const saving = ref(false)
const creatingPresets = ref(false)
const deleting = ref(false)
const editingId = ref<number | null>(null)
const draft = ref<AlertRule | null>(null)

type MetricGroup = 'system' | 'group' | 'account'

interface MetricDefinition {
  type: MetricType
  group: MetricGroup
  label: string
  description: string
  recommendedOperator: Operator
  recommendedThreshold: number
  unit?: string
}

interface AlertRulePreset {
  key: string
  name: string
  description: string
  draft: AlertRule
}

interface PresetCreationResult {
  succeeded: AlertRulePreset[]
  failed: Array<{
    preset: AlertRulePreset
    detail: string
  }>
}

const groupMetricTypes = new Set<MetricType>([
  'group_available_accounts',
  'group_available_ratio',
  'group_rate_limit_ratio'
])

function parsePositiveInt(value: unknown): number | null {
  if (value == null) return null
  if (typeof value === 'boolean') return null
  const n = typeof value === 'number' ? value : Number.parseInt(String(value), 10)
  return Number.isFinite(n) && n > 0 ? n : null
}

function normalizeRuleScopeValue(value: unknown): string {
  if (typeof value !== 'string') return ''
  return value.trim().toLowerCase()
}

function normalizeRuleThreshold(value: unknown): string {
  if (!(typeof value === 'number' && Number.isFinite(value))) return ''
  return String(value)
}

function buildRuleSemanticKey(rule: Pick<AlertRule, 'metric_type' | 'operator' | 'threshold' | 'window_minutes' | 'sustained_minutes' | 'filters'>): string {
  return [
    rule.metric_type,
    rule.operator,
    normalizeRuleThreshold(rule.threshold),
    String(rule.window_minutes ?? ''),
    String(rule.sustained_minutes ?? ''),
    normalizeRuleScopeValue(rule.filters?.platform),
    String(parsePositiveInt(rule.filters?.group_id) ?? '')
  ].join('|')
}

const groupOptionsBase = ref<SelectOption[]>([])

async function loadGroups() {
  try {
    const list = await adminAPI.groups.getAll()
    groupOptionsBase.value = list.map((g) => ({ value: g.id, label: g.name }))
  } catch (err) {
    console.error('[OpsAlertRulesCard] Failed to load groups', err)
    groupOptionsBase.value = []
  }
}

const isGroupMetricSelected = computed(() => {
  const metricType = draft.value?.metric_type
  return metricType ? groupMetricTypes.has(metricType) : false
})

const draftGroupId = computed<number | null>({
  get() {
    return parsePositiveInt(draft.value?.filters?.group_id)
  },
  set(value) {
    if (!draft.value) return
    if (value == null) {
      if (!draft.value.filters) return
      delete draft.value.filters.group_id
      if (Object.keys(draft.value.filters).length === 0) {
        delete draft.value.filters
      }
      return
    }
    if (!draft.value.filters) draft.value.filters = {}
    draft.value.filters.group_id = value
  }
})

const groupOptions = computed<SelectOption[]>(() => {
  if (isGroupMetricSelected.value) return groupOptionsBase.value
  return [{ value: null, label: t('admin.ops.alertRules.form.allGroups') }, ...groupOptionsBase.value]
})

const metricDefinitions = computed(() => {
  return [
    // System-level metrics
    {
      type: 'success_rate',
      group: 'system',
      label: t('admin.ops.alertRules.metrics.successRate'),
      description: t('admin.ops.alertRules.metricDescriptions.successRate'),
      recommendedOperator: '<',
      recommendedThreshold: 99,
      unit: '%'
    },
    {
      type: 'error_rate',
      group: 'system',
      label: t('admin.ops.alertRules.metrics.errorRate'),
      description: t('admin.ops.alertRules.metricDescriptions.errorRate'),
      recommendedOperator: '>',
      recommendedThreshold: 1,
      unit: '%'
    },
    {
      type: 'upstream_error_rate',
      group: 'system',
      label: t('admin.ops.alertRules.metrics.upstreamErrorRate'),
      description: t('admin.ops.alertRules.metricDescriptions.upstreamErrorRate'),
      recommendedOperator: '>',
      recommendedThreshold: 1,
      unit: '%'
    },
    {
      type: 'cpu_usage_percent',
      group: 'system',
      label: t('admin.ops.alertRules.metrics.cpu'),
      description: t('admin.ops.alertRules.metricDescriptions.cpu'),
      recommendedOperator: '>',
      recommendedThreshold: 80,
      unit: '%'
    },
    {
      type: 'memory_usage_percent',
      group: 'system',
      label: t('admin.ops.alertRules.metrics.memory'),
      description: t('admin.ops.alertRules.metricDescriptions.memory'),
      recommendedOperator: '>',
      recommendedThreshold: 80,
      unit: '%'
    },
    {
      type: 'concurrency_queue_depth',
      group: 'system',
      label: t('admin.ops.alertRules.metrics.queueDepth'),
      description: t('admin.ops.alertRules.metricDescriptions.queueDepth'),
      recommendedOperator: '>',
      recommendedThreshold: 10
    },
    {
      type: 'scheduler_acquire_success_rate',
      group: 'system',
      label: t('admin.ops.alertRules.metrics.schedulerAcquireSuccessRate'),
      description: t('admin.ops.alertRules.metricDescriptions.schedulerAcquireSuccessRate'),
      recommendedOperator: '<',
      recommendedThreshold: 75,
      unit: '%'
    },
    {
      type: 'scheduler_wait_plan_success_rate',
      group: 'system',
      label: t('admin.ops.alertRules.metrics.schedulerWaitPlanSuccessRate'),
      description: t('admin.ops.alertRules.metricDescriptions.schedulerWaitPlanSuccessRate'),
      recommendedOperator: '<',
      recommendedThreshold: 60,
      unit: '%'
    },
    {
      type: 'scheduler_index_page_density',
      group: 'system',
      label: t('admin.ops.alertRules.metrics.schedulerIndexPageDensity'),
      description: t('admin.ops.alertRules.metricDescriptions.schedulerIndexPageDensity'),
      recommendedOperator: '<',
      recommendedThreshold: 8
    },
    {
      type: 'idempotency_processing_avg_ms',
      group: 'system',
      label: t('admin.ops.alertRules.metrics.idempotencyProcessingAvgMs'),
      description: t('admin.ops.alertRules.metricDescriptions.idempotencyProcessingAvgMs'),
      recommendedOperator: '>',
      recommendedThreshold: 80,
      unit: 'ms'
    },

    // Group-level metrics (requires group_id filter)
    {
      type: 'group_available_accounts',
      group: 'group',
      label: t('admin.ops.alertRules.metrics.groupAvailableAccounts'),
      description: t('admin.ops.alertRules.metricDescriptions.groupAvailableAccounts'),
      recommendedOperator: '<',
      recommendedThreshold: 1
    },
    {
      type: 'group_available_ratio',
      group: 'group',
      label: t('admin.ops.alertRules.metrics.groupAvailableRatio'),
      description: t('admin.ops.alertRules.metricDescriptions.groupAvailableRatio'),
      recommendedOperator: '<',
      recommendedThreshold: 50,
      unit: '%'
    },
    {
      type: 'group_rate_limit_ratio',
      group: 'group',
      label: t('admin.ops.alertRules.metrics.groupRateLimitRatio'),
      description: t('admin.ops.alertRules.metricDescriptions.groupRateLimitRatio'),
      recommendedOperator: '>',
      recommendedThreshold: 10,
      unit: '%'
    },

    // Account-level metrics
    {
      type: 'account_rate_limited_count',
      group: 'account',
      label: t('admin.ops.alertRules.metrics.accountRateLimitedCount'),
      description: t('admin.ops.alertRules.metricDescriptions.accountRateLimitedCount'),
      recommendedOperator: '>',
      recommendedThreshold: 0
    },
    {
      type: 'account_error_count',
      group: 'account',
      label: t('admin.ops.alertRules.metrics.accountErrorCount'),
      description: t('admin.ops.alertRules.metricDescriptions.accountErrorCount'),
      recommendedOperator: '>',
      recommendedThreshold: 0
    },
    {
      type: 'account_error_ratio',
      group: 'account',
      label: t('admin.ops.alertRules.metrics.accountErrorRatio'),
      description: t('admin.ops.alertRules.metricDescriptions.accountErrorRatio'),
      recommendedOperator: '>',
      recommendedThreshold: 5,
      unit: '%'
    },
    {
      type: 'overload_account_count',
      group: 'account',
      label: t('admin.ops.alertRules.metrics.overloadAccountCount'),
      description: t('admin.ops.alertRules.metricDescriptions.overloadAccountCount'),
      recommendedOperator: '>',
      recommendedThreshold: 0
    }
  ] satisfies MetricDefinition[]
})

const selectedMetricDefinition = computed(() => {
  const metricType = draft.value?.metric_type
  if (!metricType) return null
  return metricDefinitions.value.find((m) => m.type === metricType) ?? null
})

const metricOptions = computed(() => {
  const buildGroup = (group: MetricGroup): SelectOption[] => {
    const items = metricDefinitions.value.filter((m) => m.group === group)
    if (items.length === 0) return []
    const headerValue = `__group__${group}`
    return [
      {
        value: headerValue,
        label: t(`admin.ops.alertRules.metricGroups.${group}`),
        disabled: true,
        kind: 'group'
      },
      ...items.map((m) => ({ value: m.type, label: m.label }))
    ]
  }

  return [...buildGroup('system'), ...buildGroup('group'), ...buildGroup('account')]
})

const operatorOptions = computed(() => {
  const ops: Operator[] = ['>', '>=', '<', '<=', '==', '!=']
  return ops.map((o) => ({ value: o, label: o }))
})

const severityOptions = computed(() => {
  const sev: OpsSeverity[] = ['P0', 'P1', 'P2', 'P3']
  return sev.map((s) => ({ value: s, label: s }))
})

const windowOptions = computed(() => {
  const windows = [1, 5, 60]
  return windows.map((m) => ({ value: m, label: `${m}m` }))
})

const runtimeAlertPresets = computed<AlertRulePreset[]>(() => [
  {
    key: 'acquire_success',
    name: t('admin.ops.alertRules.presets.acquireSuccess.name'),
    description: t('admin.ops.alertRules.presets.acquireSuccess.description'),
    draft: {
      name: t('admin.ops.alertRules.presets.acquireSuccess.name'),
      description: t('admin.ops.alertRules.presets.acquireSuccess.description'),
      enabled: true,
      metric_type: 'scheduler_acquire_success_rate',
      operator: '<',
      threshold: 75,
      window_minutes: 5,
      sustained_minutes: 3,
      severity: 'P1',
      cooldown_minutes: 15,
      notify_email: true
    }
  },
  {
    key: 'wait_plan_success',
    name: t('admin.ops.alertRules.presets.waitPlanSuccess.name'),
    description: t('admin.ops.alertRules.presets.waitPlanSuccess.description'),
    draft: {
      name: t('admin.ops.alertRules.presets.waitPlanSuccess.name'),
      description: t('admin.ops.alertRules.presets.waitPlanSuccess.description'),
      enabled: true,
      metric_type: 'scheduler_wait_plan_success_rate',
      operator: '<',
      threshold: 60,
      window_minutes: 5,
      sustained_minutes: 3,
      severity: 'P1',
      cooldown_minutes: 15,
      notify_email: true
    }
  },
  {
    key: 'page_density',
    name: t('admin.ops.alertRules.presets.pageDensity.name'),
    description: t('admin.ops.alertRules.presets.pageDensity.description'),
    draft: {
      name: t('admin.ops.alertRules.presets.pageDensity.name'),
      description: t('admin.ops.alertRules.presets.pageDensity.description'),
      enabled: true,
      metric_type: 'scheduler_index_page_density',
      operator: '<',
      threshold: 8,
      window_minutes: 1,
      sustained_minutes: 2,
      severity: 'P2',
      cooldown_minutes: 10,
      notify_email: true
    }
  },
  {
    key: 'idempotency_latency',
    name: t('admin.ops.alertRules.presets.idempotencyLatency.name'),
    description: t('admin.ops.alertRules.presets.idempotencyLatency.description'),
    draft: {
      name: t('admin.ops.alertRules.presets.idempotencyLatency.name'),
      description: t('admin.ops.alertRules.presets.idempotencyLatency.description'),
      enabled: true,
      metric_type: 'idempotency_processing_avg_ms',
      operator: '>',
      threshold: 80,
      window_minutes: 1,
      sustained_minutes: 3,
      severity: 'P2',
      cooldown_minutes: 15,
      notify_email: true
    }
  }
])

const existingRuleSemanticKeys = computed(() => {
  return new Set(rules.value.map((rule) => buildRuleSemanticKey(rule)))
})

const existingRuntimePresetKeys = computed(() => {
  const existing = existingRuleSemanticKeys.value
  return new Set(
    runtimeAlertPresets.value
      .filter((preset) => existing.has(buildRuleSemanticKey(preset.draft)))
      .map((preset) => preset.key)
  )
})

const missingRuntimeAlertPresets = computed(() => {
  const existing = existingRuleSemanticKeys.value
  return runtimeAlertPresets.value.filter((preset) => !existing.has(buildRuleSemanticKey(preset.draft)))
})

function newRuleDraft(): AlertRule {
  return {
    name: '',
    description: '',
    enabled: true,
    metric_type: 'error_rate',
    operator: '>',
    threshold: 1,
    window_minutes: 1,
    sustained_minutes: 2,
    severity: 'P1',
    cooldown_minutes: 10,
    notify_email: true
  }
}

function openCreate() {
  editingId.value = null
  draft.value = newRuleDraft()
  showEditor.value = true
}

function openPreset(preset: AlertRulePreset) {
  editingId.value = null
  draft.value = JSON.parse(JSON.stringify(preset.draft))
  showEditor.value = true
}

async function createRecommendedPresets() {
  if (creatingPresets.value) return
  const missing = missingRuntimeAlertPresets.value
  if (missing.length === 0) {
    appStore.showInfo(t('admin.ops.alertRules.presets.allCreated'))
    return
  }

  const requestSequence = ++rulesRequestSequence
  loading.value = false
  creatingPresets.value = true
  try {
    const settled = await Promise.allSettled(
      missing.map(async (preset) => {
        await opsAPI.createAlertRule(JSON.parse(JSON.stringify(preset.draft)))
        return preset
      })
    )

    const result: PresetCreationResult = {
      succeeded: [],
      failed: []
    }

    for (let index = 0; index < settled.length; index++) {
      const outcome = settled[index]
      const preset = missing[index]
      if (outcome.status === 'fulfilled') {
        result.succeeded.push(outcome.value)
        continue
      }
      const detail = resolveRequestErrorMessage(
        outcome.reason,
        t('admin.ops.alertRules.presets.createFailed')
      )
      result.failed.push({ preset, detail })
    }

    if (requestSequence !== rulesRequestSequence) return
    await load(requestSequence)
    if (requestSequence !== rulesRequestSequence) return

    if (result.failed.length === 0) {
      appStore.showSuccess(t('admin.ops.alertRules.presets.createSuccess', { count: result.succeeded.length }))
      return
    }

    console.error('[OpsAlertRulesCard] Failed to create some recommended presets', result.failed)
    if (result.succeeded.length === 0) {
      appStore.showError(result.failed[0]?.detail || t('admin.ops.alertRules.presets.createFailed'))
      return
    }

    appStore.showWarning(
      t('admin.ops.alertRules.presets.createPartial', {
        success: result.succeeded.length,
        failed: result.failed.length
      })
    )
  } finally {
    if (requestSequence === rulesRequestSequence) {
      creatingPresets.value = false
    }
  }
}

function openEdit(rule: AlertRule) {
  editingId.value = rule.id ?? null
  draft.value = JSON.parse(JSON.stringify(rule))
  showEditor.value = true
}

const editorValidation = computed(() => {
  const errors: string[] = []
  const r = draft.value
  if (!r) return { valid: true, errors }
  if (!r.name || !r.name.trim()) errors.push(t('admin.ops.alertRules.validation.nameRequired'))
  if (!r.metric_type) errors.push(t('admin.ops.alertRules.validation.metricRequired'))
  if (groupMetricTypes.has(r.metric_type) && !parsePositiveInt(r.filters?.group_id)) {
    errors.push(t('admin.ops.alertRules.validation.groupIdRequired'))
  }
  if (!r.operator) errors.push(t('admin.ops.alertRules.validation.operatorRequired'))
  if (!(typeof r.threshold === 'number' && Number.isFinite(r.threshold)))
    errors.push(t('admin.ops.alertRules.validation.thresholdRequired'))
  if (!(typeof r.window_minutes === 'number' && Number.isFinite(r.window_minutes) && [1, 5, 60].includes(r.window_minutes))) {
    errors.push(t('admin.ops.alertRules.validation.windowRange'))
  }
  if (!(typeof r.sustained_minutes === 'number' && Number.isFinite(r.sustained_minutes) && r.sustained_minutes >= 1 && r.sustained_minutes <= 1440)) {
    errors.push(t('admin.ops.alertRules.validation.sustainedRange'))
  }
  if (!(typeof r.cooldown_minutes === 'number' && Number.isFinite(r.cooldown_minutes) && r.cooldown_minutes >= 0 && r.cooldown_minutes <= 1440)) {
    errors.push(t('admin.ops.alertRules.validation.cooldownRange'))
  }
  return { valid: errors.length === 0, errors }
})

async function save() {
  if (!draft.value) return
  if (!editorValidation.value.valid) {
    appStore.showError(editorValidation.value.errors[0] || t('admin.ops.alertRules.validation.invalid'))
    return
  }
  const requestSequence = ++rulesRequestSequence
  loading.value = false
  saving.value = true
  try {
    if (editingId.value) {
      await opsAPI.updateAlertRule(editingId.value, draft.value)
    } else {
      await opsAPI.createAlertRule(draft.value)
    }
    if (requestSequence !== rulesRequestSequence) return
    showEditor.value = false
    draft.value = null
    editingId.value = null
    await load(requestSequence)
    if (requestSequence !== rulesRequestSequence) return
    appStore.showSuccess(t('admin.ops.alertRules.saveSuccess'))
  } catch (err: unknown) {
    if (requestSequence !== rulesRequestSequence) return
    console.error('[OpsAlertRulesCard] Failed to save rule', err)
    appStore.showError(resolveRequestErrorMessage(err, t('admin.ops.alertRules.saveFailed')))
  } finally {
    if (requestSequence === rulesRequestSequence) {
      saving.value = false
    }
  }
}

const showDeleteConfirm = ref(false)
const pendingDelete = ref<AlertRule | null>(null)

function requestDelete(rule: AlertRule) {
  pendingDelete.value = rule
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  if (!pendingDelete.value?.id) return
  const requestSequence = ++rulesRequestSequence
  loading.value = false
  deleting.value = true
  try {
    await opsAPI.deleteAlertRule(pendingDelete.value.id)
    if (requestSequence !== rulesRequestSequence) return
    showDeleteConfirm.value = false
    pendingDelete.value = null
    await load(requestSequence)
    if (requestSequence !== rulesRequestSequence) return
    appStore.showSuccess(t('admin.ops.alertRules.deleteSuccess'))
  } catch (err: unknown) {
    if (requestSequence !== rulesRequestSequence) return
    console.error('[OpsAlertRulesCard] Failed to delete rule', err)
    appStore.showError(resolveRequestErrorMessage(err, t('admin.ops.alertRules.deleteFailed')))
  } finally {
    if (requestSequence === rulesRequestSequence) {
      deleting.value = false
    }
  }
}

function cancelDelete() {
  showDeleteConfirm.value = false
  pendingDelete.value = null
}
</script>

<template>
  <div class="ops-alert-rules-card">
    <div class="ops-alert-rules-card__header">
      <div class="ops-alert-rules-card__header-copy">
        <h3 class="ops-alert-rules-card__title">{{ t('admin.ops.alertRules.title') }}</h3>
        <p class="ops-alert-rules-card__description ops-alert-rules-card__description--header">{{ t('admin.ops.alertRules.description') }}</p>
      </div>

      <div class="ops-alert-rules-card__header-actions">
        <button class="btn btn-sm btn-primary" :disabled="loading || saving || creatingPresets || deleting" @click="openCreate">
          {{ t('admin.ops.alertRules.create') }}
        </button>
        <button
          class="ops-alert-rules-card__refresh btn btn-secondary btn-sm"
          :disabled="loading || saving || creatingPresets || deleting"
          @click="load()"
        >
          <svg class="ops-alert-rules-card__refresh-icon" :class="{ 'animate-spin': loading }" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          {{ t('common.refresh') }}
        </button>
      </div>
    </div>

    <div class="ops-alert-rules-card__presets">
      <div class="ops-alert-rules-card__presets-header">
        <div>
          <div class="ops-alert-rules-card__text-strong ops-alert-rules-card__text-strong--compact">{{ t('admin.ops.alertRules.presets.title') }}</div>
          <div class="ops-alert-rules-card__description ops-alert-rules-card__description--compact">
            {{ t('admin.ops.alertRules.presets.description') }}
          </div>
        </div>
        <button
          class="btn btn-sm btn-secondary"
          :disabled="creatingPresets || missingRuntimeAlertPresets.length === 0"
          @click="createRecommendedPresets"
        >
          {{ t('admin.ops.alertRules.presets.createAll') }}
        </button>
      </div>
      <div class="ops-alert-rules-card__presets-grid">
        <button
          v-for="preset in runtimeAlertPresets"
          :key="preset.key"
          type="button"
          class="ops-alert-rules-card__preset"
          @click="openPreset(preset)"
        >
          <div class="ops-alert-rules-card__preset-header">
            <div class="ops-alert-rules-card__text-strong ops-alert-rules-card__text-strong--compact">{{ preset.name }}</div>
            <span
              v-if="existingRuntimePresetKeys.has(preset.key)"
              class="ops-alert-rules-card__preset-badge"
            >
              {{ t('admin.ops.alertRules.presets.created') }}
            </span>
          </div>
          <div class="ops-alert-rules-card__description ops-alert-rules-card__description--compact ops-alert-rules-card__description--preset">{{ preset.description }}</div>
        </button>
      </div>
    </div>

    <div v-if="loading" class="ops-alert-rules-card__loading ops-alert-rules-card__description ops-alert-rules-card__state">
      {{ t('admin.ops.alertRules.loading') }}
    </div>

    <div v-else-if="sortedRules.length === 0" class="ops-alert-rules-card__empty">
      <div class="ops-alert-rules-card__text-strong ops-alert-rules-card__text-strong--title">
        {{ t('admin.ops.alertRules.emptyState.title') }}
      </div>
      <p class="ops-alert-rules-card__description ops-alert-rules-card__description--empty">
        {{ t('admin.ops.alertRules.emptyState.description') }}
      </p>
      <div class="ops-alert-rules-card__empty-actions">
        <button
          class="btn btn-sm btn-primary"
          :disabled="creatingPresets || missingRuntimeAlertPresets.length === 0"
          @click="createRecommendedPresets"
        >
          {{ t('admin.ops.alertRules.presets.createAll') }}
        </button>
        <button class="btn btn-sm btn-secondary" @click="openCreate">
          {{ t('admin.ops.alertRules.create') }}
        </button>
      </div>
    </div>

    <div v-else class="ops-alert-rules-card__table-shell">
      <div class="ops-alert-rules-card__table-scroll">
        <table class="ops-alert-rules-card__table">
          <thead class="ops-alert-rules-card__table-head">
            <tr>
              <th class="ops-alert-rules-card__table-header ops-alert-rules-card__table-header--regular">
                {{ t('admin.ops.alertRules.table.name') }}
              </th>
              <th class="ops-alert-rules-card__table-header ops-alert-rules-card__table-header--regular">
                {{ t('admin.ops.alertRules.table.metric') }}
              </th>
              <th class="ops-alert-rules-card__table-header ops-alert-rules-card__table-header--regular">
                {{ t('admin.ops.alertRules.table.severity') }}
              </th>
              <th class="ops-alert-rules-card__table-header ops-alert-rules-card__table-header--regular">
                {{ t('admin.ops.alertRules.table.enabled') }}
              </th>
              <th class="ops-alert-rules-card__table-header ops-alert-rules-card__table-header--regular ops-alert-rules-card__table-header--end">
                {{ t('admin.ops.alertRules.table.actions') }}
              </th>
            </tr>
          </thead>
          <tbody class="ops-alert-rules-card__table-body">
            <tr v-for="row in sortedRules" :key="row.id" class="ops-alert-rules-card__table-row">
              <td class="ops-alert-rules-card__table-cell ops-alert-rules-card__table-cell--regular">
                <div class="ops-alert-rules-card__text-strong ops-alert-rules-card__text-strong--compact">{{ row.name }}</div>
                <div v-if="row.description" class="ops-alert-rules-card__description ops-alert-rules-card__description--clamped">
                  {{ row.description }}
                </div>
                <div v-if="row.updated_at" class="ops-alert-rules-card__text-soft ops-alert-rules-card__text-soft--compact">
                  {{ formatDateTime(row.updated_at) }}
                </div>
              </td>
              <td class="ops-alert-rules-card__table-cell ops-alert-rules-card__table-cell--regular ops-alert-rules-card__text-body ops-alert-rules-card__table-cell--nowrap ops-alert-rules-card__table-cell--compact">
                <span class="ops-alert-rules-card__mono">{{ row.metric_type }}</span>
                <span class="ops-alert-rules-card__text-soft ops-alert-rules-card__text-soft--inline">{{ row.operator }}</span>
                <span class="ops-alert-rules-card__mono">{{ row.threshold }}</span>
              </td>
              <td class="ops-alert-rules-card__table-cell ops-alert-rules-card__table-cell--regular ops-alert-rules-card__text-body ops-alert-rules-card__table-cell--nowrap ops-alert-rules-card__table-cell--compact ops-alert-rules-card__text-body--strong">
                {{ row.severity }}
              </td>
              <td class="ops-alert-rules-card__table-cell ops-alert-rules-card__table-cell--regular ops-alert-rules-card__text-body ops-alert-rules-card__table-cell--nowrap ops-alert-rules-card__table-cell--compact">
                {{ row.enabled ? t('common.enabled') : t('common.disabled') }}
              </td>
              <td class="ops-alert-rules-card__table-cell ops-alert-rules-card__table-cell--regular ops-alert-rules-card__table-cell--nowrap ops-alert-rules-card__table-cell--end ops-alert-rules-card__table-cell--compact">
                <button class="btn btn-sm btn-secondary" @click="openEdit(row)">{{ t('common.edit') }}</button>
                <button class="ops-alert-rules-card__danger-action btn btn-sm btn-danger" @click="requestDelete(row)">{{ t('common.delete') }}</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <BaseDialog
      :show="showEditor"
      :title="editingId ? t('admin.ops.alertRules.editTitle') : t('admin.ops.alertRules.createTitle')"
      width="wide"
      @close="showEditor = false"
    >
      <div class="ops-alert-rules-card__editor">
        <div v-if="!editorValidation.valid" class="ops-alert-rules-card__validation">
          <div class="ops-alert-rules-card__validation-title">{{ t('admin.ops.alertRules.validation.title') }}</div>
          <ul class="ops-alert-rules-card__validation-list">
            <li v-for="e in editorValidation.errors" :key="e">{{ e }}</li>
          </ul>
        </div>

        <div class="ops-alert-rules-card__editor-grid">
          <div class="ops-alert-rules-card__editor-field ops-alert-rules-card__editor-field--full">
            <label class="input-label">{{ t('admin.ops.alertRules.form.name') }}</label>
            <input v-model="draft!.name" class="input" type="text" />
          </div>

          <div class="ops-alert-rules-card__editor-field ops-alert-rules-card__editor-field--full">
            <label class="input-label">{{ t('admin.ops.alertRules.form.description') }}</label>
            <input v-model="draft!.description" class="input" type="text" />
          </div>

          <div>
            <label class="input-label">{{ t('admin.ops.alertRules.form.metric') }}</label>
            <Select v-model="draft!.metric_type" :options="metricOptions" />
            <div v-if="selectedMetricDefinition" class="ops-alert-rules-card__description ops-alert-rules-card__description--hint">
              <p>{{ selectedMetricDefinition.description }}</p>
              <p>
                {{
                  t('admin.ops.alertRules.hints.recommended', {
                    operator: selectedMetricDefinition.recommendedOperator,
                    threshold: selectedMetricDefinition.recommendedThreshold,
                    unit: selectedMetricDefinition.unit || ''
                  })
                }}
              </p>
            </div>
          </div>

          <div>
            <label class="input-label">{{ t('admin.ops.alertRules.form.operator') }}</label>
            <Select v-model="draft!.operator" :options="operatorOptions" />
          </div>

          <div class="ops-alert-rules-card__editor-field ops-alert-rules-card__editor-field--full">
            <label class="input-label">
              {{ t('admin.ops.alertRules.form.groupId') }}
              <span v-if="isGroupMetricSelected" class="ops-alert-rules-card__required">*</span>
            </label>
            <Select
              v-model="draftGroupId"
              :options="groupOptions"
              searchable
              :placeholder="t('admin.ops.alertRules.form.groupPlaceholder')"
              :error="isGroupMetricSelected && !draftGroupId"
            />
            <p class="ops-alert-rules-card__description ops-alert-rules-card__description--hint">
              {{ isGroupMetricSelected ? t('admin.ops.alertRules.hints.groupRequired') : t('admin.ops.alertRules.hints.groupOptional') }}
            </p>
          </div>

          <div>
            <label class="input-label">{{ t('admin.ops.alertRules.form.threshold') }}</label>
            <input v-model.number="draft!.threshold" class="input" type="number" />
          </div>

          <div>
            <label class="input-label">{{ t('admin.ops.alertRules.form.severity') }}</label>
            <Select v-model="draft!.severity" :options="severityOptions" />
          </div>

          <div>
            <label class="input-label">{{ t('admin.ops.alertRules.form.window') }}</label>
            <Select v-model="draft!.window_minutes" :options="windowOptions" />
          </div>

          <div>
            <label class="input-label">{{ t('admin.ops.alertRules.form.sustained') }}</label>
            <input v-model.number="draft!.sustained_minutes" class="input" type="number" min="1" max="1440" />
          </div>

          <div>
            <label class="input-label">{{ t('admin.ops.alertRules.form.cooldown') }}</label>
            <input v-model.number="draft!.cooldown_minutes" class="input" type="number" min="0" max="1440" />
          </div>

          <div class="ops-alert-rules-card__toggle-row ops-alert-rules-card__editor-field--full">
            <span class="ops-alert-rules-card__text-body ops-alert-rules-card__text-body--strong">{{ t('admin.ops.alertRules.form.enabled') }}</span>
            <input v-model="draft!.enabled" type="checkbox" class="ops-alert-rules-card__checkbox" />
          </div>

          <div class="ops-alert-rules-card__toggle-row ops-alert-rules-card__editor-field--full">
            <span class="ops-alert-rules-card__text-body ops-alert-rules-card__text-body--strong">{{ t('admin.ops.alertRules.form.notifyEmail') }}</span>
            <input v-model="draft!.notify_email" type="checkbox" class="ops-alert-rules-card__checkbox" />
          </div>
        </div>
      </div>

      <template #footer>
        <div class="ops-alert-rules-card__footer">
          <button class="btn btn-secondary" :disabled="saving" @click="showEditor = false">
            {{ t('common.cancel') }}
          </button>
          <button class="btn btn-primary" :disabled="saving" @click="save">
            {{ saving ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <ConfirmDialog
      :show="showDeleteConfirm"
      :title="t('admin.ops.alertRules.deleteConfirmTitle')"
      :message="t('admin.ops.alertRules.deleteConfirmMessage')"
      :confirmText="t('common.delete')"
      :cancelText="t('common.cancel')"
      @confirm="confirmDelete"
      @cancel="cancelDelete"
    />
  </div>
</template>

<style scoped>
.ops-alert-rules-card {
  padding: var(--theme-ops-card-padding);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  border-radius: var(--theme-surface-radius);
}

.ops-alert-rules-card__title,
.ops-alert-rules-card__text-strong {
  color: var(--theme-page-text);
}

.ops-alert-rules-card__header {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--theme-ops-alert-rules-header-gap);
  margin-bottom: 1rem;
}

.ops-alert-rules-card__header-copy {
  min-width: 0;
}

.ops-alert-rules-card__title {
  font-size: 0.875rem;
  font-weight: 700;
}

.ops-alert-rules-card__description {
  color: var(--theme-page-muted);
}

.ops-alert-rules-card__description--header {
  margin-top: 0.25rem;
  font-size: 0.75rem;
}

.ops-alert-rules-card__description--compact {
  font-size: 0.6875rem;
}

.ops-alert-rules-card__description--preset {
  margin-top: 0.25rem;
}

.ops-alert-rules-card__description--empty {
  max-width: 42rem;
  margin: var(--theme-ops-alert-rules-empty-gap) auto 0;
  font-size: 0.75rem;
}

.ops-alert-rules-card__description--clamped {
  margin-top: 0.125rem;
  font-size: 0.6875rem;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.ops-alert-rules-card__description--hint {
  margin-top: 0.25rem;
  font-size: 0.75rem;
}

.ops-alert-rules-card__description--hint p + p {
  margin-top: 0.125rem;
}

.ops-alert-rules-card__loading {
  padding-block: calc(var(--theme-ops-card-padding) * 1.5);
}

.ops-alert-rules-card__state {
  text-align: center;
  font-size: 0.875rem;
}

.ops-alert-rules-card__text-body {
  color: color-mix(in srgb, var(--theme-page-text) 80%, var(--theme-page-muted));
}

.ops-alert-rules-card__text-body--strong {
  font-size: 0.75rem;
  font-weight: 700;
}

.ops-alert-rules-card__text-soft {
  color: color-mix(in srgb, var(--theme-page-muted) 76%, transparent);
}

.ops-alert-rules-card__text-soft--compact {
  margin-top: 0.25rem;
  font-size: 0.625rem;
}

.ops-alert-rules-card__text-soft--inline {
  margin-inline: 0.25rem;
}

.ops-alert-rules-card__refresh {
  display: inline-flex;
  align-items: center;
  gap: var(--theme-ops-alert-rules-control-gap);
  box-shadow: var(--theme-card-shadow);
}

.ops-alert-rules-card__refresh-icon {
  width: 0.875rem;
  height: 0.875rem;
}

.ops-alert-rules-card__header-actions {
  display: flex;
  align-items: center;
  gap: var(--theme-ops-alert-rules-control-gap);
}

.ops-alert-rules-card__presets {
  margin-bottom: 1rem;
  padding: var(--theme-ops-panel-padding);
  border-radius: var(--theme-select-panel-radius);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 70%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 90%, var(--theme-surface));
}

.ops-alert-rules-card__presets-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--theme-ops-alert-rules-panel-gap);
  margin-bottom: 0.5rem;
}

.ops-alert-rules-card__presets-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--theme-ops-alert-rules-grid-gap);
}

.ops-alert-rules-card__preset {
  text-align: left;
  padding: calc(var(--theme-ops-panel-padding) * 0.8);
  border-radius: var(--theme-button-radius);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 64%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 78%, var(--theme-surface));
  transition: background-color 0.2s ease, border-color 0.2s ease, transform 0.2s ease;
}

.ops-alert-rules-card__preset:hover {
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 8%, var(--theme-surface));
  border-color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 28%, transparent);
  transform: translateY(-1px);
}

.ops-alert-rules-card__preset-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.5rem;
}

.ops-alert-rules-card__preset-badge {
  padding: 0.15rem 0.45rem;
  border-radius: 999px;
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 16%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 86%, var(--theme-page-text));
  font-size: 0.625rem;
  font-weight: 700;
}

.ops-alert-rules-card__empty {
  padding: calc(var(--theme-table-mobile-empty-padding) * 0.67);
  border: 1px dashed color-mix(in srgb, var(--theme-card-border) 78%, transparent);
  border-radius: var(--theme-select-panel-radius);
  color: var(--theme-page-muted);
  background: color-mix(in srgb, var(--theme-surface-soft) 64%, var(--theme-surface));
  text-align: center;
}

.ops-alert-rules-card__empty-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: var(--theme-ops-alert-rules-control-gap);
  margin-top: 1rem;
}

.ops-alert-rules-card__table-shell {
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  border-radius: var(--theme-select-panel-radius);
  background: var(--theme-surface);
}

.ops-alert-rules-card__table-scroll {
  max-height: var(--theme-ops-table-max-height);
  overflow-y: auto;
}

.ops-alert-rules-card__table {
  width: 100%;
  min-width: var(--theme-ops-table-min-width);
}

.ops-alert-rules-card__table-header--regular,
.ops-alert-rules-card__table-cell--regular {
  padding:
    var(--theme-ops-table-cell-padding-y)
    var(--theme-ops-table-cell-padding-x);
}

.ops-alert-rules-card__table-head {
  background: var(--theme-table-head-bg);
  position: sticky;
  top: 0;
  z-index: 10;
}

.ops-alert-rules-card__table-header {
  color: var(--theme-table-head-text);
  text-align: left;
  font-size: 0.6875rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.ops-alert-rules-card__table-header--end,
.ops-alert-rules-card__table-cell--end {
  text-align: right;
}

.ops-alert-rules-card__table-row td {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 64%, transparent);
}

.ops-alert-rules-card__table-body tr:first-child td {
  border-top: none;
}

.ops-alert-rules-card__table-row:hover {
  background: color-mix(in srgb, var(--theme-table-row-hover) 100%, var(--theme-surface));
}

.ops-alert-rules-card__validation {
  padding: var(--theme-ops-panel-padding);
  border-radius: var(--theme-select-panel-radius);
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
  font-size: 0.75rem;
}

.ops-alert-rules-card__validation-title {
  font-weight: 700;
}

.ops-alert-rules-card__validation-list {
  margin-top: 0.25rem;
  padding-left: 1.25rem;
  list-style: disc;
}

.ops-alert-rules-card__required {
  margin-left: 0.25rem;
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.ops-alert-rules-card__toggle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--theme-ops-alert-rules-toggle-gap);
  padding:
    var(--theme-ops-table-cell-padding-y)
    var(--theme-ops-table-cell-padding-x);
  border-radius: var(--theme-select-panel-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 78%, var(--theme-surface));
}

.ops-alert-rules-card__checkbox {
  accent-color: var(--theme-accent);
  width: 1rem;
  height: 1rem;
  border-radius: 0.25rem;
}

.ops-alert-rules-card__text-strong--compact {
  font-size: 0.75rem;
  font-weight: 700;
}

.ops-alert-rules-card__text-strong--title {
  font-size: 0.875rem;
  font-weight: 700;
}

.ops-alert-rules-card__table-cell--nowrap {
  white-space: nowrap;
}

.ops-alert-rules-card__table-cell--compact {
  font-size: 0.75rem;
}

.ops-alert-rules-card__mono {
  font-family: var(--theme-font-mono);
}

.ops-alert-rules-card__editor {
  display: flex;
  flex-direction: column;
  gap: var(--theme-ops-alert-rules-panel-gap);
}

.ops-alert-rules-card__editor-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1rem;
}

.ops-alert-rules-card__editor-field {
  min-width: 0;
}

.ops-alert-rules-card__danger-action {
  margin-left: 0.5rem;
}

.ops-alert-rules-card__footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--theme-ops-alert-rules-control-gap);
}

@media (min-width: 768px) {
  .ops-alert-rules-card__presets-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .ops-alert-rules-card__editor-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .ops-alert-rules-card__editor-field--full {
    grid-column: 1 / -1;
  }
}
</style>
