<template>
  <div class="flex items-center gap-2">
    <div
      class="theme-avatar-badge flex h-8 w-8 items-center justify-center rounded-full"
    >
      <span class="text-sm font-medium">
        {{ email.charAt(0).toUpperCase() }}
      </span>
    </div>
    <span class="theme-text-strong font-medium">{{ email }}</span>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  email: string
}>()
</script>
