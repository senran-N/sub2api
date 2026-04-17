<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import OpsErrorLogTable from './OpsErrorLogTable.vue'
import { opsAPI, type OpsErrorLog } from '@/api/admin/ops'

interface Props {
  show: boolean
  timeRange: string
  platform?: string
  groupId?: number | null
  errorType: 'request' | 'upstream'
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'openErrorDetail', errorId: number): void
}>()

const { t } = useI18n()

const loading = ref(false)
const rows = ref<OpsErrorLog[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

const q = ref('')
const statusCode = ref<number | 'other' | null>(null)
const phase = ref<string>('')
const errorOwner = ref<string>('')
const viewMode = ref<'errors' | 'excluded' | 'all'>('errors')
let fetchSequence = 0

const modalTitle = computed(() => {
  return props.errorType === 'upstream' ? t('admin.ops.errorDetails.upstreamErrors') : t('admin.ops.errorDetails.requestErrors')
})

const statusCodeSelectOptions = computed(() => {
  const codes = [400, 401, 403, 404, 409, 422, 429, 500, 502, 503, 504, 529]
  return [
    { value: null, label: t('common.all') },
    ...codes.map((c) => ({ value: c, label: String(c) })),
    { value: 'other', label: t('admin.ops.errorDetails.statusCodeOther') || 'Other' }
  ]
})

const ownerSelectOptions = computed(() => {
  return [
    { value: '', label: t('common.all') },
    { value: 'provider', label: t('admin.ops.errorDetails.owner.provider') || 'provider' },
    { value: 'client', label: t('admin.ops.errorDetails.owner.client') || 'client' },
    { value: 'platform', label: t('admin.ops.errorDetails.owner.platform') || 'platform' }
  ]
})


const viewModeSelectOptions = computed(() => {
  return [
    { value: 'errors', label: t('admin.ops.errorDetails.viewErrors') || 'errors' },
    { value: 'excluded', label: t('admin.ops.errorDetails.viewExcluded') || 'excluded' },
    { value: 'all', label: t('common.all') }
  ]
})

type StatusCodeFilter = number | 'other' | null

function normalizeStatusCode(value: unknown): StatusCodeFilter {
  if (value === 'other') {
    return value
  }

  if (value == null) {
    return null
  }

  return typeof value === 'number' ? value : Number(value)
}

function normalizeViewMode(value: unknown): 'errors' | 'excluded' | 'all' {
  if (value === 'excluded' || value === 'all') {
    return value
  }

  return 'errors'
}

const phaseSelectOptions = computed(() => {
  const options = [
    { value: '', label: t('common.all') },
    { value: 'request', label: t('admin.ops.errorDetails.phase.request') || 'request' },
    { value: 'auth', label: t('admin.ops.errorDetails.phase.auth') || 'auth' },
    { value: 'routing', label: t('admin.ops.errorDetails.phase.routing') || 'routing' },
    { value: 'upstream', label: t('admin.ops.errorDetails.phase.upstream') || 'upstream' },
    { value: 'network', label: t('admin.ops.errorDetails.phase.network') || 'network' },
    { value: 'internal', label: t('admin.ops.errorDetails.phase.internal') || 'internal' }
  ]
  return options
})

function close() {
  emit('update:show', false)
}

async function fetchErrorLogs() {
  if (!props.show) return

  const requestSequence = ++fetchSequence
  const errorType = props.errorType
  loading.value = true
  try {
    const params: Record<string, string | number> = {
      page: page.value,
      page_size: pageSize.value,
      time_range: props.timeRange,
      view: viewMode.value
    }

    const platform = String(props.platform || '').trim()
    if (platform) params.platform = platform
    if (typeof props.groupId === 'number' && props.groupId > 0) params.group_id = props.groupId

    if (q.value.trim()) params.q = q.value.trim()
    if (statusCode.value === 'other') params.status_codes_other = '1'
    else if (typeof statusCode.value === 'number') params.status_codes = String(statusCode.value)

    const phaseVal = String(phase.value || '').trim()
    if (phaseVal) params.phase = phaseVal

    const ownerVal = String(errorOwner.value || '').trim()
    if (ownerVal) params.error_owner = ownerVal

    const res = errorType === 'upstream'
      ? await opsAPI.listUpstreamErrors(params)
      : await opsAPI.listRequestErrors(params)
    if (
      requestSequence !== fetchSequence ||
      !props.show ||
      props.errorType !== errorType
    ) {
      return
    }
    rows.value = res.items || []
    total.value = res.total || 0
  } catch (err) {
    if (
      requestSequence !== fetchSequence ||
      !props.show ||
      props.errorType !== errorType
    ) {
      return
    }
    console.error('[OpsErrorDetailsModal] Failed to fetch error logs', err)
    rows.value = []
    total.value = 0
  } finally {
    if (requestSequence === fetchSequence) {
      loading.value = false
    }
  }
}

function resetFilters() {
  q.value = ''
  statusCode.value = null
  phase.value = props.errorType === 'upstream' ? 'upstream' : ''
  errorOwner.value = ''
  viewMode.value = 'errors'
  page.value = 1
  fetchErrorLogs()
}

let searchTimeout: number | null = null
watch(
  () => [props.show, props.errorType] as const,
  ([open]) => {
    if (!open) {
      fetchSequence++
      loading.value = false
      rows.value = []
      total.value = 0
      if (searchTimeout) {
        window.clearTimeout(searchTimeout)
        searchTimeout = null
      }
      return
    }
    page.value = 1
    pageSize.value = 10
    resetFilters()
  },
  { immediate: true }
)

watch(
  () => [props.timeRange, props.platform, props.groupId] as const,
  () => {
    if (!props.show) return
    page.value = 1
    fetchErrorLogs()
  }
)

watch(
  () => [page.value, pageSize.value] as const,
  () => {
    if (!props.show) return
    fetchErrorLogs()
  }
)

watch(
  () => q.value,
  () => {
    if (!props.show) return
    if (searchTimeout) window.clearTimeout(searchTimeout)
    searchTimeout = window.setTimeout(() => {
      page.value = 1
      fetchErrorLogs()
    }, 350)
  }
)

watch(
  () => [statusCode.value, phase.value, errorOwner.value, viewMode.value] as const,
  () => {
    if (!props.show) return
    page.value = 1
    fetchErrorLogs()
  }
)
</script>

<template>
  <BaseDialog :show="show" :title="modalTitle" width="full" @close="close">
    <div class="flex h-full min-h-0 flex-col">
      <div class="ops-error-details-modal__filters">
        <div class="ops-error-details-modal__filter-grid">
          <div class="ops-error-details-modal__filter-col ops-error-details-modal__filter-col--search compact-select">
            <div class="relative group">
              <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
                <svg
                  class="ops-error-details-modal__search-icon h-3.5 w-3.5 transition-colors"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
              </div>
              <input
                v-model="q"
                type="text"
                class="ops-error-details-modal__search-input"
                :placeholder="t('admin.ops.errorDetails.searchPlaceholder')"
              />
            </div>
          </div>

          <div class="ops-error-details-modal__filter-col compact-select">
            <Select :model-value="statusCode" :options="statusCodeSelectOptions" @update:model-value="statusCode = normalizeStatusCode($event)" />
          </div>

          <div class="ops-error-details-modal__filter-col compact-select">
            <Select :model-value="phase" :options="phaseSelectOptions" @update:model-value="phase = String($event ?? '')" />
          </div>

          <div class="ops-error-details-modal__filter-col compact-select">
            <Select :model-value="errorOwner" :options="ownerSelectOptions" @update:model-value="errorOwner = String($event ?? '')" />
          </div>



          <div class="ops-error-details-modal__filter-col compact-select">
            <Select :model-value="viewMode" :options="viewModeSelectOptions" @update:model-value="viewMode = normalizeViewMode($event)" />
          </div>

          <div class="ops-error-details-modal__filter-col ops-error-details-modal__filter-col--action">
            <button type="button" class="ops-error-details-modal__reset" @click="resetFilters">
              {{ t('common.reset') }}
            </button>
          </div>
        </div>
      </div>

      <div class="flex min-h-0 flex-1 flex-col">
        <div class="ops-error-details-modal__meta mb-2 flex-shrink-0 text-xs">
          {{ t('admin.ops.errorDetails.total') }} {{ total }}
        </div>

          <OpsErrorLogTable
            class="min-h-0 flex-1"
            :rows="rows"
            :total="total"
            :loading="loading"
            :page="page"
            :page-size="pageSize"
            @openErrorDetail="emit('openErrorDetail', $event)"

            @update:page="page = $event"
            @update:pageSize="pageSize = $event"
          />

      </div>
    </div>
  </BaseDialog>
</template>

<style scoped>
.ops-error-details-modal__filters {
  border-bottom: 1px solid var(--theme-page-border);
  padding: calc(var(--theme-ops-panel-padding) * 0.75) calc(var(--theme-ops-panel-padding) * 0.5);
  margin-bottom: var(--theme-ops-panel-padding);
}

.ops-error-details-modal__filter-grid {
  display: grid;
  grid-template-columns: repeat(8, minmax(0, 1fr));
  gap: calc(var(--theme-ops-panel-padding) * 0.5);
}

.ops-error-details-modal__filter-col {
  min-width: 0;
}

.ops-error-details-modal__filter-col--search {
  grid-column: span 2;
}

.ops-error-details-modal__filter-col--action {
  display: flex;
  align-items: flex-end;
  justify-content: flex-end;
}

.ops-error-details-modal__search-icon {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.group:focus-within .ops-error-details-modal__search-icon {
  color: var(--theme-accent);
}

.ops-error-details-modal__search-input {
  width: 100%;
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 84%, transparent);
  border-radius: calc(var(--theme-button-radius) * 1.1);
  background: color-mix(in srgb, var(--theme-surface-soft) 84%, var(--theme-surface));
  color: var(--theme-page-text);
  padding:
    calc(var(--theme-button-padding-y) * 0.8)
    calc(var(--theme-button-padding-x) * 0.85);
  padding-left: calc(var(--theme-button-padding-x) * 1.8 + 0.75rem);
  font-size: 0.75rem;
  transition: border-color 0.2s ease, background 0.2s ease;
}

.ops-error-details-modal__search-input:focus {
  outline: none;
  border-color: var(--theme-accent);
  background: var(--theme-surface);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--theme-accent) 14%, transparent);
}

.ops-error-details-modal__search-input::placeholder {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.ops-error-details-modal__reset {
  padding:
    calc(var(--theme-button-padding-y) * 0.7)
    calc(var(--theme-button-padding-x) * 0.8);
  border-radius: var(--theme-button-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 86%, var(--theme-surface));
  color: color-mix(in srgb, var(--theme-page-text) 82%, transparent);
  font-weight: 600;
  transition: background 0.2s ease, color 0.2s ease;
}

.ops-error-details-modal__reset:hover {
  background: color-mix(in srgb, var(--theme-surface-soft) 72%, var(--theme-surface));
  color: var(--theme-page-text);
}

.ops-error-details-modal__meta {
  color: var(--theme-page-muted);
}

.compact-select :deep(.select-trigger) {
  border-radius: var(--theme-select-action-radius);
  padding:
    calc(var(--theme-button-padding-y) * 0.65)
    calc(var(--theme-button-padding-x) * 0.75);
  font-size: 0.75rem;
}
</style>
