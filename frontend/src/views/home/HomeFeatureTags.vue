<template>
  <div class="home-tags mb-12 flex flex-wrap items-center justify-center gap-4 md:gap-6">
    <div
      v-for="tag in tags"
      :key="tag.key"
      class="home-tags__tag"
    >
      <Icon :name="tag.icon" size="sm" class="home-tags__icon" />
      <span class="home-tags__label">{{ tag.label }}</span>
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
.home-tags__tag {
  display: inline-flex;
  align-items: center;
  gap: 0.625rem;
  padding: var(--theme-home-tag-padding);
  border-radius: var(--theme-home-tag-radius);
  border: var(--theme-home-tag-border);
  background: var(--theme-home-tag-bg);
  box-shadow: var(--theme-home-tag-shadow);
  backdrop-filter: var(--theme-home-tag-backdrop-filter);
  transition: transform 0.2s ease, border-color 0.2s ease, background 0.2s ease;
}

.home-tags__icon {
  color: var(--theme-accent);
}

.home-tags__label {
  font-family: var(--theme-home-tag-label-font);
  font-size: var(--theme-home-tag-label-size);
  font-weight: var(--theme-home-tag-label-weight);
  letter-spacing: var(--theme-home-tag-label-letter-spacing);
  text-transform: var(--theme-home-tag-label-transform);
  color: var(--theme-home-tag-label-color);
}

.home-tags__tag:hover {
  transform: var(--theme-home-tag-hover-transform);
  border-color: var(--theme-home-tag-hover-border-color);
  background: var(--theme-home-tag-hover-bg);
  box-shadow: var(--theme-home-tag-hover-shadow);
}

.home-tags__label::before {
  content: var(--theme-home-tag-label-prefix);
  color: var(--theme-home-tag-label-prefix-color);
}

.home-tags__label::after {
  content: var(--theme-home-tag-label-suffix);
  color: var(--theme-home-tag-label-suffix-color);
}
</style>
