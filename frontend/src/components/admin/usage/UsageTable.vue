<template>
  <div class="card overflow-hidden">
    <div class="overflow-auto">
      <DataTable :columns="columns" :data="data" :loading="loading">
        <template #cell-user="{ row }">
          <div class="text-sm">
            <button
              v-if="row.user?.email"
              class="usage-table__user-link font-medium underline decoration-dashed underline-offset-2 transition-colors"
              :title="t('admin.usage.clickToViewBalance')"
              @click="$emit('userClick', row.user_id, row.user?.email)"
            >
              {{ row.user.email }}
            </button>
            <span v-else class="usage-table__text-body font-medium">-</span>
            <span class="usage-table__text-muted ml-1">#{{ row.user_id }}</span>
          </div>
        </template>

        <template #cell-api_key="{ row }">
          <span class="usage-table__text-body text-sm">{{ row.api_key?.name || '-' }}</span>
        </template>

        <template #cell-account="{ row }">
          <span class="usage-table__text-body text-sm">{{ row.account?.name || '-' }}</span>
        </template>

        <template #cell-model="{ row }">
          <div
            v-if="
              row.model_mapping_chain ||
              (row.upstream_model && row.upstream_model !== row.model) ||
              row.channel_id != null
            "
            class="space-y-1 text-xs"
          >
            <div class="usage-table__text-body break-all font-medium">
              {{ row.model }}
            </div>
            <div v-if="row.model_mapping_chain" class="usage-table__text-muted break-all">
              {{ row.model_mapping_chain }}
            </div>
            <div
              v-else-if="row.upstream_model && row.upstream_model !== row.model"
              class="usage-table__text-muted break-all"
            >
              <span class="mr-0.5">↳</span>{{ row.upstream_model }}
            </div>
            <div
              v-if="row.channel_id != null"
              class="theme-chip theme-chip--regular theme-chip--warning inline-flex w-fit"
            >
              <span>{{ t('usage.channel') }}</span>
              <span>#{{ row.channel_id }}</span>
            </div>
          </div>
          <span v-else class="usage-table__text-body font-medium">{{ row.model }}</span>
        </template>

        <template #cell-reasoning_effort="{ row }">
          <span class="usage-table__text-body text-sm">
            {{ formatReasoningEffort(row.reasoning_effort) }}
          </span>
        </template>

        <template #cell-endpoint="{ row }">
          <div class="usage-table__endpoint-cell space-y-1 text-xs">
            <div class="usage-table__text-support break-all">
              <span class="usage-table__text-muted font-medium">{{ t('usage.inbound') }}:</span>
              <span class="ml-1">{{ row.inbound_endpoint?.trim() || '-' }}</span>
            </div>
            <div class="usage-table__text-support break-all">
              <span class="usage-table__text-muted font-medium">{{ t('usage.upstream') }}:</span>
              <span class="ml-1">{{ row.upstream_endpoint?.trim() || '-' }}</span>
            </div>
          </div>
        </template>

        <template #cell-group="{ row }">
          <span
            v-if="row.group"
            class="theme-chip theme-chip--regular theme-chip--brand-purple inline-flex"
          >
            {{ row.group.name }}
          </span>
          <span v-else class="usage-table__text-soft text-sm">-</span>
        </template>

        <template #cell-stream="{ row }">
          <span :class="getRequestTypeBadgeClass(row)">
            {{ getRequestTypeLabel(row) }}
          </span>
        </template>

        <template #cell-tokens="{ row }">
          <div v-if="row.image_count > 0" class="flex items-center gap-1.5">
            <svg
              class="usage-table__tone usage-table__tone--purple h-4 w-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
              />
            </svg>
            <span class="usage-table__text-body font-medium">
              {{ row.image_count }}{{ t('usage.imageUnit') }}
            </span>
            <span class="usage-table__text-soft">({{ row.image_size || '2K' }})</span>
          </div>

          <div v-else class="flex items-center gap-1.5">
            <div class="space-y-1 text-sm">
              <div class="flex items-center gap-2">
                <div class="inline-flex items-center gap-1">
                  <Icon
                    name="arrowDown"
                    size="sm"
                    class="usage-table__tone usage-table__tone--success h-3.5 w-3.5"
                  />
                  <span class="usage-table__text-body font-medium">
                    {{ row.input_tokens?.toLocaleString() || 0 }}
                  </span>
                </div>
                <div class="inline-flex items-center gap-1">
                  <Icon
                    name="arrowUp"
                    size="sm"
                    class="usage-table__tone usage-table__tone--purple h-3.5 w-3.5"
                  />
                  <span class="usage-table__text-body font-medium">
                    {{ row.output_tokens?.toLocaleString() || 0 }}
                  </span>
                </div>
              </div>

              <div
                v-if="row.cache_read_tokens > 0 || row.cache_creation_tokens > 0"
                class="flex items-center gap-2"
              >
                <div v-if="row.cache_read_tokens > 0" class="inline-flex items-center gap-1">
                  <svg
                    class="usage-table__tone usage-table__tone--info h-3.5 w-3.5"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4"
                    />
                  </svg>
                  <span class="usage-table__tone usage-table__tone--info font-medium">
                    {{ formatCacheTokens(row.cache_read_tokens) }}
                  </span>
                </div>

                <div v-if="row.cache_creation_tokens > 0" class="inline-flex items-center gap-1">
                  <svg
                    class="usage-table__tone usage-table__tone--warning h-3.5 w-3.5"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                    />
                  </svg>
                  <span class="usage-table__tone usage-table__tone--warning font-medium">
                    {{ formatCacheTokens(row.cache_creation_tokens) }}
                  </span>
                  <span
                    v-if="row.cache_creation_1h_tokens > 0"
                    class="theme-chip theme-chip--compact theme-chip--brand-orange"
                  >
                    1h
                  </span>
                  <span
                    v-if="row.cache_ttl_overridden"
                    :title="t('usage.cacheTtlOverriddenHint')"
                    class="theme-chip theme-chip--compact theme-chip--brand-rose cursor-help"
                  >
                    R
                  </span>
                </div>
              </div>
            </div>

            <div
              class="group relative"
              @mouseenter="showTokenTooltip($event, row)"
              @mouseleave="hideTokenTooltip"
            >
              <div class="usage-table__info-trigger flex h-4 w-4 cursor-help items-center justify-center rounded-full transition-colors">
                <Icon
                  name="infoCircle"
                  size="xs"
                  class="usage-table__info-icon transition-colors"
                />
              </div>
            </div>
          </div>
        </template>

        <template #cell-cost="{ row }">
          <div class="text-sm">
            <div class="flex items-center gap-1.5">
              <span class="usage-table__tone usage-table__tone--success font-medium">
                ${{ row.actual_cost?.toFixed(6) || '0.000000' }}
              </span>
              <div
                class="group relative"
                @mouseenter="showTooltip($event, row)"
                @mouseleave="hideTooltip"
              >
                <div class="usage-table__info-trigger flex h-4 w-4 cursor-help items-center justify-center rounded-full transition-colors">
                  <Icon
                    name="infoCircle"
                    size="xs"
                    class="usage-table__info-icon transition-colors"
                  />
                </div>
              </div>
            </div>
            <div v-if="row.account_rate_multiplier != null" class="usage-table__text-soft mt-0.5 text-[11px]">
              A ${{ (row.total_cost * row.account_rate_multiplier).toFixed(6) }}
            </div>
          </div>
        </template>

        <template #cell-first_token="{ row }">
          <span v-if="row.first_token_ms != null" class="usage-table__text-muted text-sm">
            {{ formatDuration(row.first_token_ms) }}
          </span>
          <span v-else class="usage-table__text-soft text-sm">-</span>
        </template>

        <template #cell-duration="{ row }">
          <span class="usage-table__text-muted text-sm">{{ formatDuration(row.duration_ms) }}</span>
        </template>

        <template #cell-created_at="{ value }">
          <span class="usage-table__text-muted text-sm">{{ formatDateTime(value) }}</span>
        </template>

        <template #cell-user_agent="{ row }">
          <span
            v-if="row.user_agent"
            class="usage-table__text-muted usage-table__user-agent block truncate text-sm"
            :title="row.user_agent"
          >
            {{ formatUserAgent(row.user_agent) }}
          </span>
          <span v-else class="usage-table__text-soft text-sm">-</span>
        </template>

        <template #cell-ip_address="{ row }">
          <span v-if="row.ip_address" class="usage-table__text-muted text-sm font-mono">
            {{ row.ip_address }}
          </span>
          <span v-else class="usage-table__text-soft text-sm">-</span>
        </template>

        <template #empty><EmptyState :message="t('usage.noRecords')" /></template>
      </DataTable>
    </div>
  </div>

  <Teleport to="body">
    <div
      v-if="tokenTooltipVisible"
      class="fixed z-[9999] pointer-events-none -translate-y-1/2"
      :style="{
        left: `${tokenTooltipPosition.x}px`,
        top: `${tokenTooltipPosition.y}px`
      }"
    >
      <div class="usage-table__tooltip usage-table__tooltip-surface whitespace-nowrap text-xs shadow-xl">
        <div class="space-y-1.5">
          <div>
            <div class="usage-table__tooltip-title mb-1 text-xs font-semibold">
              {{ t('usage.tokenDetails') }}
            </div>
            <div
              v-if="tokenTooltipData && tokenTooltipData.input_tokens > 0"
              class="flex items-center justify-between gap-4"
            >
              <span class="usage-table__tooltip-label">{{ t('admin.usage.inputTokens') }}</span>
              <span class="usage-table__tooltip-value font-medium">
                {{ tokenTooltipData.input_tokens.toLocaleString() }}
              </span>
            </div>
            <div
              v-if="tokenTooltipData && tokenTooltipData.output_tokens > 0"
              class="flex items-center justify-between gap-4"
            >
              <span class="usage-table__tooltip-label">{{ t('admin.usage.outputTokens') }}</span>
              <span class="usage-table__tooltip-value font-medium">
                {{ tokenTooltipData.output_tokens.toLocaleString() }}
              </span>
            </div>
            <div v-if="tokenTooltipData && tokenTooltipData.cache_creation_tokens > 0">
              <template
                v-if="
                  tokenTooltipData.cache_creation_5m_tokens > 0 ||
                  tokenTooltipData.cache_creation_1h_tokens > 0
                "
              >
                <div
                  v-if="tokenTooltipData.cache_creation_5m_tokens > 0"
                  class="flex items-center justify-between gap-4"
                >
                  <span class="usage-table__tooltip-label flex items-center gap-1.5">
                    {{ t('admin.usage.cacheCreation5mTokens') }}
                    <span class="usage-table__tooltip-chip usage-table__tooltip-chip--warning">
                      5m
                    </span>
                  </span>
                  <span class="usage-table__tooltip-value font-medium">
                    {{ tokenTooltipData.cache_creation_5m_tokens.toLocaleString() }}
                  </span>
                </div>
                <div
                  v-if="tokenTooltipData.cache_creation_1h_tokens > 0"
                  class="flex items-center justify-between gap-4"
                >
                  <span class="usage-table__tooltip-label flex items-center gap-1.5">
                    {{ t('admin.usage.cacheCreation1hTokens') }}
                    <span class="usage-table__tooltip-chip usage-table__tooltip-chip--orange">
                      1h
                    </span>
                  </span>
                  <span class="usage-table__tooltip-value font-medium">
                    {{ tokenTooltipData.cache_creation_1h_tokens.toLocaleString() }}
                  </span>
                </div>
              </template>
              <div v-else class="flex items-center justify-between gap-4">
                <span class="usage-table__tooltip-label">{{ t('admin.usage.cacheCreationTokens') }}</span>
                <span class="usage-table__tooltip-value font-medium">
                  {{ tokenTooltipData.cache_creation_tokens.toLocaleString() }}
                </span>
              </div>
            </div>
            <div
              v-if="tokenTooltipData && tokenTooltipData.cache_ttl_overridden"
              class="flex items-center justify-between gap-4"
            >
              <span class="usage-table__tooltip-label flex items-center gap-1.5">
                {{ t('usage.cacheTtlOverriddenLabel') }}
                <span class="usage-table__tooltip-chip usage-table__tooltip-chip--rose">
                  R-{{ tokenTooltipData.cache_creation_1h_tokens > 0 ? '5m' : '1H' }}
                </span>
              </span>
              <span class="usage-table__tooltip-value usage-table__tooltip-value--rose font-medium">
                {{
                  tokenTooltipData.cache_creation_1h_tokens > 0
                    ? t('usage.cacheTtlOverridden1h')
                    : t('usage.cacheTtlOverridden5m')
                }}
              </span>
            </div>
            <div
              v-if="tokenTooltipData && tokenTooltipData.cache_read_tokens > 0"
              class="flex items-center justify-between gap-4"
            >
              <span class="usage-table__tooltip-label">{{ t('admin.usage.cacheReadTokens') }}</span>
              <span class="usage-table__tooltip-value font-medium">
                {{ tokenTooltipData.cache_read_tokens.toLocaleString() }}
              </span>
            </div>
          </div>
          <div class="usage-table__tooltip-divider flex items-center justify-between gap-6 pt-1.5">
            <span class="usage-table__tooltip-label">{{ t('usage.totalTokens') }}</span>
            <span class="usage-table__tooltip-value usage-table__tooltip-value--info font-semibold">
              {{
                (
                  (tokenTooltipData?.input_tokens || 0) +
                  (tokenTooltipData?.output_tokens || 0) +
                  (tokenTooltipData?.cache_creation_tokens || 0) +
                  (tokenTooltipData?.cache_read_tokens || 0)
                ).toLocaleString()
              }}
            </span>
          </div>
        </div>
        <div
          :class="[
            'usage-table__tooltip-arrow absolute top-1/2 h-0 w-0 -translate-y-1/2 border-b-[6px] border-t-[6px] border-b-transparent border-t-transparent',
            tokenTooltipPlacement === 'right'
              ? 'usage-table__tooltip-arrow--right right-full border-r-[6px]'
              : 'usage-table__tooltip-arrow--left left-full border-l-[6px]'
          ]"
        />
      </div>
    </div>
  </Teleport>

  <Teleport to="body">
    <div
      v-if="tooltipVisible"
      class="fixed z-[9999] pointer-events-none -translate-y-1/2"
      :style="{
        left: `${tooltipPosition.x}px`,
        top: `${tooltipPosition.y}px`
      }"
    >
      <div class="usage-table__tooltip usage-table__tooltip-surface whitespace-nowrap text-xs shadow-xl">
        <div class="space-y-1.5">
          <div class="usage-table__tooltip-divider mb-2 pb-1.5">
            <div class="usage-table__tooltip-title mb-1 text-xs font-semibold">
              {{ t('usage.costDetails') }}
            </div>
            <div
              v-if="tooltipData && tooltipData.input_cost > 0"
              class="flex items-center justify-between gap-4"
            >
              <span class="usage-table__tooltip-label">{{ t('admin.usage.inputCost') }}</span>
              <span class="usage-table__tooltip-value font-medium">
                ${{ tooltipData.input_cost.toFixed(6) }}
              </span>
            </div>
            <div
              v-if="tooltipData && tooltipData.output_cost > 0"
              class="flex items-center justify-between gap-4"
            >
              <span class="usage-table__tooltip-label">{{ t('admin.usage.outputCost') }}</span>
              <span class="usage-table__tooltip-value font-medium">
                ${{ tooltipData.output_cost.toFixed(6) }}
              </span>
            </div>
            <div
              v-if="tooltipData && tooltipData.input_tokens > 0"
              class="flex items-center justify-between gap-4"
            >
              <span class="usage-table__tooltip-label">{{ t('usage.inputTokenPrice') }}</span>
              <span class="usage-table__tooltip-value usage-table__tooltip-value--info font-medium">
                {{ formatTokenPricePerMillion(tooltipData.input_cost, tooltipData.input_tokens) }}
                {{ t('usage.perMillionTokens') }}
              </span>
            </div>
            <div
              v-if="tooltipData && tooltipData.output_tokens > 0"
              class="flex items-center justify-between gap-4"
            >
              <span class="usage-table__tooltip-label">{{ t('usage.outputTokenPrice') }}</span>
              <span class="usage-table__tooltip-value usage-table__tooltip-value--purple font-medium">
                {{ formatTokenPricePerMillion(tooltipData.output_cost, tooltipData.output_tokens) }}
                {{ t('usage.perMillionTokens') }}
              </span>
            </div>
            <div
              v-if="tooltipData && tooltipData.cache_creation_cost > 0"
              class="flex items-center justify-between gap-4"
            >
              <span class="usage-table__tooltip-label">{{ t('admin.usage.cacheCreationCost') }}</span>
              <span class="usage-table__tooltip-value font-medium">
                ${{ tooltipData.cache_creation_cost.toFixed(6) }}
              </span>
            </div>
            <div
              v-if="tooltipData && tooltipData.cache_read_cost > 0"
              class="flex items-center justify-between gap-4"
            >
              <span class="usage-table__tooltip-label">{{ t('admin.usage.cacheReadCost') }}</span>
              <span class="usage-table__tooltip-value font-medium">
                ${{ tooltipData.cache_read_cost.toFixed(6) }}
              </span>
            </div>
          </div>

          <div class="flex items-center justify-between gap-6">
            <span class="usage-table__tooltip-label">{{ t('usage.serviceTier') }}</span>
            <span class="usage-table__tooltip-value usage-table__tooltip-value--info font-semibold">
              {{ getUsageServiceTierLabel(tooltipData?.service_tier, t) }}
            </span>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="usage-table__tooltip-label">{{ t('usage.rate') }}</span>
            <span class="usage-table__tooltip-value usage-table__tooltip-value--info font-semibold">
              {{ (tooltipData?.rate_multiplier || 1).toFixed(2) }}x
            </span>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="usage-table__tooltip-label">{{ t('usage.accountMultiplier') }}</span>
            <span class="usage-table__tooltip-value usage-table__tooltip-value--info font-semibold">
              {{ (tooltipData?.account_rate_multiplier ?? 1).toFixed(2) }}x
            </span>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="usage-table__tooltip-label">{{ t('usage.original') }}</span>
            <span class="usage-table__tooltip-value font-medium">
              ${{ tooltipData?.total_cost?.toFixed(6) || '0.000000' }}
            </span>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="usage-table__tooltip-label">{{ t('usage.userBilled') }}</span>
            <span class="usage-table__tooltip-value usage-table__tooltip-value--success font-semibold">
              ${{ tooltipData?.actual_cost?.toFixed(6) || '0.000000' }}
            </span>
          </div>
          <div class="usage-table__tooltip-divider flex items-center justify-between gap-6 pt-1.5">
            <span class="usage-table__tooltip-label">{{ t('usage.accountBilled') }}</span>
            <span class="usage-table__tooltip-value usage-table__tooltip-value--success font-semibold">
              ${{ (((tooltipData?.total_cost || 0) * (tooltipData?.account_rate_multiplier ?? 1)) || 0).toFixed(6) }}
            </span>
          </div>
        </div>
        <div
          :class="[
            'usage-table__tooltip-arrow absolute top-1/2 h-0 w-0 -translate-y-1/2 border-b-[6px] border-t-[6px] border-b-transparent border-t-transparent',
            tooltipPlacement === 'right'
              ? 'usage-table__tooltip-arrow--right right-full border-r-[6px]'
              : 'usage-table__tooltip-arrow--left left-full border-l-[6px]'
          ]"
        />
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { formatDateTime, formatReasoningEffort } from '@/utils/format'
import { formatTokenPricePerMillion } from '@/utils/usagePricing'
import { getUsageServiceTierLabel } from '@/utils/usageServiceTier'
import DataTable from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'
import type { AdminUsageLog } from '@/types'
import {
  getUsageRequestTypeBadgeClass as resolveUsageRequestTypeBadgeClass,
  getUsageRequestTypeLabel as resolveUsageRequestTypeLabel
} from '@/utils/usagePresentation'

type TooltipPlacement = 'left' | 'right'

defineProps(['data', 'loading', 'columns'])
defineEmits(['userClick'])

const { t } = useI18n()

const tooltipVisible = ref(false)
const tooltipPosition = ref({ x: 0, y: 0 })
const tooltipPlacement = ref<TooltipPlacement>('right')
const tooltipData = ref<AdminUsageLog | null>(null)

const tokenTooltipVisible = ref(false)
const tokenTooltipPosition = ref({ x: 0, y: 0 })
const tokenTooltipPlacement = ref<TooltipPlacement>('right')
const tokenTooltipData = ref<AdminUsageLog | null>(null)

const getRequestTypeLabel = (row: AdminUsageLog): string => resolveUsageRequestTypeLabel(row, t)
const getRequestTypeBadgeClass = (row: AdminUsageLog): string => resolveUsageRequestTypeBadgeClass(row)

const formatCacheTokens = (tokens: number): string => {
  if (tokens >= 1_000_000) return `${(tokens / 1_000_000).toFixed(1)}M`
  if (tokens >= 1_000) return `${(tokens / 1_000).toFixed(1)}K`
  return tokens.toString()
}

const formatUserAgent = (ua: string): string => ua

const formatDuration = (ms: number | null | undefined): string => {
  if (ms == null) return '-'
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

const resolveTooltipGeometry = (
  target: HTMLElement,
  preferredWidth: number
): { x: number; y: number; placement: TooltipPlacement } => {
  const rect = target.getBoundingClientRect()
  const horizontalGap = 8
  const viewportPadding = 12
  const canPlaceRight = rect.right + preferredWidth + horizontalGap + viewportPadding <= window.innerWidth
  const placement: TooltipPlacement = canPlaceRight ? 'right' : 'left'

  const x =
    placement === 'right'
      ? rect.right + horizontalGap
      : Math.max(viewportPadding, rect.left - preferredWidth - horizontalGap)

  const y = Math.min(
    Math.max(rect.top + rect.height / 2, 56),
    Math.max(56, window.innerHeight - 56)
  )

  return { x, y, placement }
}

const showTooltip = (event: MouseEvent, row: AdminUsageLog) => {
  const target = event.currentTarget as HTMLElement
  const geometry = resolveTooltipGeometry(target, 340)

  tooltipData.value = row
  tooltipPosition.value = { x: geometry.x, y: geometry.y }
  tooltipPlacement.value = geometry.placement
  tooltipVisible.value = true
}

const hideTooltip = () => {
  tooltipVisible.value = false
  tooltipData.value = null
}

const showTokenTooltip = (event: MouseEvent, row: AdminUsageLog) => {
  const target = event.currentTarget as HTMLElement
  const geometry = resolveTooltipGeometry(target, 360)

  tokenTooltipData.value = row
  tokenTooltipPosition.value = { x: geometry.x, y: geometry.y }
  tokenTooltipPlacement.value = geometry.placement
  tokenTooltipVisible.value = true
}

const hideTokenTooltip = () => {
  tokenTooltipVisible.value = false
  tokenTooltipData.value = null
}
</script>

<style scoped>
.usage-table__text-body {
  color: var(--theme-page-text);
}

.usage-table__text-muted {
  color: var(--theme-page-muted);
}

.usage-table__text-soft {
  color: color-mix(in srgb, var(--theme-page-muted) 76%, transparent);
}

.usage-table__text-support {
  color: color-mix(in srgb, var(--theme-page-text) 72%, var(--theme-page-muted));
}

.usage-table__endpoint-cell {
  max-width: var(--theme-usage-table-endpoint-max-width);
}

.usage-table__user-agent {
  max-width: var(--theme-usage-table-user-agent-max-width);
}

.usage-table__user-link {
  color: color-mix(in srgb, var(--theme-accent) 84%, var(--theme-page-text));
}

.usage-table__user-link:hover {
  color: color-mix(in srgb, var(--theme-accent-strong) 20%, var(--theme-accent) 80%);
}

.usage-table__tone {
  --usage-table-tone-rgb: var(--theme-info-rgb);
  color: color-mix(in srgb, rgb(var(--usage-table-tone-rgb)) 84%, var(--theme-page-text));
}

.usage-table__tone--success {
  --usage-table-tone-rgb: var(--theme-success-rgb);
}

.usage-table__tone--info {
  --usage-table-tone-rgb: var(--theme-info-rgb);
}

.usage-table__tone--warning {
  --usage-table-tone-rgb: var(--theme-warning-rgb);
}

.usage-table__tone--purple {
  --usage-table-tone-rgb: var(--theme-brand-purple-rgb);
}

.usage-table__tone--orange {
  --usage-table-tone-rgb: var(--theme-brand-orange-rgb);
}

.usage-table__tone--rose {
  --usage-table-tone-rgb: var(--theme-brand-rose-rgb);
}

.usage-table__info-trigger {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.usage-table__info-trigger:hover {
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 10%, var(--theme-surface));
}

.usage-table__info-icon {
  color: color-mix(in srgb, var(--theme-page-muted) 78%, transparent);
}

.usage-table__info-trigger:hover .usage-table__info-icon {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.usage-table__tooltip {
  border: 1px solid color-mix(in srgb, var(--theme-surface-contrast) 16%, transparent);
  background: color-mix(in srgb, var(--theme-surface-contrast) 94%, var(--theme-surface));
  color: var(--theme-surface-contrast-text);
  box-shadow: 0 20px 40px color-mix(in srgb, var(--theme-overlay-strong) 46%, transparent);
}

.usage-table__tooltip-surface {
  border-radius: var(--theme-usage-table-tooltip-radius);
  padding:
    var(--theme-usage-table-tooltip-padding-y)
    var(--theme-usage-table-tooltip-padding-x);
}

.usage-table__tooltip-title,
.usage-table__tooltip-value {
  color: var(--theme-surface-contrast-text);
}

.usage-table__tooltip-label {
  color: color-mix(in srgb, var(--theme-surface-contrast-text) 62%, transparent);
}

.usage-table__tooltip-divider {
  border-top: 1px solid color-mix(in srgb, var(--theme-surface-contrast-text) 16%, transparent);
}

.usage-table__tooltip-divider:first-child {
  border-top: none;
  border-bottom: 1px solid color-mix(in srgb, var(--theme-surface-contrast-text) 16%, transparent);
}

.usage-table__tooltip-value--info {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 74%, var(--theme-surface-contrast-text));
}

.usage-table__tooltip-value--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 74%, var(--theme-surface-contrast-text));
}

.usage-table__tooltip-value--purple {
  color: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 74%, var(--theme-surface-contrast-text));
}

.usage-table__tooltip-value--rose {
  color: color-mix(in srgb, rgb(var(--theme-brand-rose-rgb)) 74%, var(--theme-surface-contrast-text));
}

.usage-table__tooltip-chip {
  --usage-table-tooltip-chip-rgb: var(--theme-info-rgb);
  display: inline-flex;
  align-items: center;
  border: 1px solid color-mix(in srgb, rgb(var(--usage-table-tooltip-chip-rgb)) 20%, transparent);
  border-radius: calc(var(--theme-button-radius) - 4px);
  padding: 0 0.25rem;
  font-size: 10px;
  font-weight: 600;
  line-height: 1.2;
  background: color-mix(in srgb, rgb(var(--usage-table-tooltip-chip-rgb)) 18%, transparent);
  color: color-mix(in srgb, rgb(var(--usage-table-tooltip-chip-rgb)) 74%, var(--theme-surface-contrast-text));
}

.usage-table__tooltip-chip--warning {
  --usage-table-tooltip-chip-rgb: var(--theme-warning-rgb);
}

.usage-table__tooltip-chip--orange {
  --usage-table-tooltip-chip-rgb: var(--theme-brand-orange-rgb);
}

.usage-table__tooltip-chip--rose {
  --usage-table-tooltip-chip-rgb: var(--theme-brand-rose-rgb);
}

.usage-table__tooltip-arrow--right {
  border-right-color: color-mix(in srgb, var(--theme-surface-contrast) 94%, var(--theme-surface));
}

.usage-table__tooltip-arrow--left {
  border-left-color: color-mix(in srgb, var(--theme-surface-contrast) 94%, var(--theme-surface));
}
</style>
