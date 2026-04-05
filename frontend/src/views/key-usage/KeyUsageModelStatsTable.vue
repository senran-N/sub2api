<template>
  <div
    v-if="items.length > 0"
    class="fade-up fade-up-delay-4 overflow-hidden rounded-2xl border border-gray-200 bg-white/90 backdrop-blur-sm dark:border-dark-700 dark:bg-dark-900/90"
  >
    <div class="border-b border-gray-200 px-8 py-5 dark:border-dark-700">
      <h3 class="text-sm font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">
        {{ title }}
      </h3>
    </div>
    <div class="overflow-x-auto">
      <table class="w-full">
        <thead>
          <tr class="border-b border-gray-200 bg-gray-50 dark:border-dark-700 dark:bg-dark-950">
            <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">{{ labels.model }}</th>
            <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">{{ labels.requests }}</th>
            <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">{{ labels.inputTokens }}</th>
            <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">{{ labels.outputTokens }}</th>
            <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">{{ labels.cacheCreationTokens }}</th>
            <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">{{ labels.cacheReadTokens }}</th>
            <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">{{ labels.totalTokens }}</th>
            <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">{{ labels.cost }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(item, index) in items"
            :key="index"
            class="border-b border-gray-100 last:border-b-0 dark:border-dark-800"
          >
            <td class="whitespace-nowrap px-4 py-3 text-sm font-medium text-gray-900 dark:text-white">
              {{ item.model || '-' }}
            </td>
            <td class="px-4 py-3 text-right text-sm tabular-nums text-gray-700 dark:text-dark-200">{{ fmtNum(item.requests) }}</td>
            <td class="px-4 py-3 text-right text-sm tabular-nums text-gray-700 dark:text-dark-200">{{ fmtNum(item.input_tokens) }}</td>
            <td class="px-4 py-3 text-right text-sm tabular-nums text-gray-700 dark:text-dark-200">{{ fmtNum(item.output_tokens) }}</td>
            <td class="px-4 py-3 text-right text-sm tabular-nums text-gray-700 dark:text-dark-200">{{ fmtNum(item.cache_creation_tokens) }}</td>
            <td class="px-4 py-3 text-right text-sm tabular-nums text-gray-700 dark:text-dark-200">{{ fmtNum(item.cache_read_tokens) }}</td>
            <td class="px-4 py-3 text-right text-sm tabular-nums text-gray-700 dark:text-dark-200">{{ fmtNum(item.total_tokens) }}</td>
            <td class="px-4 py-3 text-right text-sm font-medium tabular-nums text-gray-900 dark:text-white">
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
