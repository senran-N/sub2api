<template>
  <GroupCapacityBadge
    v-if="capacity"
    :concurrency-used="capacity.concurrencyUsed"
    :concurrency-max="capacity.concurrencyMax"
    :sessions-used="capacity.sessionsUsed"
    :sessions-max="capacity.sessionsMax"
    :rpm-used="capacity.rpmUsed"
    :rpm-max="capacity.rpmMax"
  />
  <span v-else class="theme-text-subtle text-xs">—</span>
</template>

<script setup lang="ts">
import GroupCapacityBadge from '@/components/common/GroupCapacityBadge.vue'
import type { GroupCapacitySnapshot } from './groupDisplay'

defineProps<{
  capacity?: GroupCapacitySnapshot
}>()
</script>
