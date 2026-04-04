<template>
  <span class="text-sm text-gray-600 dark:text-gray-300">
    {{ usedCount }} / {{ maxUses === 0 ? '∞' : maxUses }}
  </span>
</template>

<script setup lang="ts">
defineProps<{
  usedCount: number
  maxUses: number
}>()
</script>
