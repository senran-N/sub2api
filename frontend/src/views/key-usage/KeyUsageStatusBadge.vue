<template>
  <div v-if="status" class="mb-2 flex items-center justify-center fade-up">
    <div
      class="inline-flex items-center gap-2 rounded-full border border-gray-200 bg-white/90 px-5 py-2.5 shadow-sm backdrop-blur-sm dark:border-dark-700 dark:bg-dark-900/90"
    >
      <span
        class="pulse-dot h-2.5 w-2.5 rounded-full"
        :class="status.isActive ? 'bg-emerald-500' : 'bg-rose-500'"
      ></span>
      <span class="text-sm font-medium text-gray-900 dark:text-white">{{ status.label }}</span>
      <span class="text-xs text-gray-400 dark:text-dark-500">|</span>
      <span class="text-xs text-gray-500 dark:text-dark-400">{{ status.statusText }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { KeyUsageStatusInfo } from './keyUsageView'

defineProps<{
  status: KeyUsageStatusInfo | null
}>()
</script>
