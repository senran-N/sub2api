<template>
  <div class="group relative inline-flex">
    <Icon
      name="questionCircle"
      size="sm"
      :stroke-width="2"
      class="group-section-info-tooltip__icon cursor-help transition-colors"
    />
    <div class="pointer-events-none absolute bottom-full left-0 z-50 mb-2 hidden group-hover:block group-hover:pointer-events-auto">
      <div :class="['group-section-info-tooltip__panel', widthClass]">
        <p class="group-section-info-tooltip__text text-xs leading-relaxed">
          {{ text }}
        </p>
        <div class="group-section-info-tooltip__arrow absolute -bottom-1.5 left-3 h-3 w-3 rotate-45"></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'

withDefaults(
  defineProps<{
    text: string
    widthClass?: string
  }>(),
  {
    widthClass: 'w-72'
  }
)
</script>

<style scoped>
.group-section-info-tooltip__icon {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.group:hover .group-section-info-tooltip__icon {
  color: var(--theme-accent);
}

.group-section-info-tooltip__panel {
  border: 1px solid color-mix(in srgb, var(--theme-dropdown-border) 88%, transparent);
  background: var(--theme-dropdown-bg);
  box-shadow: var(--theme-dropdown-shadow);
  border-radius: var(--theme-surface-radius);
  padding: var(--theme-group-selector-padding);
}

.group-section-info-tooltip__text {
  color: var(--theme-dropdown-text);
}

.group-section-info-tooltip__arrow {
  background: var(--theme-dropdown-bg);
}
</style>
