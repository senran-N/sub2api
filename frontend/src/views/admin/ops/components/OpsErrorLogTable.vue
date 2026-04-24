<template>
  <div class="ops-error-log-table">
    <!-- Loading State -->
    <div v-if="loading" class="ops-error-log-table__loading">
      <div class="ops-error-log-table__spinner animate-spin"></div>
    </div>

    <!-- Table Container -->
    <div v-else class="ops-error-log-table__body">
      <div class="ops-error-log-table__scroll">
        <table class="ops-error-log-table__table">
          <thead class="ops-error-log-table__head">
            <tr>
              <th class="ops-error-log-table__header">
                {{ t('admin.ops.errorLog.time') }}
              </th>
              <th class="ops-error-log-table__header">
                {{ t('admin.ops.errorLog.type') }}
              </th>
              <th class="ops-error-log-table__header">
                {{ t('admin.ops.errorLog.endpoint') }}
              </th>
              <th class="ops-error-log-table__header">
                {{ t('admin.ops.errorLog.platform') }}
              </th>
              <th class="ops-error-log-table__header">
                {{ t('admin.ops.errorLog.model') }}
              </th>
              <th class="ops-error-log-table__header">
                {{ t('admin.ops.errorLog.group') }}
              </th>
              <th class="ops-error-log-table__header">
                {{ t('admin.ops.errorLog.user') }}
              </th>
              <th class="ops-error-log-table__header">
                {{ t('admin.ops.errorLog.status') }}
              </th>
              <th class="ops-error-log-table__header">
                {{ t('admin.ops.errorLog.message') }}
              </th>
              <th class="ops-error-log-table__header ops-error-log-table__header--action">
                {{ t('admin.ops.errorLog.action') }}
              </th>
            </tr>
          </thead>
          <tbody class="ops-error-log-table__body">
            <tr v-if="rows.length === 0">
              <td colspan="10" class="ops-error-log-table__empty ops-error-log-table__table-cell">
                {{ t('admin.ops.errorLog.noErrors') }}
              </td>
            </tr>

            <tr
              v-for="log in rows"
              :key="log.id"
              class="ops-error-log-table__row"
              @click="emit('openErrorDetail', log.id)"
            >
              <!-- Time -->
              <td class="ops-error-log-table__table-cell ops-error-log-table__table-cell--nowrap">
                <el-tooltip :content="log.request_id || log.client_request_id" placement="top" :show-after="500">
                  <span class="ops-error-log-table__text-strong ops-error-log-table__text-strong--mono ops-error-log-table__text-strong--compact">
                    {{ formatDateTime(log.created_at).split(' ')[1] }}
                  </span>
                </el-tooltip>
              </td>

              <!-- Type -->
              <td class="ops-error-log-table__table-cell ops-error-log-table__table-cell--nowrap">
                <span
                  :class="[
                    'ops-error-log-table__inline-chip',
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
                    <span class="ops-error-log-table__text-body ops-error-log-table__text-body--mono ops-error-log-table__text-body--compact ops-error-log-table__text-body--truncate">
                      {{ log.inbound_endpoint }}
                    </span>
                  </el-tooltip>
                  <span v-else class="ops-error-log-table__text-soft ops-error-log-table__text-soft--compact">-</span>
                </div>
              </td>

              <!-- Platform -->
              <td class="ops-error-log-table__table-cell ops-error-log-table__table-cell--nowrap">
                <span class="theme-chip theme-chip--compact theme-chip--neutral ops-error-log-table__platform-chip">
                  {{ log.platform || '-' }}
                </span>
              </td>

              <!-- Model -->
              <td class="ops-error-log-table__table-cell">
                <div class="ops-error-log-table__cell-box ops-error-log-table__cell-box--model">
                  <template v-if="hasModelMapping(log)">
                    <el-tooltip :content="modelMappingTooltip(log)" placement="top" :show-after="500">
                      <span class="ops-error-log-table__model-map">
                        <span class="ops-error-log-table__model-segment ops-error-log-table__text-body ops-error-log-table__text-body--mono ops-error-log-table__text-body--compact">{{ log.requested_model }}</span>
                        <span class="ops-error-log-table__text-soft ops-error-log-table__model-arrow">→</span>
                        <span class="ops-error-log-table__text-accent ops-error-log-table__model-segment">{{ log.upstream_model }}</span>
                      </span>
                    </el-tooltip>
                  </template>
                  <template v-else>
                    <span v-if="displayModel(log)" class="ops-error-log-table__text-body ops-error-log-table__text-body--mono ops-error-log-table__text-body--compact ops-error-log-table__text-body--truncate" :title="displayModel(log)">
                      {{ displayModel(log) }}
                    </span>
                    <span v-else class="ops-error-log-table__text-soft ops-error-log-table__text-soft--compact">-</span>
                  </template>
                </div>
              </td>

              <!-- Group -->
              <td class="ops-error-log-table__table-cell">
                 <el-tooltip v-if="log.group_id" :content="t('admin.ops.errorLog.id') + ' ' + log.group_id" placement="top" :show-after="500">
                  <span class="ops-error-log-table__text-strong ops-error-log-table__compact-label ops-error-log-table__compact-label--truncate ops-error-log-table__text-strong--compact">
                    {{ log.group_name || '-' }}
                  </span>
                </el-tooltip>
                <span v-else class="ops-error-log-table__text-soft ops-error-log-table__text-soft--compact">-</span>
              </td>

              <!-- User / Account -->
              <td class="ops-error-log-table__table-cell">
                <template v-if="isUpstreamRow(log)">
                  <el-tooltip v-if="log.account_id" :content="t('admin.ops.errorLog.accountId') + ' ' + log.account_id" placement="top" :show-after="500">
                    <span class="ops-error-log-table__text-strong ops-error-log-table__compact-label ops-error-log-table__compact-label--truncate ops-error-log-table__text-strong--compact">
                      {{ log.account_name || '-' }}
                    </span>
                  </el-tooltip>
                  <span v-else class="ops-error-log-table__text-soft ops-error-log-table__text-soft--compact">-</span>
                </template>
                <template v-else>
                  <el-tooltip v-if="log.user_id" :content="t('admin.ops.errorLog.userId') + ' ' + log.user_id" placement="top" :show-after="500">
                    <span class="ops-error-log-table__text-strong ops-error-log-table__compact-label ops-error-log-table__compact-label--truncate ops-error-log-table__text-strong--compact">
                      {{ log.user_email || '-' }}
                    </span>
                  </el-tooltip>
                  <span v-else class="ops-error-log-table__text-soft ops-error-log-table__text-soft--compact">-</span>
                </template>
              </td>

              <!-- Status -->
              <td class="ops-error-log-table__table-cell ops-error-log-table__table-cell--nowrap">
                <div class="ops-error-log-table__status-group">
                  <span
                    :class="[
                      'ops-error-log-table__inline-chip',
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
                  <p class="ops-error-log-table__text-muted ops-error-log-table__text-muted--compact ops-error-log-table__text-muted--truncate" :title="log.message">
                    {{ formatSmartMessage(log.message) || '-' }}
                  </p>
                </div>
              </td>

              <!-- Actions -->
              <td class="ops-error-log-table__table-cell ops-error-log-table__header--action ops-error-log-table__table-cell--nowrap" @click.stop>
                <div class="ops-error-log-table__action-group">
                  <button type="button" class="ops-error-log-table__details" @click="emit('openErrorDetail', log.id)">
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
  display: flex;
  min-height: 0;
  height: 100%;
  flex-direction: column;
}

.ops-error-log-table__loading {
  display: flex;
  flex: 1 1 auto;
  align-items: center;
  justify-content: center;
  padding-block: calc(var(--theme-ops-card-padding) * 1.5);
}

.ops-error-log-table__spinner {
  width: var(--theme-ops-error-log-loading-size);
  height: var(--theme-ops-error-log-loading-size);
  border-width: 0 0 2px 0;
  border-style: solid;
  border-radius: 9999px;
  border-bottom-color: var(--theme-accent);
}

.ops-error-log-table__body {
  display: flex;
  min-height: 0;
  flex: 1 1 auto;
  flex-direction: column;
}

.ops-error-log-table__scroll {
  min-height: 0;
  flex: 1 1 auto;
  overflow: auto;
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 70%, transparent);
}

.ops-error-log-table__table {
  min-width: var(--theme-ops-table-min-width);
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
}

.ops-error-log-table__head {
  background: var(--theme-table-head-bg);
  position: sticky;
  top: 0;
  z-index: 10;
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
  text-align: left;
}

.ops-error-log-table__header--action {
  min-width: fit-content;
  text-align: right;
}

.ops-error-log-table__table-cell {
  padding:
    calc(var(--theme-ops-table-cell-padding-y) * 0.8)
    var(--theme-ops-table-cell-padding-x);
}

.ops-error-log-table__table-cell--nowrap {
  white-space: nowrap;
}

.ops-error-log-table__body tr + tr td {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 62%, transparent);
}

.ops-error-log-table__row:hover {
  background: color-mix(in srgb, var(--theme-table-row-hover) 100%, var(--theme-surface));
  cursor: pointer;
}

.ops-error-log-table__text-strong {
  color: var(--theme-page-text);
}

.ops-error-log-table__text-strong--mono {
  font-family: var(--theme-font-mono);
}

.ops-error-log-table__text-strong--compact {
  font-size: 0.75rem;
  font-weight: 500;
}

.ops-error-log-table__text-body {
  color: color-mix(in srgb, var(--theme-page-text) 80%, var(--theme-page-muted));
}

.ops-error-log-table__text-body--mono {
  font-family: var(--theme-font-mono);
}

.ops-error-log-table__text-body--compact {
  font-size: var(--theme-ops-error-log-content-size);
}

.ops-error-log-table__text-body--truncate,
.ops-error-log-table__text-muted--truncate,
.ops-error-log-table__compact-label--truncate,
.ops-error-log-table__model-segment {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ops-error-log-table__text-muted {
  color: var(--theme-page-muted);
}

.ops-error-log-table__text-muted--compact {
  font-size: var(--theme-ops-error-log-content-size);
  font-weight: 500;
}

.ops-error-log-table__text-soft,
.ops-error-log-table__empty {
  color: color-mix(in srgb, var(--theme-page-muted) 76%, transparent);
}

.ops-error-log-table__text-soft--compact {
  font-size: 0.75rem;
}

.ops-error-log-table__empty {
  padding-block: calc(var(--theme-ops-card-padding) * 2);
  text-align: center;
  font-size: 0.875rem;
}

.ops-error-log-table__text-accent,
.ops-error-log-table__details {
  color: color-mix(in srgb, var(--theme-accent) 84%, var(--theme-page-text));
}

.ops-error-log-table__details:hover {
  color: color-mix(in srgb, var(--theme-accent-strong) 20%, var(--theme-accent) 80%);
}

.ops-error-log-table__details {
  font-size: 0.75rem;
  font-weight: 700;
}

.ops-error-log-table__pagination {
  background: color-mix(in srgb, var(--theme-surface-soft) 56%, var(--theme-surface));
}

.ops-error-log-table__inline-chip {
  display: inline-flex;
  align-items: center;
  padding:
    calc(var(--theme-button-padding-y) * 0.2)
    calc(var(--theme-button-padding-x) * 0.32);
  border-radius: calc(var(--theme-button-radius) * 0.65);
  font-size: var(--theme-ops-error-log-inline-chip-size);
  font-weight: 700;
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

.ops-error-log-table__platform-chip {
  display: inline-flex;
  align-items: center;
  text-transform: uppercase;
}

.ops-error-log-table__model-map,
.ops-error-log-table__status-group,
.ops-error-log-table__action-group {
  display: flex;
  align-items: center;
}

.ops-error-log-table__model-map,
.ops-error-log-table__status-group {
  gap: var(--theme-ops-error-log-status-gap);
}

.ops-error-log-table__model-map {
  min-width: 0;
}

.ops-error-log-table__model-arrow {
  flex-shrink: 0;
}

.ops-error-log-table__action-group {
  justify-content: flex-end;
  gap: var(--theme-ops-error-log-action-gap);
}
</style>
