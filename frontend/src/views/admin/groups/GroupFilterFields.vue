<template>
  <div class="flex flex-1 flex-wrap items-center gap-3">
    <div class="relative w-full sm:w-64">
      <Icon
        name="search"
        size="md"
        class="theme-text-subtle absolute left-3 top-1/2 -translate-y-1/2"
      />
      <input
        :value="searchQuery"
        type="text"
        :placeholder="t('admin.groups.searchGroups')"
        class="input pl-10"
        @input="handleSearchInput"
      />
    </div>
    <Select
      :model-value="platform"
      :options="platformOptions"
      :placeholder="t('admin.groups.allPlatforms')"
      class="w-44"
      @update:modelValue="handlePlatformUpdate"
      @change="emit('platform-change')"
    />
    <Select
      :model-value="status"
      :options="statusOptions"
      :placeholder="t('admin.groups.allStatus')"
      class="w-40"
      @update:modelValue="handleStatusUpdate"
      @change="emit('status-change')"
    />
    <Select
      :model-value="isExclusive"
      :options="exclusiveOptions"
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
