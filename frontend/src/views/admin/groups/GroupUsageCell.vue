<template>
  <div v-if="loading" class="group-usage-cell__placeholder text-xs">—</div>
  <div v-else class="group-usage-cell space-y-0.5 text-xs">
    <div class="group-usage-cell__row">
      <span class="group-usage-cell__label">{{ t('admin.groups.usageToday') }}</span>
      <span class="group-usage-cell__value">${{ formatGroupCost(summary?.today_cost ?? 0) }}</span>
    </div>
    <div class="group-usage-cell__row">
      <span class="group-usage-cell__label">{{ t('admin.groups.usageTotal') }}</span>
      <span class="group-usage-cell__value">${{ formatGroupCost(summary?.total_cost ?? 0) }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { formatGroupCost } from '../groupsTable'

defineProps<{
  loading: boolean
  summary?: {
    today_cost: number
    total_cost: number
  }
}>()

const { t } = useI18n()
</script>

<style scoped>
.group-usage-cell__placeholder,
.group-usage-cell__label {
  color: color-mix(in srgb, var(--theme-page-muted) 82%, transparent);
}

.group-usage-cell__row {
  color: var(--theme-page-muted);
}

.group-usage-cell__value {
  margin-left: 0.25rem;
  font-weight: 600;
  color: color-mix(in srgb, var(--theme-page-text) 84%, transparent);
}
</style>
