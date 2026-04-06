<template>
  <div class="account-column-settings-menu overflow-y-auto">
    <button
      v-for="column in toggleableColumns"
      :key="column.key"
      class="account-column-settings-menu__item flex w-full items-center justify-between text-sm"
      @click="emit('toggle-column', column.key)"
    >
      <span>{{ column.label }}</span>
      <Icon
        v-if="isColumnVisible(column.key)"
        name="check"
        size="sm"
        class="account-column-settings-menu__check"
      />
    </button>
  </div>
</template>

<script setup lang="ts">
import type { Column } from '@/components/common/types'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  toggleableColumns: Column[]
  isColumnVisible: (key: string) => boolean
}>()

const emit = defineEmits<{
  'toggle-column': [key: string]
}>()
</script>

<style scoped>
.account-column-settings-menu {
  max-height: var(--theme-settings-menu-max-height);
  padding: var(--theme-group-selector-padding);
}

.account-column-settings-menu__item {
  padding: var(--theme-dropdown-item-padding-y) var(--theme-dropdown-item-padding-x);
  border-radius: calc(var(--theme-button-radius) + 2px);
  color: var(--theme-page-text);
  transition:
    background-color 0.2s ease,
    color 0.2s ease;
}

.account-column-settings-menu__item:hover {
  background: var(--theme-dropdown-item-hover-bg);
}

.account-column-settings-menu__check {
  color: var(--theme-accent);
}
</style>
