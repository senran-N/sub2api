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
  padding:
    calc(var(--theme-markdown-block-padding) - 0.375rem)
    calc(var(--theme-markdown-block-padding) + 0.25rem);
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 48%, transparent);
  background: color-mix(in srgb, var(--theme-surface) 82%, transparent);
  backdrop-filter: blur(6px);
  transition: transform 0.2s ease, border-color 0.2s ease, background 0.2s ease;
}

.home-tags__icon {
  color: var(--theme-accent);
}

.home-tags__label {
  font-size: 0.875rem;
  font-weight: 500;
  color: color-mix(in srgb, var(--theme-page-text) 82%, transparent);
}

/* ============== Factory: hard spec chip with status prefix ============== */
:root[data-brand-theme='factory'] .home-tags__tag {
  border-radius: 0;
  border: 2px solid var(--theme-page-text);
  background: var(--theme-surface);
  box-shadow: 2px 2px 0 var(--theme-page-text);
  font-family: var(--theme-font-mono);
  padding: 0.4rem 0.85rem;
}

:root[data-brand-theme='factory'] .home-tags__tag:hover {
  transform: translate(-1px, -1px);
  box-shadow: 3px 3px 0 var(--theme-page-text);
}

:root[data-brand-theme='factory'] .home-tags__label {
  font-family: var(--theme-font-mono);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-weight: 700;
  color: var(--theme-page-text);
}

:root[data-brand-theme='factory'] .home-tags__label::before {
  content: '[ ';
  color: var(--theme-accent);
}

:root[data-brand-theme='factory'] .home-tags__label::after {
  content: ' ]';
  color: var(--theme-accent);
}

.dark[data-brand-theme='factory'] .home-tags__tag {
  border-color: rgba(255, 255, 255, 0.3);
  box-shadow: 2px 2px 0 rgba(255, 255, 255, 0.2);
}

.dark[data-brand-theme='factory'] .home-tags__tag:hover {
  box-shadow: 3px 3px 0 rgba(255, 255, 255, 0.3);
}

/* ============== Claude: warm pill with serif italic label ============== */
:root[data-brand-theme='claude'] .home-tags__tag {
  border: 1px solid color-mix(in srgb, var(--theme-accent) 20%, var(--theme-page-border));
  background: color-mix(in srgb, var(--theme-accent-soft) 72%, var(--theme-surface));
}

:root[data-brand-theme='claude'] .home-tags__tag:hover {
  transform: translateY(-2px);
  background: color-mix(in srgb, var(--theme-accent-soft) 90%, var(--theme-surface));
  border-color: color-mix(in srgb, var(--theme-accent) 40%, var(--theme-page-border));
}

:root[data-brand-theme='claude'] .home-tags__label {
  font-family: var(--theme-font-display);
  font-style: italic;
  font-weight: 600;
  font-size: 0.9rem;
  letter-spacing: -0.01em;
  color: color-mix(in srgb, var(--theme-accent) 75%, var(--theme-page-text));
}
</style>
