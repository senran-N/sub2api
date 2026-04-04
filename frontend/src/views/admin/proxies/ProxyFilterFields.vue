<template>
  <div class="flex flex-wrap items-center gap-3">
    <div class="relative w-full sm:w-64">
      <Icon
        name="search"
        size="md"
        class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500"
      />
      <input
        :value="searchQuery"
        type="text"
        :placeholder="t('admin.proxies.searchProxies')"
        class="input pl-10"
        @input="handleSearchInput"
      />
    </div>

    <div class="w-full sm:w-40">
      <Select
        :model-value="protocol"
        :options="protocolOptions"
        :placeholder="t('admin.proxies.allProtocols')"
        @update:modelValue="handleProtocolUpdate"
        @change="emit('protocol-change')"
      />
    </div>

    <div class="w-full sm:w-36">
      <Select
        :model-value="status"
        :options="statusOptions"
        :placeholder="t('admin.proxies.allStatus')"
        @update:modelValue="handleStatusUpdate"
        @change="emit('status-change')"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  searchQuery: string
  protocol: string
  status: string
  protocolOptions: SelectOption[]
  statusOptions: SelectOption[]
}>()

const emit = defineEmits<{
  'update:searchQuery': [value: string]
  'update:protocol': [value: string]
  'update:status': [value: string]
  'search-input': []
  'protocol-change': []
  'status-change': []
}>()

const { t } = useI18n()

const normalizeSelectValue = (value: string | number | boolean | null) =>
  value === null ? '' : String(value)

const handleSearchInput = (event: Event) => {
  const target = event.target as HTMLInputElement
  emit('update:searchQuery', target.value)
  emit('search-input')
}

const handleProtocolUpdate = (value: string | number | boolean | null) => {
  emit('update:protocol', normalizeSelectValue(value))
}

const handleStatusUpdate = (value: string | number | boolean | null) => {
  emit('update:status', normalizeSelectValue(value))
}
</script>
