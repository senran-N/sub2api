<template>
  <div class="card user-usage-filters-bar">
    <div class="user-usage-filters-bar__content">
      <div class="user-usage-filters-bar__row">
        <div class="user-usage-filters-bar__api-key-filter">
          <label class="input-label">{{ t('usage.apiKeyFilter') }}</label>
          <Select
            :model-value="apiKeyId ?? null"
            :options="apiKeyOptions"
            :placeholder="t('usage.allApiKeys')"
            @update:model-value="handleApiKeyUpdate"
            @change="emit('apply-filters')"
          />
        </div>

        <div class="user-usage-filters-bar__date-range-filter">
          <label class="input-label">{{ t('usage.timeRange') }}</label>
          <DateRangePicker
            :start-date="startDate"
            :end-date="endDate"
            @update:start-date="emit('update:startDate', $event)"
            @update:end-date="emit('update:endDate', $event)"
            @change="emit('date-range-change', $event)"
          />
        </div>

        <div class="user-usage-filters-bar__actions">
          <button type="button" :disabled="loading" class="btn btn-secondary" @click="emit('apply-filters')">
            {{ t('common.refresh') }}
          </button>
          <button type="button" class="btn btn-secondary" @click="emit('reset')">
            {{ t('common.reset') }}
          </button>
          <button type="button" :disabled="exporting" class="btn btn-primary" @click="emit('export')">
            <svg
              v-if="exporting"
              class="-ml-1 mr-2 h-4 w-4 animate-spin"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            {{ exporting ? t('usage.exporting') : t('usage.exportCsv') }}
          </button>
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
  apiKeyId?: number
  apiKeyOptions: Array<{ value: number | null; label: string }>
  startDate: string
  endDate: string
  loading: boolean
  exporting: boolean
}>()

const emit = defineEmits<{
  'update:apiKeyId': [value: number | undefined]
  'update:startDate': [value: string]
  'update:endDate': [value: string]
  'date-range-change': [value: { startDate: string; endDate: string; preset: string | null }]
  'apply-filters': []
  reset: []
  export: []
}>()

const { t } = useI18n()

const handleApiKeyUpdate = (value: string | number | boolean | null) => {
  if (value == null || value === '') {
    emit('update:apiKeyId', undefined)
    return
  }

  if (typeof value === 'number') {
    emit('update:apiKeyId', value)
  }
}
</script>

<style scoped>
.user-usage-filters-bar__content {
  padding: calc(var(--theme-table-mobile-card-padding) * 1.5) calc(var(--theme-table-mobile-card-padding) * 1.5) var(--theme-table-mobile-card-padding);
}

.user-usage-filters-bar__row {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: var(--theme-table-mobile-card-padding);
}

.user-usage-filters-bar__api-key-filter {
  min-width: min(100%, var(--theme-balance-history-filter-width));
}

.user-usage-filters-bar__date-range-filter {
  min-width: min(100%, var(--theme-balance-history-filter-width));
}

.user-usage-filters-bar__actions {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: calc(var(--theme-table-mobile-card-padding) * 0.75);
}

@media (max-width: 767px) {
  .user-usage-filters-bar__actions {
    width: 100%;
    margin-left: 0;
    justify-content: flex-start;
    flex-wrap: wrap;
  }
}
</style>
