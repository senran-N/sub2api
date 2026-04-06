<template>
  <div v-if="expiresAt">
    <span
      class="text-sm"
      :class="
        isSubscriptionExpiringSoon(expiresAt)
          ? 'theme-text-warning'
          : 'theme-text-default'
      "
    >
      {{ formatDateOnly(expiresAt) }}
    </span>
    <div v-if="daysRemaining !== null" class="theme-text-muted text-xs">
      {{ daysRemaining }} {{ t('admin.subscriptions.daysRemaining') }}
    </div>
  </div>
  <span v-else class="theme-text-muted text-sm">
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
