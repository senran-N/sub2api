<template>
  <div class="space-y-1">
    <span
      :class="[
        'badge',
        group.subscription_type === 'subscription'
          ? 'badge-purple'
          : 'badge-gray'
      ]"
    >
      {{
        group.subscription_type === 'subscription'
          ? t('admin.groups.subscription.subscription')
          : t('admin.groups.subscription.standard')
      }}
    </span>
    <div
      v-if="group.subscription_type === 'subscription'"
      class="group-billing-type-cell__details text-xs"
    >
      <template
        v-if="group.daily_limit_usd || group.weekly_limit_usd || group.monthly_limit_usd"
      >
        <span v-if="group.daily_limit_usd"
          >${{ group.daily_limit_usd }}/{{ t('admin.groups.limitDay') }}</span
        >
        <span
          v-if="group.daily_limit_usd && (group.weekly_limit_usd || group.monthly_limit_usd)"
          class="group-billing-type-cell__separator mx-1"
          >·</span
        >
        <span v-if="group.weekly_limit_usd"
          >${{ group.weekly_limit_usd }}/{{ t('admin.groups.limitWeek') }}</span
        >
        <span
          v-if="group.weekly_limit_usd && group.monthly_limit_usd"
          class="group-billing-type-cell__separator mx-1"
          >·</span
        >
        <span v-if="group.monthly_limit_usd"
          >${{ group.monthly_limit_usd }}/{{ t('admin.groups.limitMonth') }}</span
        >
      </template>
      <span v-else class="group-billing-type-cell__empty">
        {{ t('admin.groups.subscription.noLimit') }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { AdminGroup } from '@/types'

defineProps<{
  group: AdminGroup
}>()

const { t } = useI18n()
</script>

<style scoped>
.group-billing-type-cell__details {
  color: var(--theme-page-muted);
}

.group-billing-type-cell__separator,
.group-billing-type-cell__empty {
  color: color-mix(in srgb, var(--theme-page-muted) 82%, transparent);
}
</style>
