<template>
  <div class="mb-12 flex flex-wrap items-center justify-center gap-4 md:gap-6">
    <div
      v-for="tag in tags"
      :key="tag.key"
      class="home-feature-tags__tag inline-flex items-center gap-2.5 rounded-full shadow-sm backdrop-blur-sm"
    >
      <Icon :name="tag.icon" size="sm" class="home-feature-tags__icon" />
      <span class="home-feature-tags__label text-sm font-medium">{{ tag.label }}</span>
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

<style scoped>
.home-feature-tags__tag {
  padding:
    calc(var(--theme-markdown-block-padding) - 0.375rem)
    calc(var(--theme-markdown-block-padding) + 0.25rem);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 48%, transparent);
  background: color-mix(in srgb, var(--theme-surface) 82%, transparent);
}

.home-feature-tags__icon {
  color: var(--theme-accent);
}

.home-feature-tags__label {
  color: color-mix(in srgb, var(--theme-page-text) 82%, transparent);
}
</style>
