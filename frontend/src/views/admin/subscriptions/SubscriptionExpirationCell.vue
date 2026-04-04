<template>
  <div v-if="expiresAt">
    <span
      class="text-sm"
      :class="
        isSubscriptionExpiringSoon(expiresAt)
          ? 'text-orange-600 dark:text-orange-400'
          : 'text-gray-700 dark:text-gray-300'
      "
    >
      {{ formatDateOnly(expiresAt) }}
    </span>
    <div v-if="daysRemaining !== null" class="text-xs text-gray-500">
      {{ daysRemaining }} {{ t('admin.subscriptions.daysRemaining') }}
    </div>
  </div>
  <span v-else class="text-sm text-gray-500">
    {{ t('admin.subscriptions.noExpiration') }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { formatDateOnly } from '@/utils/format'
import {
  getSubscriptionDaysRemaining,
  isSubscriptionExpiringSoon
} from '../subscriptionForm'

const props = defineProps<{
  expiresAt: string | null
}>()

const { t } = useI18n()

const daysRemaining = computed(() => getSubscriptionDaysRemaining(props.expiresAt))
</script>
