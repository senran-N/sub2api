<template>
  <div class="flex flex-wrap items-center justify-end gap-2">
    <div class="flex items-center gap-2 md:contents">
      <button
        class="btn btn-secondary px-2 md:px-3"
        :disabled="loading"
        :title="t('common.refresh')"
        @click="emit('refresh')"
      >
        <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
      </button>

      <div class="relative" ref="filterDropdownRef">
        <button
          class="btn btn-secondary px-2 md:px-3"
          :title="t('admin.users.filterSettings')"
          @click="toggleFilterDropdown"
        >
          <Icon name="filter" size="sm" class="md:mr-1.5" />
          <span class="hidden md:inline">{{ t('admin.users.filterSettings') }}</span>
        </button>
        <UserFilterSettingsMenu
          v-if="showFilterDropdown"
          :visible-filters="visibleFilters"
          :built-in-filters="builtInFilters"
          :filterable-attributes="filterableAttributes"
          @toggle-built-in-filter="emit('toggle-built-in-filter', $event)"
          @toggle-attribute-filter="emit('toggle-attribute-filter', $event)"
        />
      </div>

      <div class="relative" ref="columnDropdownRef">
        <button
          class="btn btn-secondary px-2 md:px-3"
          :title="t('admin.users.columnSettings')"
          @click="toggleColumnDropdown"
        >
          <Icon name="grid" size="sm" class="md:mr-1.5" />
          <span class="hidden md:inline">{{ t('admin.users.columnSettings') }}</span>
        </button>
        <UserColumnSettingsMenu
          v-if="showColumnDropdown"
          :toggleable-columns="toggleableColumns"
          :is-column-visible="isColumnVisible"
          @toggle-column="emit('toggle-column', $event)"
        />
      </div>

      <button
        class="btn btn-secondary px-2 md:px-3"
        :title="t('admin.users.attributes.configButton')"
        @click="emit('open-attributes')"
      >
        <Icon name="cog" size="sm" class="md:mr-1.5" />
        <span class="hidden md:inline">{{ t('admin.users.attributes.configButton') }}</span>
      </button>
    </div>

    <button class="btn btn-primary flex-1 md:flex-initial" @click="emit('create')">
      <Icon name="plus" size="md" class="mr-2" />
      {{ t('admin.users.createUser') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Column } from '@/components/common/types'
import Icon from '@/components/icons/Icon.vue'
import type { UserAttributeDefinition } from '@/types'
import type { BuiltInUserFilterKey } from '../usersTable'
import UserColumnSettingsMenu from './UserColumnSettingsMenu.vue'
import UserFilterSettingsMenu from './UserFilterSettingsMenu.vue'

defineProps<{
  loading: boolean
  visibleFilters: Set<string>
  builtInFilters: Array<{ key: BuiltInUserFilterKey; name: string }>
  filterableAttributes: UserAttributeDefinition[]
  toggleableColumns: Column[]
  isColumnVisible: (key: string) => boolean
}>()

const emit = defineEmits<{
  refresh: []
  'toggle-built-in-filter': [key: BuiltInUserFilterKey]
  'toggle-attribute-filter': [attribute: UserAttributeDefinition]
  'toggle-column': [key: string]
  'open-attributes': []
  create: []
}>()

const { t } = useI18n()

const showFilterDropdown = ref(false)
const showColumnDropdown = ref(false)
const filterDropdownRef = ref<HTMLElement | null>(null)
const columnDropdownRef = ref<HTMLElement | null>(null)

const toggleFilterDropdown = () => {
  showFilterDropdown.value = !showFilterDropdown.value
  showColumnDropdown.value = false
}

const toggleColumnDropdown = () => {
  showColumnDropdown.value = !showColumnDropdown.value
  showFilterDropdown.value = false
}

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  if (filterDropdownRef.value && !filterDropdownRef.value.contains(target)) {
    showFilterDropdown.value = false
  }
  if (columnDropdownRef.value && !columnDropdownRef.value.contains(target)) {
    showColumnDropdown.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>
