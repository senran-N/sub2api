<template>
  <div class="flex flex-wrap items-center gap-3">
    <div class="flex-1 sm:max-w-64">
      <input
        :value="searchQuery"
        type="text"
        :placeholder="t('admin.redeem.searchCodes')"
        class="input"
        @input="handleSearchInput"
      />
    </div>
    <Select
      :model-value="type"
      :options="typeOptions"
      class="w-36"
      @update:model-value="handleTypeChange"
      @change="emit('refresh')"
    />
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
      <button class="btn btn-secondary" @click="emit('export')">
        {{ t('admin.redeem.exportCsv') }}
      </button>
      <button class="btn btn-primary" @click="emit('generate')">
        {{ t('admin.redeem.generateCodes') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import type { RedeemCodeType } from '@/types'
import type { RedeemStatusFilter } from '../redeemForm'

defineProps<{
  searchQuery: string
  type: '' | RedeemCodeType
  status: RedeemStatusFilter
  typeOptions: Array<{ value: string; label: string }>
  statusOptions: Array<{ value: string; label: string }>
  loading: boolean
}>()

const emit = defineEmits<{
  'update:searchQuery': [value: string]
  'update:type': [value: '' | RedeemCodeType]
  'update:status': [value: RedeemStatusFilter]
  search: []
  refresh: []
  export: []
  generate: []
}>()

const { t } = useI18n()

const handleSearchInput = (event: Event) => {
  emit('update:searchQuery', (event.target as HTMLInputElement).value)
  emit('search')
}

const handleTypeChange = (value: string | number | boolean | null) => {
  if (
    value === '' ||
    value === 'balance' ||
    value === 'concurrency' ||
    value === 'subscription' ||
    value === 'invitation'
  ) {
    emit('update:type', value)
  }
}

const handleStatusChange = (value: string | number | boolean | null) => {
  if (value === '' || value === 'unused' || value === 'used' || value === 'expired') {
    emit('update:status', value)
  }
}
</script>
