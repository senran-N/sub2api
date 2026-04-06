<template>
  <div class="ops-error-log-table flex h-full min-h-0 flex-col">
    <!-- Loading State -->
    <div v-if="loading" class="ops-error-log-table__loading flex flex-1 items-center justify-center">
      <div class="ops-error-log-table__spinner h-8 w-8 animate-spin rounded-full border-b-2"></div>
    </div>

    <!-- Table Container -->
    <div v-else class="flex min-h-0 flex-1 flex-col">
      <div class="ops-error-log-table__scroll min-h-0 flex-1 overflow-auto">
        <table class="ops-error-log-table__table w-full border-separate border-spacing-0">
          <thead class="ops-error-log-table__head sticky top-0 z-10">
            <tr>
              <th class="ops-error-log-table__header text-left">
                {{ t('admin.ops.errorLog.time') }}
              </th>
              <th class="ops-error-log-table__header text-left">
                {{ t('admin.ops.errorLog.type') }}
              </th>
              <th class="ops-error-log-table__header text-left">
                {{ t('admin.ops.errorLog.endpoint') }}
              </th>
              <th class="ops-error-log-table__header text-left">
                {{ t('admin.ops.errorLog.platform') }}
              </th>
              <th class="ops-error-log-table__header text-left">
                {{ t('admin.ops.errorLog.model') }}
              </th>
              <th class="ops-error-log-table__header text-left">
                {{ t('admin.ops.errorLog.group') }}
              </th>
              <th class="ops-error-log-table__header text-left">
                {{ t('admin.ops.errorLog.user') }}
              </th>
              <th class="ops-error-log-table__header text-left">
                {{ t('admin.ops.errorLog.status') }}
              </th>
              <th class="ops-error-log-table__header text-left">
                {{ t('admin.ops.errorLog.message') }}
              </th>
              <th class="ops-error-log-table__header ops-error-log-table__header--action text-right">
                {{ t('admin.ops.errorLog.action') }}
              </th>
            </tr>
          </thead>
          <tbody class="ops-error-log-table__body">
            <tr v-if="rows.length === 0">
              <td colspan="10" class="ops-error-log-table__empty ops-error-log-table__table-cell text-center text-sm">
                {{ t('admin.ops.errorLog.noErrors') }}
              </td>
            </tr>

            <tr
              v-for="log in rows"
              :key="log.id"
              class="ops-error-log-table__row group cursor-pointer transition-colors"
              @click="emit('openErrorDetail', log.id)"
            >
              <!-- Time -->
              <td class="ops-error-log-table__table-cell whitespace-nowrap">
                <el-tooltip :content="log.request_id || log.client_request_id" placement="top" :show-after="500">
                  <span class="ops-error-log-table__text-strong font-mono text-xs font-medium">
                    {{ formatDateTime(log.created_at).split(' ')[1] }}
                  </span>
                </el-tooltip>
              </td>

              <!-- Type -->
              <td class="ops-error-log-table__table-cell whitespace-nowrap">
                <span
                  :class="[
                    'ops-error-log-table__inline-chip inline-flex items-center text-[10px] font-bold ring-1 ring-inset',
                    getTypeBadge(log).className
                  ]"
                >
                  {{ getTypeBadge(log).label }}
                </span>
              </td>

              <!-- Endpoint -->
              <td class="ops-error-log-table__table-cell">
                <div class="ops-error-log-table__cell-box ops-error-log-table__cell-box--endpoint">
                  <el-tooltip v-if="log.inbound_endpoint" :content="formatEndpointTooltip(log)" placement="top" :show-after="500">
                    <span class="ops-error-log-table__text-body truncate font-mono text-[11px]">
                      {{ log.inbound_endpoint }}
                    </span>
                  </el-tooltip>
                  <span v-else class="ops-error-log-table__text-soft text-xs">-</span>
                </div>
              </td>

              <!-- Platform -->
              <td class="ops-error-log-table__table-cell whitespace-nowrap">
                <span class="theme-chip theme-chip--compact theme-chip--neutral inline-flex items-center uppercase">
                  {{ log.platform || '-' }}
                </span>
              </td>

              <!-- Model -->
              <td class="ops-error-log-table__table-cell">
                <div class="ops-error-log-table__cell-box ops-error-log-table__cell-box--model">
                  <template v-if="hasModelMapping(log)">
                    <el-tooltip :content="modelMappingTooltip(log)" placement="top" :show-after="500">
                      <span class="ops-error-log-table__text-body flex items-center gap-1 truncate font-mono text-[11px]">
                        <span class="truncate">{{ log.requested_model }}</span>
                        <span class="ops-error-log-table__text-soft flex-shrink-0">→</span>
                        <span class="ops-error-log-table__text-accent truncate">{{ log.upstream_model }}</span>
                      </span>
                    </el-tooltip>
                  </template>
                  <template v-else>
                    <span v-if="displayModel(log)" class="ops-error-log-table__text-body truncate font-mono text-[11px]" :title="displayModel(log)">
                      {{ displayModel(log) }}
                    </span>
                    <span v-else class="ops-error-log-table__text-soft text-xs">-</span>
                  </template>
                </div>
              </td>

              <!-- Group -->
              <td class="ops-error-log-table__table-cell">
                 <el-tooltip v-if="log.group_id" :content="t('admin.ops.errorLog.id') + ' ' + log.group_id" placement="top" :show-after="500">
                  <span class="ops-error-log-table__text-strong ops-error-log-table__compact-label truncate text-xs font-medium">
                    {{ log.group_name || '-' }}
                  </span>
                </el-tooltip>
                <span v-else class="ops-error-log-table__text-soft text-xs">-</span>
              </td>

              <!-- User / Account -->
              <td class="ops-error-log-table__table-cell">
                <template v-if="isUpstreamRow(log)">
                  <el-tooltip v-if="log.account_id" :content="t('admin.ops.errorLog.accountId') + ' ' + log.account_id" placement="top" :show-after="500">
                    <span class="ops-error-log-table__text-strong ops-error-log-table__compact-label truncate text-xs font-medium">
                      {{ log.account_name || '-' }}
                    </span>
                  </el-tooltip>
                  <span v-else class="ops-error-log-table__text-soft text-xs">-</span>
                </template>
                <template v-else>
                  <el-tooltip v-if="log.user_id" :content="t('admin.ops.errorLog.userId') + ' ' + log.user_id" placement="top" :show-after="500">
                    <span class="ops-error-log-table__text-strong ops-error-log-table__compact-label truncate text-xs font-medium">
                      {{ log.user_email || '-' }}
                    </span>
                  </el-tooltip>
                  <span v-else class="ops-error-log-table__text-soft text-xs">-</span>
                </template>
              </td>

              <!-- Status -->
              <td class="ops-error-log-table__table-cell whitespace-nowrap">
                <div class="flex items-center gap-1.5">
                  <span
                    :class="[
                      'ops-error-log-table__inline-chip inline-flex items-center text-[10px] font-bold ring-1 ring-inset',
                      getStatusClass(log.status_code)
                    ]"
                  >
                    {{ log.status_code }}
                  </span>
                  <span
                    v-if="log.severity"
                    :class="getSeverityClass(log.severity)"
                  >
                    {{ log.severity }}
                  </span>
                  <span
                    v-if="log.request_type != null && log.request_type > 0"
                    class="theme-chip theme-chip--compact theme-chip--neutral"
                  >
                    {{ formatRequestType(log.request_type) }}
                  </span>
                </div>
              </td>

              <!-- Message (Response Content) -->
              <td class="ops-error-log-table__table-cell">
                <div class="ops-error-log-table__cell-box ops-error-log-table__cell-box--message">
                  <p class="ops-error-log-table__text-muted truncate text-[11px] font-medium" :title="log.message">
                    {{ formatSmartMessage(log.message) || '-' }}
                  </p>
                </div>
              </td>

              <!-- Actions -->
              <td class="ops-error-log-table__table-cell ops-error-log-table__header--action whitespace-nowrap text-right" @click.stop>
                <div class="flex items-center justify-end gap-3">
                  <button type="button" class="ops-error-log-table__details text-xs font-bold" @click="emit('openErrorDetail', log.id)">
                    {{ t('admin.ops.errorLog.details') }}
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div class="ops-error-log-table__pagination">
        <Pagination
          v-if="total > 0"
          :total="total"
          :page="page"
          :page-size="pageSize"
          :page-size-options="[10]"
          @update:page="emit('update:page', $event)"
          @update:pageSize="emit('update:pageSize', $event)"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Pagination from '@/components/common/Pagination.vue'
import type { OpsErrorLog } from '@/api/admin/ops'
import { getSeverityClass, formatDateTime } from '../utils/opsFormatters'

const { t } = useI18n()

function isUpstreamRow(log: OpsErrorLog): boolean {
  const phase = String(log.phase || '').toLowerCase()
  const owner = String(log.error_owner || '').toLowerCase()
  return phase === 'upstream' && owner === 'provider'
}

function formatEndpointTooltip(log: OpsErrorLog): string {
  const parts: string[] = []
  if (log.inbound_endpoint) parts.push(`Inbound: ${log.inbound_endpoint}`)
  if (log.upstream_endpoint) parts.push(`Upstream: ${log.upstream_endpoint}`)
  return parts.join('\n') || ''
}

function hasModelMapping(log: OpsErrorLog): boolean {
  const requested = String(log.requested_model || '').trim()
  const upstream = String(log.upstream_model || '').trim()
  return !!requested && !!upstream && requested !== upstream
}

function modelMappingTooltip(log: OpsErrorLog): string {
  const requested = String(log.requested_model || '').trim()
  const upstream = String(log.upstream_model || '').trim()
  if (!requested && !upstream) return ''
  if (requested && upstream) return `${requested} → ${upstream}`
  return upstream || requested
}

function displayModel(log: OpsErrorLog): string {
  const upstream = String(log.upstream_model || '').trim()
  if (upstream) return upstream
  const requested = String(log.requested_model || '').trim()
  if (requested) return requested
  return String(log.model || '').trim()
}

function formatRequestType(type: number | null | undefined): string {
  switch (type) {
    case 1: return t('admin.ops.errorLog.requestTypeSync')
    case 2: return t('admin.ops.errorLog.requestTypeStream')
    case 3: return t('admin.ops.errorLog.requestTypeWs')
    default: return ''
  }
}

function getTypeBadge(log: OpsErrorLog): { label: string; className: string } {
  const phase = String(log.phase || '').toLowerCase()
  const owner = String(log.error_owner || '').toLowerCase()

  if (isUpstreamRow(log)) {
    return { label: t('admin.ops.errorLog.typeUpstream'), className: 'theme-chip theme-chip--compact theme-chip--danger' }
  }
  if (phase === 'request' && owner === 'client') {
    return { label: t('admin.ops.errorLog.typeRequest'), className: 'theme-chip theme-chip--compact theme-chip--warning' }
  }
  if (phase === 'auth' && owner === 'client') {
    return { label: t('admin.ops.errorLog.typeAuth'), className: 'theme-chip theme-chip--compact theme-chip--info' }
  }
  if (phase === 'routing' && owner === 'platform') {
    return { label: t('admin.ops.errorLog.typeRouting'), className: 'theme-chip theme-chip--compact theme-chip--brand-purple' }
  }
  if (phase === 'internal' && owner === 'platform') {
    return { label: t('admin.ops.errorLog.typeInternal'), className: 'theme-chip theme-chip--compact theme-chip--neutral' }
  }

  const fallback = phase || owner || t('common.unknown')
  return { label: fallback, className: 'theme-chip theme-chip--compact theme-chip--neutral' }
}

interface Props {
  rows: OpsErrorLog[]
  total: number
  loading: boolean
  page: number
  pageSize: number
}

interface Emits {
  (e: 'openErrorDetail', id: number): void
  (e: 'update:page', value: number): void
  (e: 'update:pageSize', value: number): void
}

defineProps<Props>()
const emit = defineEmits<Emits>()

function getStatusClass(code: number): string {
  if (code >= 500) return 'theme-chip theme-chip--compact theme-chip--danger'
  if (code === 429) return 'theme-chip theme-chip--compact theme-chip--brand-purple'
  if (code >= 400) return 'theme-chip theme-chip--compact theme-chip--warning'
  return 'theme-chip theme-chip--compact theme-chip--neutral'
}

function formatSmartMessage(msg: string): string {
  if (!msg) return ''

  if (msg.startsWith('{') || msg.startsWith('[')) {
    try {
      const obj = JSON.parse(msg)
      if (obj?.error?.message) return String(obj.error.message)
      if (obj?.message) return String(obj.message)
      if (obj?.detail) return String(obj.detail)
      if (typeof obj === 'object') return JSON.stringify(obj).substring(0, 150)
    } catch {
      // ignore parse error
    }
  }

  if (msg.includes('context deadline exceeded')) return t('admin.ops.errorLog.commonErrors.contextDeadlineExceeded')
  if (msg.includes('connection refused')) return t('admin.ops.errorLog.commonErrors.connectionRefused')
  if (msg.toLowerCase().includes('rate limit')) return t('admin.ops.errorLog.commonErrors.rateLimit')

  return msg.length > 200 ? msg.substring(0, 200) + '...' : msg

}
</script>

<style scoped>
.ops-error-log-table {
  background: var(--theme-surface);
}

.ops-error-log-table__loading {
  padding-block: calc(var(--theme-ops-card-padding) * 1.5);
}

.ops-error-log-table__spinner {
  border-bottom-color: var(--theme-accent);
}

.ops-error-log-table__scroll {
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 70%, transparent);
}

.ops-error-log-table__table {
  min-width: var(--theme-ops-table-min-width);
}

.ops-error-log-table__head {
  background: var(--theme-table-head-bg);
}

.ops-error-log-table__header {
  padding:
    calc(var(--theme-ops-table-cell-padding-y) * 0.9)
    var(--theme-ops-table-cell-padding-x);
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 70%, transparent);
  font-size: var(--theme-table-head-font-size);
  font-weight: 700;
  letter-spacing: var(--theme-table-head-letter-spacing);
  text-transform: var(--theme-table-head-text-transform);
  color: var(--theme-table-head-text);
}

.ops-error-log-table__header--action {
  min-width: fit-content;
}

.ops-error-log-table__table-cell {
  padding:
    calc(var(--theme-ops-table-cell-padding-y) * 0.8)
    var(--theme-ops-table-cell-padding-x);
}

.ops-error-log-table__body tr + tr td {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 62%, transparent);
}

.ops-error-log-table__row:hover {
  background: color-mix(in srgb, var(--theme-table-row-hover) 100%, var(--theme-surface));
}

.ops-error-log-table__text-strong {
  color: var(--theme-page-text);
}

.ops-error-log-table__text-body {
  color: color-mix(in srgb, var(--theme-page-text) 80%, var(--theme-page-muted));
}

.ops-error-log-table__text-muted {
  color: var(--theme-page-muted);
}

.ops-error-log-table__text-soft,
.ops-error-log-table__empty {
  color: color-mix(in srgb, var(--theme-page-muted) 76%, transparent);
}

.ops-error-log-table__empty {
  padding-block: calc(var(--theme-ops-card-padding) * 2);
}

.ops-error-log-table__text-accent,
.ops-error-log-table__details {
  color: color-mix(in srgb, var(--theme-accent) 84%, var(--theme-page-text));
}

.ops-error-log-table__details:hover {
  color: color-mix(in srgb, var(--theme-accent-strong) 20%, var(--theme-accent) 80%);
}

.ops-error-log-table__pagination {
  background: color-mix(in srgb, var(--theme-surface-soft) 56%, var(--theme-surface));
}

.ops-error-log-table__inline-chip {
  padding:
    calc(var(--theme-button-padding-y) * 0.2)
    calc(var(--theme-button-padding-x) * 0.32);
  border-radius: calc(var(--theme-button-radius) * 0.65);
}

.ops-error-log-table__cell-box {
  min-width: 0;
}

.ops-error-log-table__cell-box--endpoint,
.ops-error-log-table__cell-box--model {
  max-width: calc(var(--theme-ops-table-min-width) * 0.2);
}

.ops-error-log-table__cell-box--message {
  max-width: calc(var(--theme-ops-table-min-width) * 0.25);
}

.ops-error-log-table__compact-label {
  display: inline-block;
  max-width: calc(var(--theme-ops-table-min-width) * 0.125);
}
</style>
