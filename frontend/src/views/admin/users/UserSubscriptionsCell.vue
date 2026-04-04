<template>
  <div
    v-if="user.subscriptions && user.subscriptions.length > 0"
    class="flex flex-wrap gap-1.5"
  >
    <GroupBadge
      v-for="subscription in user.subscriptions"
      :key="subscription.id"
      :name="subscription.group?.name || ''"
      :platform="subscription.group?.platform"
      :subscription-type="subscription.group?.subscription_type"
      :rate-multiplier="subscription.group?.rate_multiplier"
      :days-remaining="subscription.expires_at ? getUserSubscriptionDaysRemaining(subscription.expires_at) : null"
      :title="subscription.expires_at ? formatDateTime(subscription.expires_at) : ''"
    />
  </div>
  <span
    v-else
    class="inline-flex items-center gap-1.5 rounded-md bg-gray-50 px-2 py-1 text-xs text-gray-400 dark:bg-dark-700/50 dark:text-dark-500"
  >
    <Icon name="ban" size="xs" class="h-3.5 w-3.5" />
    <span>{{ t('admin.users.noSubscription') }}</span>
  </span>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { formatDateTime } from '@/utils/format'
import GroupBadge from '@/components/common/GroupBadge.vue'
import Icon from '@/components/icons/Icon.vue'
import type { AdminUser } from '@/types'
import { getUserSubscriptionDaysRemaining } from '../usersTable'

defineProps<{
  user: AdminUser
}>()

const { t } = useI18n()
</script>
