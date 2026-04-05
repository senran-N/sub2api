<template>
  <div
    v-if="cells.length > 0"
    class="fade-up fade-up-delay-3 overflow-hidden rounded-2xl border border-gray-200 bg-white/90 backdrop-blur-sm dark:border-dark-700 dark:bg-dark-900/90"
  >
    <div class="border-b border-gray-200 px-8 py-5 dark:border-dark-700">
      <h3 class="text-sm font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">
        {{ title }}
      </h3>
    </div>
    <div class="grid grid-cols-2 gap-px bg-gray-100 dark:bg-dark-800 md:grid-cols-4">
      <div
        v-for="(cell, index) in cells"
        :key="index"
        class="bg-white px-6 py-4 dark:bg-dark-900"
      >
        <div class="mb-1 text-xs text-gray-500 dark:text-dark-400">{{ cell.label }}</div>
        <div class="text-sm font-semibold tabular-nums text-gray-900 dark:text-white">
          {{ cell.value }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { KeyUsageStatCell } from './keyUsageView'

defineProps<{
  cells: KeyUsageStatCell[]
  title: string
}>()
</script>
