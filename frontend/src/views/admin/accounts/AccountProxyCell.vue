<template>
  <div v-if="proxy" class="flex items-center gap-2">
    <span class="text-sm text-gray-700 dark:text-gray-300">{{ proxy.name }}</span>
    <span v-if="proxy.country_code" class="text-xs text-gray-500 dark:text-gray-400">
      ({{ proxy.country_code }})
    </span>
  </div>
  <span v-else class="text-sm text-gray-400 dark:text-dark-500">-</span>
</template>

<script setup lang="ts">
import type { Account } from '@/types'

defineProps<{
  proxy?: Account['proxy'] | null
}>()
</script>
