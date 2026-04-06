<template>
  <div
    v-if="items.length > 0"
    class="key-usage-model-stats-table fade-up fade-up-delay-4"
  >
    <div class="key-usage-model-stats-table__header">
      <h3 class="key-usage-model-stats-table__title">
        {{ title }}
      </h3>
    </div>
    <div class="overflow-x-auto">
      <table class="w-full">
        <thead>
          <tr class="key-usage-model-stats-table__head-row">
            <th class="key-usage-model-stats-table__head-cell key-usage-model-stats-table__head-cell--left">{{ labels.model }}</th>
            <th class="key-usage-model-stats-table__head-cell">{{ labels.requests }}</th>
            <th class="key-usage-model-stats-table__head-cell">{{ labels.inputTokens }}</th>
            <th class="key-usage-model-stats-table__head-cell">{{ labels.outputTokens }}</th>
            <th class="key-usage-model-stats-table__head-cell">{{ labels.cacheCreationTokens }}</th>
            <th class="key-usage-model-stats-table__head-cell">{{ labels.cacheReadTokens }}</th>
            <th class="key-usage-model-stats-table__head-cell">{{ labels.totalTokens }}</th>
            <th class="key-usage-model-stats-table__head-cell">{{ labels.cost }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(item, index) in items"
            :key="index"
            class="key-usage-model-stats-table__row"
          >
            <td class="key-usage-model-stats-table__cell key-usage-model-stats-table__cell--primary">
              {{ item.model || '-' }}
            </td>
            <td class="key-usage-model-stats-table__cell">{{ fmtNum(item.requests) }}</td>
            <td class="key-usage-model-stats-table__cell">{{ fmtNum(item.input_tokens) }}</td>
            <td class="key-usage-model-stats-table__cell">{{ fmtNum(item.output_tokens) }}</td>
            <td class="key-usage-model-stats-table__cell">{{ fmtNum(item.cache_creation_tokens) }}</td>
            <td class="key-usage-model-stats-table__cell">{{ fmtNum(item.cache_read_tokens) }}</td>
            <td class="key-usage-model-stats-table__cell">{{ fmtNum(item.total_tokens) }}</td>
            <td class="key-usage-model-stats-table__cell key-usage-model-stats-table__cell--emphasis">
              {{ usd(item.actual_cost != null ? item.actual_cost : item.cost) }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { KeyUsageModelStat } from './keyUsageView'

defineProps<{
  items: KeyUsageModelStat[]
  title: string
  labels: {
    model: string
    requests: string
    inputTokens: string
    outputTokens: string
    cacheCreationTokens: string
    cacheReadTokens: string
    totalTokens: string
    cost: string
  }
  fmtNum: (value: number | null | undefined) => string
  usd: (value: number | null | undefined) => string
}>()
</script>

<style scoped>
.key-usage-model-stats-table {
  overflow: hidden;
  border: 1px solid var(--theme-card-border);
  border-radius: calc(var(--theme-surface-radius) + 4px);
  background: color-mix(in srgb, var(--theme-surface) 90%, transparent);
  backdrop-filter: blur(14px);
}

.key-usage-model-stats-table__header {
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 84%, transparent);
  padding: 1.25rem 2rem;
}

.key-usage-model-stats-table__title {
  color: var(--theme-page-muted);
  font-size: 0.875rem;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.key-usage-model-stats-table__head-row {
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 84%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.key-usage-model-stats-table__head-cell {
  padding: 0.75rem 1rem;
  color: var(--theme-page-muted);
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-align: right;
  text-transform: uppercase;
}

.key-usage-model-stats-table__head-cell--left {
  text-align: left;
}

.key-usage-model-stats-table__row {
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.key-usage-model-stats-table__row:last-child {
  border-bottom: 0;
}

.key-usage-model-stats-table__cell {
  padding: 0.75rem 1rem;
  color: var(--theme-page-text);
  font-size: 0.875rem;
  text-align: right;
  font-variant-numeric: tabular-nums;
}

.key-usage-model-stats-table__cell--primary,
.key-usage-model-stats-table__cell--emphasis {
  font-weight: 600;
}

.key-usage-model-stats-table__cell--primary {
  text-align: left;
  white-space: nowrap;
}
</style>
