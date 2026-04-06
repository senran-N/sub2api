<template>
  <span class="theme-text-default text-sm">
    {{ usedCount }} / {{ maxUses === 0 ? '∞' : maxUses }}
  </span>
</template>

<script setup lang="ts">
defineProps<{
  usedCount: number
  maxUses: number
}>()
</script>
