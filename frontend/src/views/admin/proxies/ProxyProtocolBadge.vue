<template>
  <span v-if="protocol" :class="badgeClass">
    {{ protocol.toUpperCase() }}
  </span>
  <span v-else class="theme-text-subtle text-sm">-</span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  protocol?: string | null
}>()

const badgeClass = computed(() => [
  'badge',
  props.protocol?.startsWith('socks5') ? 'badge-primary' : 'badge-gray'
])
</script>
