<template>
  <div class="usage-charts-toolbar card">
    <div class="flex flex-wrap items-center gap-4">
      <div class="flex items-center gap-2">
        <span class="toolbar-label">{{ t('admin.dashboard.timeRange') }}:</span>
        <DateRangePicker
          :start-date="startDate"
          :end-date="endDate"
          @update:start-date="emit('update:startDate', $event)"
          @update:end-date="emit('update:endDate', $event)"
          @change="emit('date-range-change', $event)"
        />
      </div>
      <div class="ml-auto flex items-center gap-2">
        <span class="toolbar-label">{{ t('admin.dashboard.granularity') }}:</span>
        <div class="w-28">
          <Select
            :model-value="granularity"
            :options="granularityOptions"
            @update:model-value="handleGranularityUpdate"
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
  granularityOptions: Array<{ value: 'day' | 'hour'; label: string }>
}>()

const emit = defineEmits<{
  'update:startDate': [value: string]
  'update:endDate': [value: string]
  'update:granularity': [value: 'day' | 'hour']
  'date-range-change': [value: { startDate: string; endDate: string; preset: string | null }]
  'granularity-change': []
}>()

const { t } = useI18n()

const handleGranularityUpdate = (value: string | number | boolean | null) => {
  if (value === 'day' || value === 'hour') {
    emit('update:granularity', value)
  }
}
</script>

<style scoped>
.usage-charts-toolbar {
  padding: var(--theme-user-dashboard-charts-card-padding);
}
</style>
