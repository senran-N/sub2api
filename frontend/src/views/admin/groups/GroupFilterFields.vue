<template>
  <div class="flex flex-1 flex-wrap items-center gap-3">
    <div class="relative w-full sm:w-64">
      <Icon
        name="search"
        size="md"
        class="theme-text-subtle absolute left-3 top-1/2 -translate-y-1/2"
      />
      <input
        id="group-filter-search"
        :value="searchQuery"
        name="search_groups"
        type="text"
        autocomplete="off"
        :aria-label="t('admin.groups.searchGroups')"
        :placeholder="t('admin.groups.searchGroups')"
        class="input pl-10"
        @input="handleSearchInput"
      />
    </div>
    <Select
      id="group-filter-platform"
      :model-value="platform"
      name="platform"
      :options="platformOptions"
      :aria-label="t('admin.groups.allPlatforms')"
      :placeholder="t('admin.groups.allPlatforms')"
      class="w-44"
      @update:modelValue="handlePlatformUpdate"
      @change="emit('platform-change')"
    />
    <Select
      id="group-filter-status"
      :model-value="status"
      name="status"
      :options="statusOptions"
      :aria-label="t('admin.groups.allStatus')"
      :placeholder="t('admin.groups.allStatus')"
      class="w-40"
      @update:modelValue="handleStatusUpdate"
      @change="emit('status-change')"
    />
    <Select
      id="group-filter-exclusive"
      :model-value="isExclusive"
      name="is_exclusive"
      :options="exclusiveOptions"
      :aria-label="t('admin.groups.allGroups')"
      :placeholder="t('admin.groups.allGroups')"
      class="w-44"
      @update:modelValue="handleExclusiveUpdate"
      @change="emit('exclusive-change')"
    />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  searchQuery: string
  platform: string
  status: string
  isExclusive: string
  platformOptions: SelectOption[]
  statusOptions: SelectOption[]
  exclusiveOptions: SelectOption[]
}>()

const emit = defineEmits<{
  'update:searchQuery': [value: string]
  'update:platform': [value: string]
  'update:status': [value: string]
  'update:isExclusive': [value: string]
  'search-input': []
  'platform-change': []
  'status-change': []
  'exclusive-change': []
}>()

const { t } = useI18n()

const normalizeSelectValue = (value: string | number | boolean | null) =>
  value === null ? '' : String(value)

const handleSearchInput = (event: Event) => {
  const target = event.target as HTMLInputElement
  emit('update:searchQuery', target.value)
  emit('search-input')
}

const handlePlatformUpdate = (value: string | number | boolean | null) => {
  emit('update:platform', normalizeSelectValue(value))
}

const handleStatusUpdate = (value: string | number | boolean | null) => {
  emit('update:status', normalizeSelectValue(value))
}

const handleExclusiveUpdate = (value: string | number | boolean | null) => {
  emit('update:isExclusive', normalizeSelectValue(value))
}
</script>
