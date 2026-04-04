<template>
  <div class="space-y-1">
    <span
      :class="[
        'inline-block rounded-full px-2 py-0.5 text-xs font-medium',
        group.subscription_type === 'subscription'
          ? 'bg-violet-100 text-violet-700 dark:bg-violet-900/30 dark:text-violet-400'
          : 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-300'
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
      class="text-xs text-gray-500 dark:text-gray-400"
    >
      <template
        v-if="group.daily_limit_usd || group.weekly_limit_usd || group.monthly_limit_usd"
      >
        <span v-if="group.daily_limit_usd"
          >${{ group.daily_limit_usd }}/{{ t('admin.groups.limitDay') }}</span
        >
        <span
          v-if="group.daily_limit_usd && (group.weekly_limit_usd || group.monthly_limit_usd)"
          class="mx-1 text-gray-300 dark:text-gray-600"
          >·</span
        >
        <span v-if="group.weekly_limit_usd"
          >${{ group.weekly_limit_usd }}/{{ t('admin.groups.limitWeek') }}</span
        >
        <span
          v-if="group.weekly_limit_usd && group.monthly_limit_usd"
          class="mx-1 text-gray-300 dark:text-gray-600"
          >·</span
        >
        <span v-if="group.monthly_limit_usd"
          >${{ group.monthly_limit_usd }}/{{ t('admin.groups.limitMonth') }}</span
        >
      </template>
      <span v-else class="text-gray-400 dark:text-gray-500">
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
