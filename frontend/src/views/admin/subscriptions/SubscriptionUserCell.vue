<template>
  <div class="flex items-center gap-2">
    <div
      class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-100 dark:bg-primary-900/30"
    >
      <span class="text-sm font-medium text-primary-700 dark:text-primary-300">
        {{ userInitial }}
      </span>
    </div>
    <span class="font-medium text-gray-900 dark:text-white">
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
