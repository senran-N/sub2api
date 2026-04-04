<template>
  <div class="card p-4">
    <div class="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center sm:gap-4">
      <div class="flex items-center gap-2">
        <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('admin.dashboard.timeRange') }}:
        </span>
        <DateRangePicker
          :start-date="startDate"
          :end-date="endDate"
          @update:start-date="emit('update:startDate', $event)"
          @update:end-date="emit('update:endDate', $event)"
          @change="emit('date-range-change', $event)"
        />
      </div>
      <button class="btn btn-secondary" :disabled="loading" @click="emit('refresh')">
        {{ t('common.refresh') }}
      </button>
      <div class="flex items-center gap-2 sm:ml-auto">
        <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('admin.dashboard.granularity') }}:
        </span>
        <div class="w-28">
          <Select
            :model-value="granularity"
            :options="granularityOptions"
            @update:model-value="handleGranularityChange"
            @change="emit('granularity-change')"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'

defineProps<{
  startDate: string
  endDate: string
  granularity: 'day' | 'hour'
  granularityOptions: Array<{ value: string; label: string }>
  loading: boolean
}>()

const emit = defineEmits<{
  'update:startDate': [value: string]
  'update:endDate': [value: string]
  'update:granularity': [value: 'day' | 'hour']
  'date-range-change': [range: { startDate: string; endDate: string; preset: string | null }]
  refresh: []
  'granularity-change': []
}>()

const { t } = useI18n()

const handleGranularityChange = (value: string | number | boolean | null) => {
  if (value === 'day' || value === 'hour') {
    emit('update:granularity', value)
  }
}
</script>
