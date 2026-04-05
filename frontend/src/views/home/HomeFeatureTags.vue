<template>
  <div class="mb-12 flex flex-wrap items-center justify-center gap-4 md:gap-6">
    <div
      v-for="tag in tags"
      :key="tag.key"
      class="inline-flex items-center gap-2.5 rounded-full border border-gray-200/50 bg-white/80 px-5 py-2.5 shadow-sm backdrop-blur-sm dark:border-dark-700/50 dark:bg-dark-800/80"
    >
      <Icon :name="tag.icon" size="sm" class="text-primary-500" />
      <span class="text-sm font-medium text-gray-700 dark:text-dark-200">{{ tag.label }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'
import type { HomeFeatureTag } from './homeView'

defineProps<{
  tags: HomeFeatureTag[]
}>()
</script>
