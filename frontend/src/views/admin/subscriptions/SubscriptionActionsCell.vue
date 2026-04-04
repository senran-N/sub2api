<template>
  <div class="flex items-center gap-1">
    <button
      v-if="subscription.status === 'active' || subscription.status === 'expired'"
      class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-blue-50 hover:text-blue-600 dark:hover:bg-blue-900/20 dark:hover:text-blue-400"
      @click="emit('adjust', subscription)"
    >
      <Icon name="calendar" size="sm" />
      <span class="text-xs">{{ t('admin.subscriptions.adjust') }}</span>
    </button>
    <button
      v-if="subscription.status === 'active'"
      class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-orange-50 hover:text-orange-600 dark:hover:bg-orange-900/20 dark:hover:text-orange-400 disabled:cursor-not-allowed disabled:opacity-50"
      :disabled="resetting"
      @click="emit('reset-quota', subscription)"
    >
      <Icon name="refresh" size="sm" />
      <span class="text-xs">{{ t('admin.subscriptions.resetQuota') }}</span>
    </button>
    <button
      v-if="subscription.status === 'active'"
      class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
      @click="emit('revoke', subscription)"
    >
      <Icon name="ban" size="sm" />
      <span class="text-xs">{{ t('admin.subscriptions.revoke') }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { UserSubscription } from '@/types'

defineProps<{
  subscription: UserSubscription
  resetting: boolean
}>()

const emit = defineEmits<{
  adjust: [subscription: UserSubscription]
  'reset-quota': [subscription: UserSubscription]
  revoke: [subscription: UserSubscription]
}>()

const { t } = useI18n()
</script>
