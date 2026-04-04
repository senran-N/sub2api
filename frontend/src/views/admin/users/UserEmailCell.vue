<template>
  <div class="flex items-center gap-2">
    <div
      class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-100 dark:bg-primary-900/30"
    >
      <span class="text-sm font-medium text-primary-700 dark:text-primary-300">
        {{ email.charAt(0).toUpperCase() }}
      </span>
    </div>
    <span class="font-medium text-gray-900 dark:text-white">{{ email }}</span>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  email: string
}>()
</script>
