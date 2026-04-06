<template>
  <div class="user-filter-settings-menu absolute right-0 top-full z-50">
    <button
      v-for="filter in builtInFilters"
      :key="filter.key"
      class="user-filter-settings-menu__item flex w-full items-center justify-between text-left text-sm"
      @click="emit('toggle-built-in-filter', filter.key)"
    >
      <span>{{ filter.name }}</span>
      <Icon
        v-if="visibleFilters.has(filter.key)"
        name="check"
        size="sm"
        class="user-filter-settings-menu__check"
        :stroke-width="2"
      />
    </button>
    <div
      v-if="filterableAttributes.length > 0"
      class="user-filter-settings-menu__divider border-t"
    ></div>
    <button
      v-for="attribute in filterableAttributes"
      :key="attribute.id"
      class="user-filter-settings-menu__item flex w-full items-center justify-between text-left text-sm"
      @click="emit('toggle-attribute-filter', attribute)"
    >
      <span>{{ attribute.name }}</span>
      <Icon
        v-if="visibleFilters.has(`attr_${attribute.id}`)"
        name="check"
        size="sm"
        class="user-filter-settings-menu__check"
        :stroke-width="2"
      />
    </button>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'
import type { UserAttributeDefinition } from '@/types'
import type { BuiltInUserFilterKey } from '../usersTable'

defineProps<{
  visibleFilters: Set<string>
  builtInFilters: Array<{ key: BuiltInUserFilterKey; name: string }>
  filterableAttributes: UserAttributeDefinition[]
}>()

const emit = defineEmits<{
  'toggle-built-in-filter': [key: BuiltInUserFilterKey]
  'toggle-attribute-filter': [attribute: UserAttributeDefinition]
}>()
</script>

<style scoped>
.user-filter-settings-menu {
  width: var(--theme-settings-menu-width-sm);
  max-height: var(--theme-settings-menu-max-height);
  overflow-y: auto;
  margin-top: var(--theme-floating-panel-gap);
  padding-block: calc(var(--theme-floating-panel-gap) * 0.5 + 0.125rem);
  border: 1px solid color-mix(in srgb, var(--theme-dropdown-border) 88%, transparent);
  border-radius: calc(var(--theme-surface-radius) + 2px);
  background: var(--theme-dropdown-bg);
  box-shadow: var(--theme-dropdown-shadow);
}

.user-filter-settings-menu__item {
  padding: calc(var(--theme-button-padding-y) * 0.8) var(--theme-button-padding-x);
  color: var(--theme-dropdown-text);
  transition: background-color 0.2s ease;
}

.user-filter-settings-menu__item:hover {
  background: var(--theme-dropdown-item-hover-bg);
}

.user-filter-settings-menu__divider {
  margin: calc(var(--theme-floating-panel-gap) * 0.5 + 0.125rem) 0;
  border-color: color-mix(in srgb, var(--theme-dropdown-border) 76%, transparent);
}

.user-filter-settings-menu__check {
  color: var(--theme-accent);
}
</style>
