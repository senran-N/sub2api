<template>
  <div
    class="user-column-settings-menu absolute right-0 top-full z-50 overflow-y-auto"
  >
    <button
      v-for="column in toggleableColumns"
      :key="column.key"
      class="user-column-settings-menu__button flex w-full items-center justify-between text-left text-sm"
      @click="emit('toggle-column', column.key)"
    >
      <span>{{ column.label }}</span>
      <Icon
        v-if="isColumnVisible(column.key)"
        name="check"
        size="sm"
        class="user-column-settings-menu__check"
        :stroke-width="2"
      />
    </button>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'
import type { Column } from '@/components/common/types'

defineProps<{
  toggleableColumns: Column[]
  isColumnVisible: (key: string) => boolean
}>()

const emit = defineEmits<{
  'toggle-column': [key: string]
}>()
</script>

<style scoped>
.user-column-settings-menu {
  width: var(--theme-settings-menu-width-sm);
  max-height: var(--theme-settings-menu-max-height);
  margin-top: var(--theme-floating-panel-gap);
  padding-block: calc(var(--theme-floating-panel-gap) * 0.5 + 0.125rem);
  border: 1px solid color-mix(in srgb, var(--theme-dropdown-border) 88%, transparent);
  border-radius: calc(var(--theme-surface-radius) + 2px);
  background: var(--theme-dropdown-bg);
  box-shadow: var(--theme-dropdown-shadow);
}

.user-column-settings-menu__button {
  padding: calc(var(--theme-button-padding-y) * 0.8) var(--theme-button-padding-x);
  color: var(--theme-page-text);
  transition: background-color 0.2s ease, color 0.2s ease;
}

.user-column-settings-menu__button:hover {
  background: var(--theme-dropdown-item-hover-bg);
}

.user-column-settings-menu__check {
  color: var(--theme-accent);
}
</style>
