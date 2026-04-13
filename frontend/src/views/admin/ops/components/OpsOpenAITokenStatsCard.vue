<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import {
  opsAPI,
  type OpsOpenAITokenStatsParams,
  type OpsOpenAITokenStatsResponse,
  type OpsOpenAITokenStatsTimeRange
} from '@/api/admin/ops'
import { formatNumber } from '@/utils/format'
import { resolveErrorMessage } from '@/utils/errorMessage'

interface Props {
  platformFilter?: string
  groupIdFilter?: number | null
  refreshToken: number
}

type ViewMode = 'topn' | 'pagination'

const props = withDefaults(defineProps<Props>(), {
  platformFilter: '',
  groupIdFilter: null
})

const { t } = useI18n()

const loading = ref(false)
const errorMessage = ref('')
const response = ref<OpsOpenAITokenStatsResponse | null>(null)

const timeRange = ref<OpsOpenAITokenStatsTimeRange>('30d')
const viewMode = ref<ViewMode>('topn')
const topN = ref<number>(20)
const page = ref<number>(1)
const pageSize = ref<number>(20)

const items = computed(() => response.value?.items ?? [])
const total = computed(() => response.value?.total ?? 0)
const totalPages = computed(() => {
  if (viewMode.value !== 'pagination') return 1
  const size = pageSize.value > 0 ? pageSize.value : 20
  return Math.max(1, Math.ceil(total.value / size))
})

const timeRangeOptions = computed(() => [
  { value: '30m', label: t('admin.ops.timeRange.30m') },
  { value: '1h', label: t('admin.ops.timeRange.1h') },
  { value: '1d', label: t('admin.ops.timeRange.1d') },
  { value: '15d', label: t('admin.ops.timeRange.15d') },
  { value: '30d', label: t('admin.ops.timeRange.30d') }
])

const viewModeOptions = computed(() => [
  { value: 'topn', label: t('admin.ops.openaiTokenStats.viewModeTopN') },
  { value: 'pagination', label: t('admin.ops.openaiTokenStats.viewModePagination') }
])

const topNOptions = computed(() => [
  { value: 10, label: 'Top 10' },
  { value: 20, label: 'Top 20' },
  { value: 50, label: 'Top 50' },
  { value: 100, label: 'Top 100' }
])

const pageSizeOptions = computed(() => [
  { value: 10, label: '10' },
  { value: 20, label: '20' },
  { value: 50, label: '50' },
  { value: 100, label: '100' }
])

function getErrorMessage(error: unknown, fallback: string): string {
  return resolveErrorMessage(error, fallback)
}

function formatRate(v?: number | null): string {
  if (typeof v !== 'number' || !Number.isFinite(v)) return '-'
  return v.toFixed(2)
}

function formatInt(v?: number | null): string {
  if (typeof v !== 'number' || !Number.isFinite(v)) return '-'
  return formatNumber(Math.round(v))
}

function buildParams(): OpsOpenAITokenStatsParams {
  const params: OpsOpenAITokenStatsParams = {
    time_range: timeRange.value,
    platform: props.platformFilter || undefined,
    group_id: typeof props.groupIdFilter === 'number' && props.groupIdFilter > 0 ? props.groupIdFilter : undefined
  }

  if (viewMode.value === 'topn') {
    params.top_n = topN.value
  } else {
    params.page = page.value
    params.page_size = pageSize.value
  }
  return params
}

async function loadData() {
  loading.value = true
  errorMessage.value = ''
  try {
    response.value = await opsAPI.getOpenAITokenStats(buildParams())
    // 防御：若 total 变化导致当前页超出最大页，则回退到末页并重新拉取一次。
    if (viewMode.value === 'pagination' && page.value > totalPages.value) {
      page.value = totalPages.value
      response.value = await opsAPI.getOpenAITokenStats(buildParams())
    }
  } catch (error) {
    console.error('[OpsOpenAITokenStatsCard] Failed to load data', error)
    response.value = null
    errorMessage.value = getErrorMessage(error, t('admin.ops.openaiTokenStats.failedToLoad'))
  } finally {
    loading.value = false
  }
}

watch(
  () => ({
    timeRange: timeRange.value,
    viewMode: viewMode.value,
    topN: topN.value,
    page: page.value,
    pageSize: pageSize.value,
    platform: props.platformFilter,
    groupId: props.groupIdFilter,
    refreshToken: props.refreshToken
  }),
  (next, prev) => {
    // 避免“筛选变化 -> 重置页码 -> 触发两次请求”：
    // 先只重置页码，等待下一次 watch（仅 page 变化）再发起请求。
    const filtersChanged = !prev ||
      next.timeRange !== prev.timeRange ||
      next.viewMode !== prev.viewMode ||
      next.pageSize !== prev.pageSize ||
      next.platform !== prev.platform ||
      next.groupId !== prev.groupId

    if (next.viewMode === 'pagination' && filtersChanged && next.page !== 1) {
      page.value = 1
      return
    }

    void loadData()
  },
  { immediate: true }
)

function onPrevPage() {
  if (viewMode.value !== 'pagination') return
  if (page.value > 1) page.value -= 1
}

function onNextPage() {
  if (viewMode.value !== 'pagination') return
  if (page.value < totalPages.value) page.value += 1
}
</script>

<template>
  <section class="card ops-openai-token-stats-card">
    <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
      <h3 class="ops-openai-token-stats-card__title text-sm font-bold">
        {{ t('admin.ops.openaiTokenStats.title') }}
      </h3>
      <div class="flex flex-wrap items-center gap-2">
        <div class="w-36">
          <Select v-model="timeRange" :options="timeRangeOptions" />
        </div>
        <div class="w-36">
          <Select v-model="viewMode" :options="viewModeOptions" />
        </div>
        <div v-if="viewMode === 'topn'" class="w-28">
          <Select v-model="topN" :options="topNOptions" />
        </div>
        <template v-else>
          <div class="w-24">
            <Select v-model="pageSize" :options="pageSizeOptions" />
          </div>
          <button
            class="btn btn-secondary btn-sm"
            :disabled="loading || page <= 1"
            @click="onPrevPage"
          >
            {{ t('admin.ops.openaiTokenStats.prevPage') }}
          </button>
          <button
            class="btn btn-secondary btn-sm"
            :disabled="loading || page >= totalPages"
            @click="onNextPage"
          >
            {{ t('admin.ops.openaiTokenStats.nextPage') }}
          </button>
          <span class="ops-openai-token-stats-card__muted text-xs">
            {{ t('admin.ops.openaiTokenStats.pageInfo', { page, total: totalPages }) }}
          </span>
        </template>
      </div>
    </div>

    <div v-if="errorMessage" class="ops-openai-token-stats-card__error mb-4 text-xs">
      {{ errorMessage }}
    </div>

    <div v-if="loading" class="ops-openai-token-stats-card__loading ops-openai-token-stats-card__muted text-center text-sm">
      {{ t('admin.ops.loadingText') }}
    </div>

    <EmptyState
      v-else-if="items.length === 0"
      :title="t('common.noData')"
      :description="t('admin.ops.openaiTokenStats.empty')"
    />

    <div v-else class="space-y-3">
      <div class="ops-openai-token-stats-card__table-wrap overflow-hidden border">
        <div class="ops-openai-token-stats-card__table-scroll overflow-auto">
          <table class="ops-openai-token-stats-card__table min-w-full text-left text-xs md:text-sm">
            <thead class="ops-openai-token-stats-card__thead sticky top-0 z-10">
              <tr class="ops-openai-token-stats-card__thead-row border-b">
                <th class="ops-openai-token-stats-card__cell ops-openai-token-stats-card__cell--head font-semibold">{{ t('admin.ops.openaiTokenStats.table.model') }}</th>
                <th class="ops-openai-token-stats-card__cell ops-openai-token-stats-card__cell--head font-semibold">{{ t('admin.ops.openaiTokenStats.table.requestCount') }}</th>
                <th class="ops-openai-token-stats-card__cell ops-openai-token-stats-card__cell--head font-semibold">{{ t('admin.ops.openaiTokenStats.table.avgTokensPerSec') }}</th>
                <th class="ops-openai-token-stats-card__cell ops-openai-token-stats-card__cell--head font-semibold">{{ t('admin.ops.openaiTokenStats.table.avgFirstTokenMs') }}</th>
                <th class="ops-openai-token-stats-card__cell ops-openai-token-stats-card__cell--head font-semibold">{{ t('admin.ops.openaiTokenStats.table.totalOutputTokens') }}</th>
                <th class="ops-openai-token-stats-card__cell ops-openai-token-stats-card__cell--head font-semibold">{{ t('admin.ops.openaiTokenStats.table.avgDurationMs') }}</th>
                <th class="ops-openai-token-stats-card__cell ops-openai-token-stats-card__cell--head font-semibold">{{ t('admin.ops.openaiTokenStats.table.requestsWithFirstToken') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="row in items"
                :key="row.model"
                class="ops-openai-token-stats-card__row border-b last:border-b-0"
              >
                <td class="ops-openai-token-stats-card__cell ops-openai-token-stats-card__model font-medium">{{ row.model }}</td>
                <td class="ops-openai-token-stats-card__cell">{{ formatInt(row.request_count) }}</td>
                <td class="ops-openai-token-stats-card__cell">{{ formatRate(row.avg_tokens_per_sec) }}</td>
                <td class="ops-openai-token-stats-card__cell">{{ formatRate(row.avg_first_token_ms) }}</td>
                <td class="ops-openai-token-stats-card__cell">{{ formatInt(row.total_output_tokens) }}</td>
                <td class="ops-openai-token-stats-card__cell">{{ formatInt(row.avg_duration_ms) }}</td>
                <td class="ops-openai-token-stats-card__cell">{{ formatInt(row.requests_with_first_token) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <div v-if="viewMode === 'topn'" class="ops-openai-token-stats-card__muted mt-3 text-xs">
        {{ t('admin.ops.openaiTokenStats.totalModels', { total }) }}
      </div>
    </div>
  </section>
</template>

<style scoped>
.ops-openai-token-stats-card {
  padding: var(--theme-ops-card-padding);
}

.ops-openai-token-stats-card__title,
.ops-openai-token-stats-card__model,
.ops-openai-token-stats-card__row {
  color: var(--theme-page-text);
}

.ops-openai-token-stats-card__muted,
.ops-openai-token-stats-card__thead-row {
  color: var(--theme-page-muted);
}

.ops-openai-token-stats-card__error {
  padding: calc(var(--theme-ops-panel-padding) * 0.75);
  border-radius: var(--theme-button-radius);
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.ops-openai-token-stats-card__loading {
  padding-block: calc(var(--theme-ops-card-padding) * 1.25);
}

.ops-openai-token-stats-card__table-wrap {
  border-color: color-mix(in srgb, var(--theme-card-border) 88%, transparent);
  border-radius: var(--theme-select-panel-radius);
  background: var(--theme-surface);
}

.ops-openai-token-stats-card__table-scroll {
  max-height: calc(var(--theme-ops-table-max-height) * 0.7);
}

.ops-openai-token-stats-card__table {
  min-width: var(--theme-ops-table-min-width);
}

.ops-openai-token-stats-card__thead {
  background: var(--theme-table-head-bg);
}

.ops-openai-token-stats-card__thead-row {
  border-color: color-mix(in srgb, var(--theme-card-border) 76%, transparent);
}

.ops-openai-token-stats-card__cell {
  padding:
    var(--theme-ops-table-cell-padding-compact-y)
    var(--theme-ops-table-cell-padding-compact-x);
}

.ops-openai-token-stats-card__cell--head {
  color: var(--theme-table-head-text);
}

.ops-openai-token-stats-card__row {
  border-color: color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}
</style>
