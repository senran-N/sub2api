<template>
  <div class="flex flex-wrap items-center gap-3">
    <div class="flex-1 sm:max-w-64">
      <input
        :value="searchQuery"
        type="text"
        :placeholder="t('admin.promo.searchCodes')"
        class="input"
        @input="handleSearchInput"
      />
    </div>

    <Select
      :model-value="status"
      :options="statusOptions"
      class="w-36"
      @update:model-value="handleStatusChange"
      @change="emit('refresh')"
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
        {{ t('admin.promo.createCode') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  searchQuery: string
  status: '' | 'active' | 'disabled'
  statusOptions: Array<{ value: string; label: string }>
  loading: boolean
}>()

const emit = defineEmits<{
  'update:searchQuery': [value: string]
  'update:status': [value: '' | 'active' | 'disabled']
  search: []
  refresh: []
  create: []
}>()

const { t } = useI18n()

const handleSearchInput = (event: Event) => {
  emit('update:searchQuery', (event.target as HTMLInputElement).value)
  emit('search')
}

const handleStatusChange = (value: string | number | boolean | null) => {
  if (value === '' || value === 'active' || value === 'disabled') {
    emit('update:status', value)
  }
}
</script>
