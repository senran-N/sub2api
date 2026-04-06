<template>
  <div class="flex items-center gap-2">
    <div
      class="theme-avatar-badge flex h-8 w-8 items-center justify-center rounded-full"
    >
      <span class="text-sm font-medium">
        {{ userInitial }}
      </span>
    </div>
    <span class="theme-text-strong font-medium">
      {{ userLabel }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { UserSubscription } from '@/types'

const props = defineProps<{
  subscription: UserSubscription
  mode: 'email' | 'username'
}>()

const { t } = useI18n()

const userLabel = computed(() => {
  if (props.mode === 'email') {
    return props.subscription.user?.email || t('admin.redeem.userPrefix', { id: props.subscription.user_id })
  }

  return props.subscription.user?.username || '-'
})

const userInitial = computed(() => userLabel.value.charAt(0).toUpperCase() || '?')
</script>
