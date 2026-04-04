<template>
  <div class="flex flex-wrap items-center gap-3">
    <div class="flex-1 sm:max-w-64">
      <input
        :value="searchQuery"
        type="text"
        :placeholder="t('admin.announcements.searchAnnouncements')"
        class="input"
        @input="handleSearchInput"
      />
    </div>
    <Select
      :model-value="status"
      :options="statusOptions"
      class="w-40"
      @update:model-value="handleStatusUpdate"
      @change="emit('status-change')"
    />

    <div class="flex flex-1 flex-wrap items-center justify-end gap-2">
      <button
        class="btn btn-secondary"
        :disabled="loading"
        :title="t('common.refresh')"
        @click="emit('refresh')"
      >
        <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
      </button>
      <button class="btn btn-primary" @click="emit('create')">
        <Icon name="plus" size="md" class="mr-1" />
        {{ t('admin.announcements.createAnnouncement') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import type { AnnouncementStatus } from '@/types'

defineProps<{
  searchQuery: string
  status: '' | AnnouncementStatus
  statusOptions: Array<{ value: string; label: string }>
  loading: boolean
}>()

const emit = defineEmits<{
  'update:searchQuery': [value: string]
  'update:status': [value: '' | AnnouncementStatus]
  search: []
  'status-change': []
  refresh: []
  create: []
}>()

const { t } = useI18n()

const handleSearchInput = (event: Event) => {
  emit('update:searchQuery', (event.target as HTMLInputElement).value)
  emit('search')
}

const handleStatusUpdate = (value: string | number | boolean | null) => {
  if (value === '' || value === 'draft' || value === 'active' || value === 'archived') {
    emit('update:status', value)
  }
}
</script>
