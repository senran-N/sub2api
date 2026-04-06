<template>
  <div v-if="proxy" class="flex items-center gap-2">
    <span class="theme-text-default text-sm">{{ proxy.name }}</span>
    <span v-if="proxy.country_code" class="theme-text-muted text-xs">
      ({{ proxy.country_code }})
    </span>
  </div>
  <span v-else class="theme-text-subtle text-sm">-</span>
</template>

<script setup lang="ts">
import type { Account } from '@/types'

defineProps<{
  proxy?: Account['proxy'] | null
}>()
</script>
