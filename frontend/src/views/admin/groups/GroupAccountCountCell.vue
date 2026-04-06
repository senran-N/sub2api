<template>
  <div class="group-account-count-cell space-y-0.5 text-xs">
    <div>
      <span class="group-account-count-cell__label">{{ t('admin.groups.accountsAvailable') }}</span>
      <span class="group-account-count-cell__value group-account-count-cell__value--success ml-1 font-medium">{{ availableCount }}</span>
      <span class="group-account-count-cell__unit ml-1 inline-flex items-center font-medium">{{ t('admin.groups.accountsUnit') }}</span>
    </div>
    <div v-if="group.rate_limited_account_count">
      <span class="group-account-count-cell__label">{{ t('admin.groups.accountsRateLimited') }}</span>
      <span class="group-account-count-cell__value group-account-count-cell__value--warning ml-1 font-medium">{{ group.rate_limited_account_count }}</span>
      <span class="group-account-count-cell__unit ml-1 inline-flex items-center font-medium">{{ t('admin.groups.accountsUnit') }}</span>
    </div>
    <div>
      <span class="group-account-count-cell__label">{{ t('admin.groups.accountsTotal') }}</span>
      <span class="group-account-count-cell__value ml-1 font-medium">{{ group.account_count || 0 }}</span>
      <span class="group-account-count-cell__unit ml-1 inline-flex items-center font-medium">{{ t('admin.groups.accountsUnit') }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { AdminGroup } from '@/types'

const props = defineProps<{
  group: AdminGroup
}>()

const { t } = useI18n()

const availableCount = computed(() => {
  return (props.group.active_account_count || 0) - (props.group.rate_limited_account_count || 0)
})
</script>

<style scoped>
.group-account-count-cell__label {
  color: var(--theme-page-muted);
}

.group-account-count-cell__value {
  color: color-mix(in srgb, var(--theme-page-text) 86%, transparent);
}

.group-account-count-cell__value--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.group-account-count-cell__value--warning {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.group-account-count-cell__unit {
  border-radius: var(--theme-group-account-unit-radius);
  padding: var(--theme-group-account-unit-padding-y) var(--theme-group-account-unit-padding-x);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: color-mix(in srgb, var(--theme-page-text) 82%, transparent);
}
</style>
