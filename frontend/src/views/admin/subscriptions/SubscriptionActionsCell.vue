<template>
  <div class="flex items-center gap-1">
    <button
      v-if="subscription.status === 'active' || subscription.status === 'expired'"
      class="theme-action-button theme-action-button--info"
      @click="emit('adjust', subscription)"
    >
      <Icon name="calendar" size="sm" />
      <span class="text-xs">{{ t('admin.subscriptions.adjust') }}</span>
    </button>
    <button
      v-if="subscription.status === 'active'"
      class="theme-action-button theme-action-button--brand-orange"
      :disabled="resetting"
      @click="emit('reset-quota', subscription)"
    >
      <Icon name="refresh" size="sm" />
      <span class="text-xs">{{ t('admin.subscriptions.resetQuota') }}</span>
    </button>
    <button
      v-if="subscription.status === 'active'"
      class="theme-action-button theme-action-button--danger"
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
