<template>
  <div
    class="absolute right-0 top-full z-50 mt-1 w-48 rounded-lg border border-gray-200 bg-white py-1 shadow-lg dark:border-dark-600 dark:bg-dark-800"
  >
    <button
      v-for="filter in builtInFilters"
      :key="filter.key"
      class="flex w-full items-center justify-between px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
      @click="emit('toggle-built-in-filter', filter.key)"
    >
      <span>{{ filter.name }}</span>
      <Icon
        v-if="visibleFilters.has(filter.key)"
        name="check"
        size="sm"
        class="text-primary-500"
        :stroke-width="2"
      />
    </button>
    <div
      v-if="filterableAttributes.length > 0"
      class="my-1 border-t border-gray-100 dark:border-dark-700"
    ></div>
    <button
      v-for="attribute in filterableAttributes"
      :key="attribute.id"
      class="flex w-full items-center justify-between px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
      @click="emit('toggle-attribute-filter', attribute)"
    >
      <span>{{ attribute.name }}</span>
      <Icon
        v-if="visibleFilters.has(`attr_${attribute.id}`)"
        name="check"
        size="sm"
        class="text-primary-500"
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
