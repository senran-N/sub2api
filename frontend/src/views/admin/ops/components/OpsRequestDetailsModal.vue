<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Pagination from '@/components/common/Pagination.vue'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores'
import { opsAPI, type OpsRequestDetailsParams, type OpsRequestDetail } from '@/api/admin/ops'
import { parseTimeRangeMinutes, formatDateTime } from '../utils/opsFormatters'

export interface OpsRequestDetailsPreset {
  title: string
  kind?: OpsRequestDetailsParams['kind']
  sort?: OpsRequestDetailsParams['sort']
  min_duration_ms?: number
  max_duration_ms?: number
}

interface Props {
  modelValue: boolean
  timeRange: string
  preset: OpsRequestDetailsPreset
  platform?: string
  groupId?: number | null
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'openErrorDetail', errorId: number): void
}>()

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const loading = ref(false)
const items = ref<OpsRequestDetail[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

const close = () => emit('update:modelValue', false)

const rangeLabel = computed(() => {
  const minutes = parseTimeRangeMinutes(props.timeRange)
  if (minutes >= 60) return t('admin.ops.requestDetails.rangeHours', { n: Math.round(minutes / 60) })
  return t('admin.ops.requestDetails.rangeMinutes', { n: minutes })
})

function buildTimeParams(): Pick<OpsRequestDetailsParams, 'start_time' | 'end_time'> {
  const minutes = parseTimeRangeMinutes(props.timeRange)
  const endTime = new Date()
  const startTime = new Date(endTime.getTime() - minutes * 60 * 1000)
  return {
    start_time: startTime.toISOString(),
    end_time: endTime.toISOString()
  }
}

const fetchData = async () => {
  if (!props.modelValue) return
  loading.value = true
  try {
    const params: OpsRequestDetailsParams = {
      ...buildTimeParams(),
      page: page.value,
      page_size: pageSize.value,
      kind: props.preset.kind ?? 'all',
      sort: props.preset.sort ?? 'created_at_desc'
    }

    const platform = (props.platform || '').trim()
    if (platform) params.platform = platform
    if (typeof props.groupId === 'number' && props.groupId > 0) params.group_id = props.groupId

    if (typeof props.preset.min_duration_ms === 'number') params.min_duration_ms = props.preset.min_duration_ms
    if (typeof props.preset.max_duration_ms === 'number') params.max_duration_ms = props.preset.max_duration_ms

    const res = await opsAPI.listRequestDetails(params)
    items.value = res.items || []
    total.value = res.total || 0
  } catch (e: any) {
    console.error('[OpsRequestDetailsModal] Failed to fetch request details', e)
    appStore.showError(e?.message || t('admin.ops.requestDetails.failedToLoad'))
    items.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      page.value = 1
      pageSize.value = 10
      fetchData()
    }
  }
)

watch(
  () => [
    props.timeRange,
    props.platform,
    props.groupId,
    props.preset.kind,
    props.preset.sort,
    props.preset.min_duration_ms,
    props.preset.max_duration_ms
  ],
  () => {
    if (!props.modelValue) return
    page.value = 1
    fetchData()
  }
)

function handlePageChange(next: number) {
  page.value = next
  fetchData()
}

function handlePageSizeChange(next: number) {
  pageSize.value = next
  page.value = 1
  fetchData()
}

async function handleCopyRequestId(requestId: string) {
  const ok = await copyToClipboard(requestId, t('admin.ops.requestDetails.requestIdCopied'))
  if (ok) return
  // `useClipboard` already shows toast on failure; this keeps UX consistent with older ops modal.
  appStore.showWarning(t('admin.ops.requestDetails.copyFailed'))
}

function openErrorDetail(errorId: number | null | undefined) {
  if (!errorId) return
  close()
  emit('openErrorDetail', errorId)
}

const kindBadgeClass = (kind: string) => {
  if (kind === 'error') return 'theme-chip theme-chip--compact theme-chip--danger'
  return 'theme-chip theme-chip--compact theme-chip--success'
}
</script>

<template>
  <BaseDialog :show="modelValue" :title="props.preset.title || t('admin.ops.requestDetails.title')" width="full" @close="close">
    <template #default>
      <div class="ops-request-details-modal flex h-full min-h-0 flex-col">
        <div class="mb-4 flex flex-shrink-0 items-center justify-between">
          <div class="ops-request-details-modal__subtitle text-xs">
            {{ t('admin.ops.requestDetails.rangeLabel', { range: rangeLabel }) }}
          </div>
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            @click="fetchData"
          >
            {{ t('common.refresh') }}
          </button>
        </div>

        <!-- Loading -->
        <div v-if="loading" class="ops-request-details-modal__loading flex flex-1 items-center justify-center">
          <div class="ops-request-details-modal__loading-stack flex flex-col items-center">
            <svg class="ops-request-details-modal__spinner h-8 w-8 animate-spin" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            <span class="ops-request-details-modal__subtitle text-sm font-medium">{{ t('common.loading') }}</span>
          </div>
        </div>

        <!-- Table -->
        <div v-else class="flex min-h-0 flex-1 flex-col">
          <div v-if="items.length === 0" class="ops-request-details-modal__empty text-center">
            <div class="ops-request-details-modal__text-body text-sm font-medium">{{ t('admin.ops.requestDetails.empty') }}</div>
            <div class="ops-request-details-modal__text-soft mt-1 text-xs">{{ t('admin.ops.requestDetails.emptyHint') }}</div>
          </div>

          <div v-else class="ops-request-details-modal__table-shell flex min-h-0 flex-1 flex-col overflow-hidden">
            <div class="min-h-0 flex-1 overflow-auto">
              <table class="ops-request-details-modal__table min-w-full">
                <thead class="ops-request-details-modal__table-head sticky top-0 z-10">
                <tr>
                  <th class="ops-request-details-modal__table-header text-left">
                    {{ t('admin.ops.requestDetails.table.time') }}
                  </th>
                  <th class="ops-request-details-modal__table-header text-left">
                    {{ t('admin.ops.requestDetails.table.kind') }}
                  </th>
                  <th class="ops-request-details-modal__table-header text-left">
                    {{ t('admin.ops.requestDetails.table.platform') }}
                  </th>
                  <th class="ops-request-details-modal__table-header text-left">
                    {{ t('admin.ops.requestDetails.table.model') }}
                  </th>
                  <th class="ops-request-details-modal__table-header text-left">
                    {{ t('admin.ops.requestDetails.table.duration') }}
                  </th>
                  <th class="ops-request-details-modal__table-header text-left">
                    {{ t('admin.ops.requestDetails.table.status') }}
                  </th>
                  <th class="ops-request-details-modal__table-header text-left">
                    {{ t('admin.ops.requestDetails.table.requestId') }}
                  </th>
                  <th class="ops-request-details-modal__table-header ops-request-details-modal__table-header--actions text-right">
                    {{ t('admin.ops.requestDetails.table.actions') }}
                  </th>
                </tr>
              </thead>
              <tbody class="ops-request-details-modal__table-body">
                <tr v-for="(row, idx) in items" :key="idx" class="ops-request-details-modal__table-row">
                  <td class="ops-request-details-modal__table-cell ops-request-details-modal__text-body whitespace-nowrap text-xs">
                    {{ formatDateTime(row.created_at) }}
                  </td>
                  <td class="ops-request-details-modal__table-cell whitespace-nowrap">
                    <span :class="kindBadgeClass(row.kind)">
                      {{ row.kind === 'error' ? t('admin.ops.requestDetails.kind.error') : t('admin.ops.requestDetails.kind.success') }}
                    </span>
                  </td>
                  <td class="ops-request-details-modal__table-cell ops-request-details-modal__text-strong whitespace-nowrap text-xs font-medium">
                    {{ (row.platform || 'unknown').toUpperCase() }}
                  </td>
                  <td class="ops-request-details-modal__table-cell ops-request-details-modal__table-cell--model ops-request-details-modal__text-body truncate text-xs" :title="row.model || ''">
                    {{ row.model || '-' }}
                  </td>
                  <td class="ops-request-details-modal__table-cell ops-request-details-modal__text-body whitespace-nowrap text-xs">
                    {{ typeof row.duration_ms === 'number' ? `${row.duration_ms} ms` : '-' }}
                  </td>
                  <td class="ops-request-details-modal__table-cell ops-request-details-modal__text-body whitespace-nowrap text-xs">
                    {{ row.status_code ?? '-' }}
                  </td>
                  <td class="ops-request-details-modal__table-cell">
                    <div v-if="row.request_id" class="flex items-center gap-2">
                      <span class="ops-request-details-modal__request-id ops-request-details-modal__text-strong truncate font-mono text-[11px]" :title="row.request_id">
                        {{ row.request_id }}
                      </span>
                      <button
                        class="ops-request-details-modal__copy-button text-[10px] font-bold"
                        @click="handleCopyRequestId(row.request_id)"
                      >
                        {{ t('admin.ops.requestDetails.copy') }}
                      </button>
                    </div>
                    <span v-else class="ops-request-details-modal__text-soft text-xs">-</span>
                  </td>
                  <td class="ops-request-details-modal__table-cell ops-request-details-modal__table-cell--actions whitespace-nowrap text-right">
                    <button
                      v-if="row.kind === 'error' && row.error_id"
                      class="ops-request-details-modal__error-button text-xs font-bold"
                      @click="openErrorDetail(row.error_id)"
                    >
                      {{ t('admin.ops.requestDetails.viewError') }}
                    </button>
                    <span v-else class="ops-request-details-modal__text-soft text-xs">-</span>
                  </td>
                </tr>
              </tbody>
            </table>
            </div>

            <Pagination
              :total="total"
              :page="page"
              :page-size="pageSize"
              @update:page="handlePageChange"
              @update:pageSize="handlePageSizeChange"
            />
          </div>
        </div>
      </div>
    </template>
  </BaseDialog>
</template>

<style scoped>
.ops-request-details-modal__subtitle {
  color: var(--theme-page-muted);
}

.ops-request-details-modal__loading {
  padding-block: calc(var(--theme-ops-card-padding) * 2);
}

.ops-request-details-modal__loading-stack {
  gap: calc(var(--theme-ops-panel-padding) * 0.75);
}

.ops-request-details-modal__spinner {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.ops-request-details-modal__text-strong {
  color: var(--theme-page-text);
}

.ops-request-details-modal__text-body {
  color: color-mix(in srgb, var(--theme-page-text) 80%, var(--theme-page-muted));
}

.ops-request-details-modal__text-soft {
  color: color-mix(in srgb, var(--theme-page-muted) 76%, transparent);
}

.ops-request-details-modal__empty {
  padding: calc(var(--theme-ops-card-padding) * 1.6);
  border-radius: var(--theme-select-panel-radius);
  border: 1px dashed color-mix(in srgb, var(--theme-card-border) 78%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 72%, var(--theme-surface));
}

.ops-request-details-modal__table {
  min-width: var(--theme-ops-table-min-width);
}

.ops-request-details-modal__table-shell {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  border-radius: var(--theme-select-panel-radius);
  background: var(--theme-surface);
}

.ops-request-details-modal__table-head {
  background: var(--theme-table-head-bg);
}

.ops-request-details-modal__table-header {
  padding:
    var(--theme-ops-table-cell-padding-y)
    var(--theme-ops-table-cell-padding-x);
  font-size: var(--theme-table-head-font-size);
  font-weight: 700;
  letter-spacing: var(--theme-table-head-letter-spacing);
  text-transform: var(--theme-table-head-text-transform);
  color: var(--theme-table-head-text);
}

.ops-request-details-modal__table-cell {
  padding:
    var(--theme-ops-table-cell-padding-y)
    var(--theme-ops-table-cell-padding-x);
}

.ops-request-details-modal__table-cell--model {
  max-width: calc(var(--theme-ops-table-min-width) * 0.3);
}

.ops-request-details-modal__table-cell--actions,
.ops-request-details-modal__table-header--actions {
  min-width: fit-content;
}

.ops-request-details-modal__request-id {
  display: inline-block;
  max-width: calc(var(--theme-ops-table-min-width) * 0.275);
}

.ops-request-details-modal__table-row td {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 64%, transparent);
}

.ops-request-details-modal__table-body tr:first-child td {
  border-top: none;
}

.ops-request-details-modal__table-row:hover {
  background: color-mix(in srgb, var(--theme-table-row-hover) 100%, var(--theme-surface));
}

.ops-request-details-modal__copy-button {
  padding: calc(var(--theme-button-padding-y) * 0.4) calc(var(--theme-button-padding-x) * 0.45);
  border-radius: calc(var(--theme-button-radius) * 0.8);
  background: color-mix(in srgb, var(--theme-surface-soft) 82%, var(--theme-surface));
  color: color-mix(in srgb, var(--theme-page-text) 76%, var(--theme-page-muted));
}

.ops-request-details-modal__copy-button:hover {
  background: color-mix(in srgb, var(--theme-button-secondary-hover-bg) 90%, var(--theme-surface));
}

.ops-request-details-modal__error-button {
  padding: calc(var(--theme-button-padding-y) * 0.55) calc(var(--theme-button-padding-x) * 0.65);
  border-radius: var(--theme-button-radius);
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.ops-request-details-modal__error-button:hover {
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 16%, var(--theme-surface));
}
</style>
