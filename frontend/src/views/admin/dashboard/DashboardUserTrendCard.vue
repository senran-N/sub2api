<template>
  <div class="dashboard-user-trend-card card">
    <h3 class="theme-text-strong mb-4 text-sm font-semibold">
      {{ t('admin.dashboard.recentUsage') }} (Top 12)
    </h3>
    <div class="h-64">
      <div v-if="loading" class="flex h-full items-center justify-center">
        <LoadingSpinner size="md" />
      </div>
      <Line v-else-if="chartData" :data="chartData" :options="chartOptions" />
      <div
        v-else
        class="theme-text-muted flex h-full items-center justify-center text-sm"
      >
        {{ t('admin.dashboard.noDataAvailable') }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { Line } from 'vue-chartjs'

defineProps<{
  loading: boolean
  chartData: {
    labels: string[]
    datasets: Array<{
      label: string
      data: number[]
      borderColor: string
      backgroundColor: string
      fill: boolean
      tension: number
    }>
  } | null
  chartOptions: Record<string, unknown>
}>()

const { t } = useI18n()
</script>
<style scoped>
.dashboard-user-trend-card {
  padding: var(--theme-settings-card-panel-padding);
}
</style>
