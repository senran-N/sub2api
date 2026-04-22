<template>
  <BaseDialog :show="show" :title="title" width="full" :close-on-click-outside="true" @close="close">
    <div v-if="loading" class="ops-error-detail-modal__state ops-error-detail-modal__state--loading">
      <div class="ops-error-detail-modal__stack">
        <div class="ops-error-detail-modal__spinner animate-spin"></div>
        <div class="ops-error-detail-modal__loading-label">{{ t('admin.ops.errorDetail.loading') }}</div>
      </div>
    </div>

    <div v-else-if="!detail" class="ops-error-detail-modal__state ops-error-detail-modal__state--empty">
      {{ emptyText }}
    </div>

    <div v-else class="ops-error-detail-modal__content">
      <div class="ops-error-detail-modal__summary-grid">
        <div class="ops-error-detail-modal__summary-card">
          <div class="ops-error-detail-modal__summary-kicker">{{ t('admin.ops.errorDetail.requestId') }}</div>
          <div class="ops-error-detail-modal__summary-value ops-error-detail-modal__summary-value--meta ops-error-detail-modal__summary-value--break ops-error-detail-modal__mono">
            {{ requestId || '—' }}
          </div>
        </div>

        <div class="ops-error-detail-modal__summary-card">
          <div class="ops-error-detail-modal__summary-kicker">{{ t('admin.ops.errorDetail.time') }}</div>
          <div class="ops-error-detail-modal__summary-value ops-error-detail-modal__summary-value--meta">
            {{ formatDateTime(detail.created_at) }}
          </div>
        </div>

        <div class="ops-error-detail-modal__summary-card">
          <div class="ops-error-detail-modal__summary-kicker">
            {{ isUpstreamError(detail) ? t('admin.ops.errorDetail.account') : t('admin.ops.errorDetail.user') }}
          </div>
          <div class="ops-error-detail-modal__summary-value ops-error-detail-modal__summary-value--meta">
            <template v-if="isUpstreamError(detail)">
              {{ detail.account_name || (detail.account_id != null ? String(detail.account_id) : '—') }}
            </template>
            <template v-else>
              {{ detail.user_email || (detail.user_id != null ? String(detail.user_id) : '—') }}
            </template>
          </div>
        </div>

        <div class="ops-error-detail-modal__summary-card">
          <div class="ops-error-detail-modal__summary-kicker">{{ t('admin.ops.errorDetail.platform') }}</div>
          <div class="ops-error-detail-modal__summary-value ops-error-detail-modal__summary-value--meta">
            {{ detail.platform || '—' }}
          </div>
        </div>

        <div class="ops-error-detail-modal__summary-card">
          <div class="ops-error-detail-modal__summary-kicker">{{ t('admin.ops.errorDetail.group') }}</div>
          <div class="ops-error-detail-modal__summary-value ops-error-detail-modal__summary-value--meta">
            {{ detail.group_name || (detail.group_id != null ? String(detail.group_id) : '—') }}
          </div>
        </div>

        <div class="ops-error-detail-modal__summary-card">
          <div class="ops-error-detail-modal__summary-kicker">{{ t('admin.ops.errorDetail.model') }}</div>
          <div class="ops-error-detail-modal__summary-value ops-error-detail-modal__summary-value--meta">
            <template v-if="hasModelMapping(detail)">
              <span class="ops-error-detail-modal__mono">{{ detail.requested_model }}</span>
              <span class="ops-error-detail-modal__summary-arrow">→</span>
              <span class="ops-error-detail-modal__accent ops-error-detail-modal__mono">{{ detail.upstream_model }}</span>
            </template>
            <template v-else>
              {{ displayModel(detail) || '—' }}
            </template>
          </div>
        </div>

        <div class="ops-error-detail-modal__summary-card">
          <div class="ops-error-detail-modal__summary-kicker">{{ t('admin.ops.errorDetail.inboundEndpoint') }}</div>
          <div class="ops-error-detail-modal__summary-value ops-error-detail-modal__summary-value--meta ops-error-detail-modal__summary-value--break ops-error-detail-modal__mono">
            {{ detail.inbound_endpoint || '—' }}
          </div>
        </div>

        <div class="ops-error-detail-modal__summary-card">
          <div class="ops-error-detail-modal__summary-kicker">{{ t('admin.ops.errorDetail.upstreamEndpoint') }}</div>
          <div class="ops-error-detail-modal__summary-value ops-error-detail-modal__summary-value--meta ops-error-detail-modal__summary-value--break ops-error-detail-modal__mono">
            {{ detail.upstream_endpoint || '—' }}
          </div>
        </div>

        <div class="ops-error-detail-modal__summary-card">
          <div class="ops-error-detail-modal__summary-kicker">{{ t('admin.ops.errorDetail.status') }}</div>
          <div class="ops-error-detail-modal__summary-value ops-error-detail-modal__summary-value--meta">
            <span :class="statusClass">
              {{ detail.status_code }}
            </span>
          </div>
        </div>

        <div class="ops-error-detail-modal__summary-card">
          <div class="ops-error-detail-modal__summary-kicker">{{ t('admin.ops.errorDetail.requestType') }}</div>
          <div class="ops-error-detail-modal__summary-value ops-error-detail-modal__summary-value--meta">
            {{ formatRequestTypeLabel(detail.request_type) }}
          </div>
        </div>

        <div class="ops-error-detail-modal__summary-card">
          <div class="ops-error-detail-modal__summary-kicker">{{ t('admin.ops.errorDetail.message') }}</div>
          <div class="ops-error-detail-modal__summary-value ops-error-detail-modal__summary-value--meta ops-error-detail-modal__summary-value--truncate" :title="detail.message">
            {{ detail.message || '—' }}
          </div>
        </div>
      </div>

      <div v-if="timingEntries.length" class="ops-error-detail-modal__panel">
        <div class="ops-error-detail-modal__title-row">
          <h3 class="ops-error-detail-modal__title">
            {{ t('admin.ops.errorDetail.timings') }}
          </h3>
          <div class="ops-error-detail-modal__subtitle ops-error-detail-modal__subtitle--compact">
            {{ t('admin.ops.errorDetail.timingsHint') }}
          </div>
        </div>

        <div class="ops-error-detail-modal__timing-grid">
          <div
            v-for="entry in timingEntries"
            :key="entry.key"
            class="ops-error-detail-modal__timing-card"
          >
            <div class="ops-error-detail-modal__summary-kicker ops-error-detail-modal__summary-kicker--compact">
              {{ entry.label }}
            </div>
            <div class="ops-error-detail-modal__timing-value">
              {{ entry.value }}ms
            </div>
          </div>
        </div>
      </div>

      <div class="ops-error-detail-modal__panel">
        <h3 class="ops-error-detail-modal__title">{{ t('admin.ops.errorDetail.responseBody') }}</h3>
        <pre class="ops-error-detail-modal__code ops-error-detail-modal__code--primary"><code>{{ prettyJSON(primaryResponseBody || '') }}</code></pre>
      </div>

      <div v-if="showUpstreamList" class="ops-error-detail-modal__panel">
        <div class="ops-error-detail-modal__title-row">
          <h3 class="ops-error-detail-modal__title">{{ t('admin.ops.errorDetails.upstreamErrors') }}</h3>
          <div v-if="correlatedUpstreamLoading" class="ops-error-detail-modal__subtitle ops-error-detail-modal__subtitle--compact">{{ t('common.loading') }}</div>
        </div>

        <div v-if="!correlatedUpstreamLoading && !correlatedUpstreamErrors.length" class="ops-error-detail-modal__subtitle ops-error-detail-modal__subtitle--empty">
          {{ t('common.noData') }}
        </div>

        <div v-else class="ops-error-detail-modal__item-list">
          <div
            v-for="(ev, idx) in correlatedUpstreamErrors"
            :key="ev.id"
            class="ops-error-detail-modal__item"
          >
            <div class="ops-error-detail-modal__item-header">
              <div class="ops-error-detail-modal__item-title">
                #{{ idx + 1 }}
                <span v-if="ev.type" class="ops-error-detail-modal__type-chip theme-chip theme-chip--neutral theme-chip--compact ops-error-detail-modal__mono">{{ ev.type }}</span>
              </div>
              <div class="ops-error-detail-modal__item-actions">
                <div class="ops-error-detail-modal__subtitle ops-error-detail-modal__subtitle--compact ops-error-detail-modal__mono">
                  {{ ev.status_code ?? '—' }}
                </div>
                <button
                  type="button"
                  class="ops-error-detail-modal__toggle"
                  :disabled="!getUpstreamResponsePreview(ev)"
                  :title="getUpstreamResponsePreview(ev) ? '' : t('common.noData')"
                  @click="toggleUpstreamDetail(ev.id)"
                >
                  <Icon
                    :name="expandedUpstreamDetailIds.has(ev.id) ? 'chevronDown' : 'chevronRight'"
                    size="xs"
                    :stroke-width="2"
                  />
                  <span>
                    {{
                      expandedUpstreamDetailIds.has(ev.id)
                        ? t('admin.ops.errorDetail.responsePreview.collapse')
                        : t('admin.ops.errorDetail.responsePreview.expand')
                    }}
                  </span>
                </button>
              </div>
            </div>

            <div class="ops-error-detail-modal__meta-grid">
              <div>
                <span class="ops-error-detail-modal__summary-kicker">{{ t('admin.ops.errorDetail.upstreamEvent.status') }}:</span>
                <span class="ops-error-detail-modal__meta-value ops-error-detail-modal__mono">{{ ev.status_code ?? '—' }}</span>
              </div>
              <div>
                <span class="ops-error-detail-modal__summary-kicker">{{ t('admin.ops.errorDetail.upstreamEvent.requestId') }}:</span>
                <span class="ops-error-detail-modal__meta-value ops-error-detail-modal__mono">{{ ev.request_id || ev.client_request_id || '—' }}</span>
              </div>
            </div>

            <div v-if="ev.message" class="ops-error-detail-modal__summary-value ops-error-detail-modal__summary-value--meta ops-error-detail-modal__summary-value--break">
              {{ ev.message }}
            </div>

            <pre
              v-if="expandedUpstreamDetailIds.has(ev.id)"
              class="ops-error-detail-modal__code ops-error-detail-modal__code--secondary"
            ><code>{{ prettyJSON(getUpstreamResponsePreview(ev)) }}</code></pre>
          </div>
        </div>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import { opsAPI, type OpsErrorDetail } from '@/api/admin/ops'
import { formatDateTime } from '@/utils/format'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import { resolvePrimaryResponseBody, resolveUpstreamPayload } from '../utils/errorDetailResponse'

interface Props {
  show: boolean
  errorId: number | null
  errorType?: 'request' | 'upstream'
}

interface Emits {
  (e: 'update:show', value: boolean): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const detail = ref<OpsErrorDetail | null>(null)

const showUpstreamList = computed(() => props.errorType === 'request')

const requestId = computed(() => detail.value?.request_id || detail.value?.client_request_id || '')

const primaryResponseBody = computed(() => {
  return resolvePrimaryResponseBody(detail.value, props.errorType)
})

type TimingEntry = {
  key: string
  label: string
  value: number
}

const timingEntries = computed<TimingEntry[]>(() => {
  const current = detail.value
  if (!current) {
    return []
  }

  const definitions = [
    { key: 'auth_latency_ms', label: t('admin.ops.errorDetail.auth'), value: current.auth_latency_ms },
    { key: 'routing_latency_ms', label: t('admin.ops.errorDetail.routing'), value: current.routing_latency_ms },
    { key: 'wait_user_ms', label: t('admin.ops.errorDetail.waitUser'), value: current.wait_user_ms },
    { key: 'wait_account_ms', label: t('admin.ops.errorDetail.waitAccount'), value: current.wait_account_ms },
    { key: 'ws_acquire_ms', label: t('admin.ops.errorDetail.wsAcquire'), value: current.ws_acquire_ms },
    { key: 'ws_healthcheck_ms', label: t('admin.ops.errorDetail.wsHealthcheck'), value: current.ws_healthcheck_ms },
    { key: 'upstream_latency_ms', label: t('admin.ops.errorDetail.upstream'), value: current.upstream_latency_ms },
    { key: 'response_latency_ms', label: t('admin.ops.errorDetail.response'), value: current.response_latency_ms },
    { key: 'time_to_first_token_ms', label: t('admin.ops.errorDetail.firstToken'), value: current.time_to_first_token_ms }
  ]

  return definitions.flatMap((definition) => {
    if (typeof definition.value !== 'number' || Number.isNaN(definition.value)) {
      return []
    }

    return [{
      key: definition.key,
      label: definition.label,
      value: definition.value
    }]
  })
})




const title = computed(() => {
  if (!props.errorId) return t('admin.ops.errorDetail.title')
  return t('admin.ops.errorDetail.titleWithId', { id: String(props.errorId) })
})

const emptyText = computed(() => t('admin.ops.errorDetail.noErrorSelected'))

function isUpstreamError(d: OpsErrorDetail | null): boolean {
  if (!d) return false
  const phase = String(d.phase || '').toLowerCase()
  const owner = String(d.error_owner || '').toLowerCase()
  return phase === 'upstream' && owner === 'provider'
}

function formatRequestTypeLabel(type: number | null | undefined): string {
  switch (type) {
    case 1: return t('admin.ops.errorDetail.requestTypeSync')
    case 2: return t('admin.ops.errorDetail.requestTypeStream')
    case 3: return t('admin.ops.errorDetail.requestTypeWs')
    default: return t('admin.ops.errorDetail.requestTypeUnknown')
  }
}

function hasModelMapping(d: OpsErrorDetail | null): boolean {
  if (!d) return false
  const requested = String(d.requested_model || '').trim()
  const upstream = String(d.upstream_model || '').trim()
  return !!requested && !!upstream && requested !== upstream
}

function displayModel(d: OpsErrorDetail | null): string {
  if (!d) return ''
  const upstream = String(d.upstream_model || '').trim()
  if (upstream) return upstream
  const requested = String(d.requested_model || '').trim()
  if (requested) return requested
  return String(d.model || '').trim()
}

const correlatedUpstream = ref<OpsErrorDetail[]>([])
const correlatedUpstreamLoading = ref(false)

const correlatedUpstreamErrors = computed<OpsErrorDetail[]>(() => correlatedUpstream.value)

const expandedUpstreamDetailIds = ref(new Set<number>())
let detailRequestSequence = 0
let correlatedUpstreamSequence = 0

function getUpstreamResponsePreview(ev: OpsErrorDetail): string {
  const upstreamPayload = resolveUpstreamPayload(ev)
  if (upstreamPayload) return upstreamPayload
  return String(ev.error_body || '').trim()
}

function toggleUpstreamDetail(id: number) {
  const next = new Set(expandedUpstreamDetailIds.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  expandedUpstreamDetailIds.value = next
}

async function fetchCorrelatedUpstreamErrors(requestErrorId: number) {
  const requestSequence = ++correlatedUpstreamSequence
  correlatedUpstreamLoading.value = true
  try {
    const res = await opsAPI.listRequestErrorUpstreamErrors(
      requestErrorId,
      { page: 1, page_size: 100, view: 'all' },
      { include_detail: true }
    )
    if (
      requestSequence !== correlatedUpstreamSequence ||
      !props.show ||
      props.errorType !== 'request' ||
      props.errorId !== requestErrorId
    ) {
      return
    }
    correlatedUpstream.value = res.items || []
  } catch (err) {
    if (
      requestSequence !== correlatedUpstreamSequence ||
      !props.show ||
      props.errorType !== 'request' ||
      props.errorId !== requestErrorId
    ) {
      return
    }
    console.error('[OpsErrorDetailModal] Failed to load correlated upstream errors', err)
    correlatedUpstream.value = []
  } finally {
    if (requestSequence === correlatedUpstreamSequence) {
      correlatedUpstreamLoading.value = false
    }
  }
}

function close() {
  emit('update:show', false)
}

function prettyJSON(raw?: string): string {
  if (!raw) return 'N/A'
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
}

async function fetchDetail(id: number) {
  const requestSequence = ++detailRequestSequence
  const detailType = props.errorType === 'upstream' ? 'upstream' : 'request'
  loading.value = true
  detail.value = null
  try {
    const d = detailType === 'upstream' ? await opsAPI.getUpstreamErrorDetail(id) : await opsAPI.getRequestErrorDetail(id)
    if (
      requestSequence !== detailRequestSequence ||
      !props.show ||
      props.errorId !== id ||
      (props.errorType === 'upstream' ? 'upstream' : 'request') !== detailType
    ) {
      return
    }
    detail.value = d
  } catch (err: unknown) {
    if (
      requestSequence !== detailRequestSequence ||
      !props.show ||
      props.errorId !== id ||
      (props.errorType === 'upstream' ? 'upstream' : 'request') !== detailType
    ) {
      return
    }
    detail.value = null
    appStore.showError(resolveRequestErrorMessage(err, t('admin.ops.failedToLoadErrorDetail')))
  } finally {
    if (requestSequence === detailRequestSequence) {
      loading.value = false
    }
  }
}

watch(
  () => [props.show, props.errorId, props.errorType] as const,
  ([show, id, errorType]) => {
    if (!show || typeof id !== 'number' || id <= 0) {
      detailRequestSequence++
      correlatedUpstreamSequence++
      loading.value = false
      correlatedUpstreamLoading.value = false
      detail.value = null
      correlatedUpstream.value = []
      expandedUpstreamDetailIds.value = new Set()
      return
    }

    expandedUpstreamDetailIds.value = new Set()
    correlatedUpstream.value = []
    correlatedUpstreamLoading.value = false
    fetchDetail(id)

    if (errorType === 'request') {
      fetchCorrelatedUpstreamErrors(id)
    } else {
      correlatedUpstreamSequence++
    }
  },
  { immediate: true }
)

const statusClass = computed(() => {
  const code = detail.value?.status_code ?? 0
  if (code >= 500) return 'ops-error-detail-modal__status-chip theme-chip theme-chip--danger theme-chip--regular'
  if (code === 429) return 'ops-error-detail-modal__status-chip theme-chip theme-chip--brand-purple theme-chip--regular'
  if (code >= 400) return 'ops-error-detail-modal__status-chip theme-chip theme-chip--warning theme-chip--regular'
  return 'ops-error-detail-modal__status-chip theme-chip theme-chip--neutral theme-chip--regular'
})

</script>

<style scoped>
.ops-error-detail-modal__state,
.ops-error-detail-modal__subtitle,
.ops-error-detail-modal__summary-kicker {
  color: var(--theme-page-muted);
}

.ops-error-detail-modal__state {
  text-align: center;
}

.ops-error-detail-modal__state--loading {
  padding-block: calc(var(--theme-ops-card-padding) * 2);
}

.ops-error-detail-modal__state--empty {
  padding-block: calc(var(--theme-ops-card-padding) * 1.4);
  font-size: var(--theme-ops-error-detail-text-regular);
}

.ops-error-detail-modal__stack {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: calc(var(--theme-ops-panel-padding) * 0.75);
}

.ops-error-detail-modal__loading-label {
  font-size: var(--theme-ops-error-detail-text-regular);
  font-weight: 500;
}

.ops-error-detail-modal__content {
  padding: calc(var(--theme-ops-card-padding) * 0.95);
  display: flex;
  flex-direction: column;
  gap: var(--theme-ops-error-detail-content-gap);
}

.ops-error-detail-modal__title,
.ops-error-detail-modal__summary-value,
.ops-error-detail-modal__meta-grid {
  color: var(--theme-page-text);
}

.ops-error-detail-modal__spinner {
  width: var(--theme-ops-error-detail-spinner-size);
  height: var(--theme-ops-error-detail-spinner-size);
  border-width: 0 0 2px 0;
  border-style: solid;
  border-radius: 9999px;
  border-color: color-mix(in srgb, var(--theme-page-border) 68%, transparent);
  border-bottom-color: var(--theme-accent);
}

.ops-error-detail-modal__summary-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--theme-ops-error-detail-grid-gap);
}

.ops-error-detail-modal__summary-card,
.ops-error-detail-modal__panel,
.ops-error-detail-modal__item,
.ops-error-detail-modal__timing-card {
  padding: var(--theme-ops-panel-padding);
  border-radius: var(--theme-select-panel-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.ops-error-detail-modal__summary-kicker {
  font-size: var(--theme-ops-error-detail-text-compact);
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.ops-error-detail-modal__summary-kicker--compact {
  font-size: var(--theme-ops-error-detail-text-subtle);
}

.ops-error-detail-modal__summary-value--meta {
  margin-top: 0.25rem;
  font-size: var(--theme-ops-error-detail-text-regular);
  font-weight: 500;
}

.ops-error-detail-modal__summary-value--break {
  overflow-wrap: anywhere;
}

.ops-error-detail-modal__summary-value--truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ops-error-detail-modal__item,
.ops-error-detail-modal__timing-card {
  border-color: color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.ops-error-detail-modal__timing-card {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 58%, transparent);
  background:
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--theme-accent-soft) 52%, var(--theme-surface)) 0%,
      color-mix(in srgb, var(--theme-surface-soft) 92%, var(--theme-surface)) 100%
    );
}

.ops-error-detail-modal__timing-value {
  margin-top: 0.5rem;
  color: color-mix(in srgb, var(--theme-accent) 72%, var(--theme-page-text));
  font-size: var(--theme-ops-error-detail-timing-size);
  font-weight: 600;
}

.ops-error-detail-modal__summary-arrow {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
  margin-inline: 0.25rem;
}

.ops-error-detail-modal__accent,
.ops-error-detail-modal__toggle {
  color: color-mix(in srgb, var(--theme-accent) 84%, var(--theme-page-text));
}

.ops-error-detail-modal__toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: calc(var(--theme-button-padding-y) * 0.4) calc(var(--theme-button-padding-x) * 0.32);
  border-radius: calc(var(--theme-button-radius) * 0.8);
  font-size: var(--theme-ops-error-detail-text-micro);
  font-weight: 700;
}

.ops-error-detail-modal__toggle:hover {
  background: color-mix(in srgb, var(--theme-accent-soft) 86%, var(--theme-surface));
}

.ops-error-detail-modal__toggle:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.ops-error-detail-modal__code {
  border-radius: var(--theme-select-panel-radius);
  border: 1px solid color-mix(in srgb, var(--theme-page-border) 74%, transparent);
  border-color: color-mix(in srgb, var(--theme-page-border) 74%, transparent);
  background: var(--theme-surface);
  color: var(--theme-page-text);
  overflow: auto;
  font-size: var(--theme-ops-error-detail-text-compact);
}

.ops-error-detail-modal__code--primary {
  max-height: calc(var(--theme-ops-table-max-height) * 0.92);
  padding: var(--theme-ops-panel-padding);
  margin-top: var(--theme-ops-error-detail-code-gap);
}

.ops-error-detail-modal__code--secondary {
  max-height: calc(var(--theme-ops-list-max-height) * 0.9);
  padding: calc(var(--theme-ops-panel-padding) * 0.75);
  margin-top: var(--theme-ops-error-detail-item-gap);
}

.ops-error-detail-modal__status-chip,
.ops-error-detail-modal__type-chip {
  display: inline-flex;
  align-items: center;
  padding:
    calc(var(--theme-button-padding-y) * 0.35)
    calc(var(--theme-button-padding-x) * 0.5);
}

.ops-error-detail-modal__status-chip {
  border-radius: var(--theme-button-radius);
  font-size: var(--theme-ops-error-detail-text-compact);
  font-weight: 900;
}

.ops-error-detail-modal__type-chip {
  border-radius: calc(var(--theme-button-radius) * 0.8);
  margin-left: 0.5rem;
  font-size: var(--theme-ops-error-detail-text-micro);
  font-weight: 700;
}

.ops-error-detail-modal__title-row,
.ops-error-detail-modal__item-header,
.ops-error-detail-modal__item-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
}

.ops-error-detail-modal__title-row,
.ops-error-detail-modal__item-header {
  justify-content: space-between;
  gap: var(--theme-ops-error-detail-title-gap);
}

.ops-error-detail-modal__item-actions {
  gap: 0.5rem;
}

.ops-error-detail-modal__title {
  font-size: var(--theme-ops-error-detail-text-regular);
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.ops-error-detail-modal__subtitle--compact {
  font-size: var(--theme-ops-error-detail-text-compact);
}

.ops-error-detail-modal__subtitle--empty {
  margin-top: 0.75rem;
  font-size: var(--theme-ops-error-detail-text-regular);
}

.ops-error-detail-modal__timing-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.75rem;
  margin-top: var(--theme-ops-error-detail-code-gap);
}

.ops-error-detail-modal__item-list {
  display: flex;
  flex-direction: column;
  gap: var(--theme-ops-error-detail-item-gap);
  margin-top: var(--theme-ops-error-detail-code-gap);
}

.ops-error-detail-modal__item {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.ops-error-detail-modal__item-title {
  font-size: var(--theme-ops-error-detail-text-compact);
  font-weight: 900;
}

.ops-error-detail-modal__meta-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--theme-ops-error-detail-meta-gap);
  margin-top: var(--theme-ops-error-detail-item-gap);
  font-size: var(--theme-ops-error-detail-text-compact);
}

.ops-error-detail-modal__meta-value {
  margin-left: 0.25rem;
}

.ops-error-detail-modal__mono {
  font-family: var(--theme-font-mono);
}

@media (min-width: 640px) {
  .ops-error-detail-modal__summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .ops-error-detail-modal__timing-grid,
  .ops-error-detail-modal__meta-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 1024px) {
  .ops-error-detail-modal__summary-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

@media (min-width: 1280px) {
  .ops-error-detail-modal__timing-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
</style>
