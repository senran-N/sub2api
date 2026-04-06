<template>
  <div class="text-sm">
    <div class="flex items-center gap-1.5">
      <span class="theme-text-muted">{{ t('keys.today') }}:</span>
      <span class="font-medium theme-text-default">${{ summary.todayCost }}</span>
    </div>
    <div class="mt-0.5 flex items-center gap-1.5">
      <span class="theme-text-muted">{{ t('keys.total') }}:</span>
      <span class="font-medium theme-text-default">${{ summary.totalCost }}</span>
    </div>

    <div v-if="row.quota > 0" class="mt-1.5">
      <div class="flex items-center gap-1.5">
        <span class="theme-text-muted">{{ t('keys.quota') }}:</span>
        <span :class="['font-medium tabular-nums', getApiKeyQuotaTextTone(row)]">
          ${{ row.quota_used?.toFixed(2) || '0.00' }} / ${{ row.quota?.toFixed(2) }}
        </span>
      </div>
      <div class="theme-progress-track mt-1 h-1.5 w-full">
        <div
          :class="['theme-progress-fill', getApiKeyQuotaBarTone(row)]"
          :style="{ width: getApiKeyQuotaProgressWidth(row) }"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ApiKey } from '@/types'
import type { BatchApiKeyUsageStats } from '@/api/usage'
import {
  getApiKeyQuotaBarTone,
  getApiKeyQuotaProgressWidth,
  getApiKeyQuotaTextTone,
  getApiKeyUsageSummary
} from './keysView'

const props = defineProps<{
  row: ApiKey
  stats?: BatchApiKeyUsageStats
}>()

const { t } = useI18n()

const summary = computed(() => getApiKeyUsageSummary(props.stats))
</script>
