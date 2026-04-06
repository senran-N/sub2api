<template>
  <div class="max-w-xs">
    <span
      class="theme-text-default block truncate text-sm"
      :title="value"
    >
      {{ value }}
    </span>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  value: string
}>()
</script>
