<template>
  <div class="p-2">
    <button
      class="flex w-full items-center justify-between rounded-md px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-200 dark:hover:bg-gray-700"
      @click="emit('toggle-enabled')"
    >
      <span>{{ t('admin.accounts.enableAutoRefresh') }}</span>
      <Icon
        v-if="enabled"
        name="check"
        size="sm"
        class="text-primary-500"
      />
    </button>
    <div class="my-1 border-t border-gray-100 dark:border-gray-700"></div>
    <button
      v-for="seconds in intervals"
      :key="seconds"
      class="flex w-full items-center justify-between rounded-md px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-200 dark:hover:bg-gray-700"
      @click="emit('set-interval', seconds)"
    >
      <span>{{ labelForInterval(seconds) }}</span>
      <Icon
        v-if="selectedIntervalSeconds === seconds"
        name="check"
        size="sm"
        class="text-primary-500"
      />
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { ACCOUNT_AUTO_REFRESH_INTERVALS } from '../accountsList'

type AccountAutoRefreshInterval = (typeof ACCOUNT_AUTO_REFRESH_INTERVALS)[number]

defineProps<{
  enabled: boolean
  intervals: readonly AccountAutoRefreshInterval[]
  selectedIntervalSeconds: AccountAutoRefreshInterval
  labelForInterval: (seconds: AccountAutoRefreshInterval) => string
}>()

const emit = defineEmits<{
  'toggle-enabled': []
  'set-interval': [seconds: AccountAutoRefreshInterval]
}>()

const { t } = useI18n()
</script>
