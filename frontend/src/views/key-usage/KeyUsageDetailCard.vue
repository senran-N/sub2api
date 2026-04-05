<template>
  <div
    v-if="rows.length > 0"
    class="fade-up fade-up-delay-3 overflow-hidden rounded-2xl border border-gray-200 bg-white/90 backdrop-blur-sm dark:border-dark-700 dark:bg-dark-900/90"
  >
    <div class="border-b border-gray-200 px-8 py-5 dark:border-dark-700">
      <h3 class="text-sm font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">
        {{ title }}
      </h3>
    </div>
    <div class="divide-y divide-gray-100 dark:divide-dark-800">
      <div
        v-for="(row, index) in rows"
        :key="index"
        class="flex items-center justify-between px-8 py-4"
      >
        <div class="flex items-center gap-3">
          <div class="flex h-8 w-8 items-center justify-center rounded-lg" :class="row.iconBg">
            <svg
              class="h-4 w-4"
              :class="row.iconColor"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              v-html="row.iconSvg"
            ></svg>
          </div>
          <span class="text-sm text-gray-700 dark:text-dark-200">{{ row.label }}</span>
        </div>
        <span
          class="text-sm font-semibold tabular-nums"
          :class="row.valueClass || 'text-gray-900 dark:text-white'"
        >
          {{ row.value }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { KeyUsageDetailRow } from './keyUsageView'

defineProps<{
  rows: KeyUsageDetailRow[]
  title: string
}>()
</script>
