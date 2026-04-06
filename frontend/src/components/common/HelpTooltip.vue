<script setup lang="ts">
import { ref, useTemplateRef, nextTick } from 'vue'

defineProps<{
  content?: string
}>()

const show = ref(false)
const triggerRef = useTemplateRef<HTMLElement>('trigger')
const tooltipStyle = ref({ top: '0px', left: '0px' })

function onEnter() {
  show.value = true
  nextTick(updatePosition)
}

function onLeave() {
  show.value = false
}

function updatePosition() {
  const el = triggerRef.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  tooltipStyle.value = {
    top: `${rect.top + window.scrollY}px`,
    left: `${rect.left + rect.width / 2 + window.scrollX}px`,
  }
}
</script>

<template>
  <div
    ref="trigger"
    class="group relative ml-1 inline-flex items-center align-middle"
    @mouseenter="onEnter"
    @mouseleave="onLeave"
  >
    <!-- Trigger Icon -->
    <slot name="trigger">
      <svg
        class="help-tooltip__icon h-4 w-4 cursor-help transition-colors"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        stroke-width="2"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
        />
      </svg>
    </slot>

    <!-- Teleport to body to escape modal overflow clipping -->
    <Teleport to="body">
      <div
        v-show="show"
        class="help-tooltip__content fixed z-[99999] text-xs leading-relaxed"
        :style="{ top: tooltipStyle.top, left: tooltipStyle.left }"
      >
        <slot>{{ content }}</slot>
        <div class="help-tooltip__arrow absolute left-1/2"></div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.help-tooltip__icon {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.help-tooltip__icon:hover {
  color: var(--theme-accent);
}

.help-tooltip__content,
.help-tooltip__arrow {
  background: var(--theme-surface-emphasis);
  color: var(--theme-page-bg);
}

.help-tooltip__content {
  width: var(--theme-tooltip-width);
  padding: var(--theme-tooltip-padding);
  border-radius: var(--theme-tooltip-radius);
  box-shadow: var(--theme-dropdown-shadow);
  border: 1px solid color-mix(in srgb, var(--theme-surface-emphasis) 88%, transparent);
  transform: translate(-50%, calc(-100% - var(--theme-tooltip-arrow-size)));
}

.help-tooltip__arrow {
  bottom: calc(var(--theme-tooltip-arrow-size) * -0.5);
  width: var(--theme-tooltip-arrow-size);
  height: var(--theme-tooltip-arrow-size);
  transform: translateX(-50%) rotate(45deg);
}
</style>
